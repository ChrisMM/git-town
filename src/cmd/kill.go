package cmd

import (
	"fmt"

	"github.com/git-town/git-town/v11/src/cli/dialog"
	"github.com/git-town/git-town/v11/src/cli/flags"
	"github.com/git-town/git-town/v11/src/cmd/cmdhelpers"
	"github.com/git-town/git-town/v11/src/config/configdomain"
	"github.com/git-town/git-town/v11/src/execute"
	"github.com/git-town/git-town/v11/src/git/gitdomain"
	"github.com/git-town/git-town/v11/src/gohacks/slice"
	"github.com/git-town/git-town/v11/src/messages"
	"github.com/git-town/git-town/v11/src/sync"
	"github.com/git-town/git-town/v11/src/vm/interpreter"
	"github.com/git-town/git-town/v11/src/vm/opcode"
	"github.com/git-town/git-town/v11/src/vm/program"
	"github.com/git-town/git-town/v11/src/vm/runstate"
	"github.com/spf13/cobra"
)

const killDesc = "Removes an obsolete feature branch"

const killHelp = `
Deletes the current or provided branch from the local and origin repositories. Does not delete perennial branches nor the main branch.`

func killCommand() *cobra.Command {
	addVerboseFlag, readVerboseFlag := flags.Verbose()
	addDryRunFlag, readDryRunFlag := flags.DryRun()
	cmd := cobra.Command{
		Use:   "kill [<branch>]",
		Args:  cobra.MaximumNArgs(1),
		Short: killDesc,
		Long:  cmdhelpers.Long(killDesc, killHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeKill(args, readDryRunFlag(cmd), readVerboseFlag(cmd))
		},
	}
	addDryRunFlag(&cmd)
	addVerboseFlag(&cmd)
	return &cmd
}

func executeKill(args []string, dryRun, verbose bool) error {
	repo, err := execute.OpenRepo(execute.OpenRepoArgs{
		Verbose:          verbose,
		DryRun:           dryRun,
		OmitBranchNames:  false,
		PrintCommands:    true,
		ValidateIsOnline: false,
		ValidateGitRepo:  true,
	})
	if err != nil {
		return err
	}
	config, initialBranchesSnapshot, initialStashSnapshot, exit, err := determineKillConfig(args, repo, dryRun, verbose)
	if err != nil || exit {
		return err
	}
	steps, finalUndoProgram := killProgram(config)
	if err != nil {
		return err
	}
	runState := runstate.RunState{
		Command:             "kill",
		DryRun:              dryRun,
		RunProgram:          steps,
		InitialActiveBranch: initialBranchesSnapshot.Active,
		FinalUndoProgram:    finalUndoProgram,
	}
	return interpreter.Execute(interpreter.ExecuteArgs{
		FullConfig:              config.FullConfig,
		RunState:                &runState,
		Run:                     repo.Runner,
		Connector:               nil,
		DialogTestInputs:        &config.dialogTestInputs,
		Verbose:                 verbose,
		RootDir:                 repo.RootDir,
		InitialBranchesSnapshot: initialBranchesSnapshot,
		InitialConfigSnapshot:   repo.ConfigSnapshot,
		InitialStashSnapshot:    initialStashSnapshot,
	})
}

type killConfig struct {
	*configdomain.FullConfig
	branchToKill     gitdomain.BranchInfo
	branchWhenDone   gitdomain.LocalBranchName
	dialogTestInputs dialog.TestInputs
	dryRun           bool
	hasOpenChanges   bool
	initialBranch    gitdomain.LocalBranchName
	previousBranch   gitdomain.LocalBranchName
}

func determineKillConfig(args []string, repo *execute.OpenRepoResult, dryRun, verbose bool) (*killConfig, gitdomain.BranchesStatus, gitdomain.StashSize, bool, error) {
	branchesSnapshot, stashSnapshot, dialogTestInputs, exit, err := execute.LoadRepoSnapshot(execute.LoadBranchesArgs{
		FullConfig:            &repo.Runner.FullConfig,
		Repo:                  repo,
		Verbose:               verbose,
		Fetch:                 true,
		HandleUnfinishedState: false,
		ValidateIsConfigured:  true,
		ValidateNoOpenChanges: false,
	})
	if err != nil || exit {
		return nil, branchesSnapshot, stashSnapshot, exit, err
	}
	branchNameToKill := gitdomain.NewLocalBranchName(slice.FirstElementOr(args, branchesSnapshot.Active.String()))
	branchToKill := branchesSnapshot.Branches.FindByLocalName(branchNameToKill)
	if branchToKill == nil {
		return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.BranchDoesntExist, branchNameToKill)
	}
	if branchToKill.SyncStatus == gitdomain.SyncStatusOtherWorktree {
		return nil, branchesSnapshot, stashSnapshot, exit, fmt.Errorf(messages.KillBranchOtherWorktree, branchNameToKill)
	}
	if branchToKill.IsLocal() {
		err = execute.EnsureKnownBranchAncestry(branchToKill.LocalName, execute.EnsureKnownBranchAncestryArgs{
			Config:           &repo.Runner.FullConfig,
			AllBranches:      branchesSnapshot.Branches,
			DefaultBranch:    repo.Runner.MainBranch,
			DialogTestInputs: &dialogTestInputs,
			Runner:           repo.Runner,
		})
		if err != nil {
			return nil, branchesSnapshot, stashSnapshot, false, err
		}
	}
	if !repo.Runner.IsFeatureBranch(branchToKill.LocalName) {
		return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.KillOnlyFeatureBranches)
	}
	previousBranch := repo.Runner.Backend.PreviouslyCheckedOutBranch()
	repoStatus, err := repo.Runner.Backend.RepoStatus()
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	var branchWhenDone gitdomain.LocalBranchName
	if branchNameToKill == branchesSnapshot.Active {
		branchWhenDone = previousBranch
	} else {
		branchWhenDone = branchesSnapshot.Active
	}
	return &killConfig{
		FullConfig:       &repo.Runner.FullConfig,
		branchToKill:     *branchToKill,
		branchWhenDone:   branchWhenDone,
		dialogTestInputs: dialogTestInputs,
		dryRun:           dryRun,
		hasOpenChanges:   repoStatus.OpenChanges,
		initialBranch:    branchesSnapshot.Active,
		previousBranch:   previousBranch,
	}, branchesSnapshot, stashSnapshot, false, nil
}

func (self killConfig) branchToKillParent() gitdomain.LocalBranchName {
	return self.Lineage.Parent(self.branchToKill.LocalName)
}

func killProgram(config *killConfig) (runProgram, finalUndoProgram program.Program) {
	prog := program.Program{}
	killFeatureBranch(&prog, &finalUndoProgram, *config)
	cmdhelpers.Wrap(&prog, cmdhelpers.WrapOptions{
		DryRun:                   config.dryRun,
		RunInGitRoot:             true,
		StashOpenChanges:         config.initialBranch != config.branchToKill.LocalName && config.hasOpenChanges,
		PreviousBranchCandidates: gitdomain.LocalBranchNames{config.previousBranch, config.initialBranch},
	})
	return prog, finalUndoProgram
}

// killFeatureBranch kills the given feature branch everywhere it exists (locally and remotely).
func killFeatureBranch(prog *program.Program, finalUndoProgram *program.Program, config killConfig) {
	if config.branchToKill.HasTrackingBranch() && config.IsOnline() {
		prog.Add(&opcode.DeleteTrackingBranch{Branch: config.branchToKill.RemoteName})
	}
	if config.initialBranch == config.branchToKill.LocalName {
		if config.hasOpenChanges {
			prog.Add(&opcode.CommitOpenChanges{})
			// update the registered initial SHA for this branch so that undo restores the just committed changes
			prog.Add(&opcode.UpdateInitialBranchLocalSHA{Branch: config.initialBranch})
			// when undoing, manually undo the just committed changes so that they are uncommitted again
			finalUndoProgram.Add(&opcode.Checkout{Branch: config.branchToKill.LocalName})
			finalUndoProgram.Add(&opcode.UndoLastCommit{})
		}
		prog.Add(&opcode.Checkout{Branch: config.branchWhenDone})
	}
	prog.Add(&opcode.DeleteLocalBranch{Branch: config.branchToKill.LocalName, Force: false})
	if !config.dryRun {
		sync.RemoveBranchFromLineage(sync.RemoveBranchFromLineageArgs{
			Branch:  config.branchToKill.LocalName,
			Lineage: config.Lineage,
			Program: prog,
			Parent:  config.branchToKillParent(),
		})
	}
}
