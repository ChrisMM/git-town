package cmd

import (
	"fmt"

	"github.com/git-town/git-town/v11/src/cli/dialog"
	"github.com/git-town/git-town/v11/src/cli/flags"
	"github.com/git-town/git-town/v11/src/cli/print"
	"github.com/git-town/git-town/v11/src/cmd/cmdhelpers"
	"github.com/git-town/git-town/v11/src/config/configdomain"
	"github.com/git-town/git-town/v11/src/config/gitconfig"
	"github.com/git-town/git-town/v11/src/execute"
	"github.com/git-town/git-town/v11/src/git/gitdomain"
	"github.com/git-town/git-town/v11/src/gohacks/slice"
	"github.com/git-town/git-town/v11/src/gohacks/stringslice"
	"github.com/git-town/git-town/v11/src/hosting"
	"github.com/git-town/git-town/v11/src/hosting/hostingdomain"
	"github.com/git-town/git-town/v11/src/messages"
	"github.com/git-town/git-town/v11/src/sync"
	"github.com/git-town/git-town/v11/src/validate"
	"github.com/git-town/git-town/v11/src/vm/interpreter"
	"github.com/git-town/git-town/v11/src/vm/opcode"
	"github.com/git-town/git-town/v11/src/vm/program"
	"github.com/git-town/git-town/v11/src/vm/runstate"
	"github.com/spf13/cobra"
)

const shipDesc = "Deliver a completed feature branch"

const shipHelp = `
Squash-merges the current branch, or <branch_name> if given, into the main branch, resulting in linear history on the main branch.

- syncs the main branch
- pulls updates for <branch_name>
- merges the main branch into <branch_name>
- squash-merges <branch_name> into the main branch
  with commit message specified by the user
- pushes the main branch to the origin repository
- deletes <branch_name> from the local and origin repositories

Ships direct children of the main branch. To ship a nested child branch, ship or kill all ancestor branches first.

If you use GitHub, this command can squash merge pull requests via the GitHub API. Setup:

1. Get a GitHub personal access token with the "repo" scope
2. Run 'git config %s <token>' (optionally add the '--global' flag)

Now anytime you ship a branch with a pull request on GitHub, it will squash merge via the GitHub API. It will also update the base branch for any pull requests against that branch.

If your origin server deletes shipped branches, for example GitHub's feature to automatically delete head branches, run "git config %s false" and Git Town will leave it up to your origin server to delete the tracking branch of the branch you are shipping.`

func shipCmd() *cobra.Command {
	addVerboseFlag, readVerboseFlag := flags.Verbose()
	addMessageFlag, readMessageFlag := flags.String("message", "m", "", "Specify the commit message for the squash commit")
	addDryRunFlag, readDryRunFlag := flags.DryRun()
	cmd := cobra.Command{
		Use:     "ship",
		GroupID: "basic",
		Args:    cobra.MaximumNArgs(1),
		Short:   shipDesc,
		Long:    cmdhelpers.Long(shipDesc, fmt.Sprintf(shipHelp, gitconfig.KeyGithubToken, gitconfig.KeyShipDeleteTrackingBranch)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeShip(args, readMessageFlag(cmd), readDryRunFlag(cmd), readVerboseFlag(cmd))
		},
	}
	addDryRunFlag(&cmd)
	addVerboseFlag(&cmd)
	addMessageFlag(&cmd)
	return &cmd
}

func executeShip(args []string, message string, dryRun, verbose bool) error {
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
	config, initialBranchesSnapshot, initialStashSnapshot, exit, err := determineShipConfig(args, repo, dryRun, verbose)
	if err != nil || exit {
		return err
	}
	if config.branchToShip.LocalName == config.initialBranch {
		repoStatus, err := repo.Runner.Backend.RepoStatus()
		if err != nil {
			return err
		}
		err = validate.NoOpenChanges(repoStatus.OpenChanges)
		if err != nil {
			return err
		}
	}
	runState := runstate.RunState{
		Command:             "ship",
		DryRun:              dryRun,
		InitialActiveBranch: initialBranchesSnapshot.Active,
		RunProgram:          shipProgram(config, message),
	}
	return interpreter.Execute(interpreter.ExecuteArgs{
		FullConfig:              config.FullConfig,
		RunState:                &runState,
		Run:                     repo.Runner,
		Connector:               config.connector,
		DialogTestInputs:        &config.dialogTestInputs,
		Verbose:                 verbose,
		RootDir:                 repo.RootDir,
		InitialBranchesSnapshot: initialBranchesSnapshot,
		InitialConfigSnapshot:   repo.ConfigSnapshot,
		InitialStashSnapshot:    initialStashSnapshot,
	})
}

type shipConfig struct {
	*configdomain.FullConfig
	allBranches              gitdomain.BranchInfos
	branchToShip             gitdomain.BranchInfo
	connector                hostingdomain.Connector
	dialogTestInputs         dialog.TestInputs
	dryRun                   bool
	initialBranch            gitdomain.LocalBranchName
	targetBranch             gitdomain.BranchInfo
	canShipViaAPI            bool
	childBranches            gitdomain.LocalBranchNames
	proposalMessage          string
	hasOpenChanges           bool
	remotes                  gitdomain.Remotes
	isShippingInitialBranch  bool
	previousBranch           gitdomain.LocalBranchName
	proposal                 *hostingdomain.Proposal
	proposalsOfChildBranches []hostingdomain.Proposal
}

func determineShipConfig(args []string, repo *execute.OpenRepoResult, dryRun, verbose bool) (*shipConfig, gitdomain.BranchesStatus, gitdomain.StashSize, bool, error) {
	branchesSnapshot, stashSnapshot, dialogTestInputs, exit, err := execute.LoadRepoSnapshot(execute.LoadBranchesArgs{
		FullConfig:            &repo.Runner.FullConfig,
		Repo:                  repo,
		Verbose:               verbose,
		Fetch:                 true,
		HandleUnfinishedState: true,
		ValidateIsConfigured:  true,
		ValidateNoOpenChanges: len(args) == 0,
	})
	if err != nil || exit {
		return nil, branchesSnapshot, stashSnapshot, exit, err
	}
	previousBranch := repo.Runner.Backend.PreviouslyCheckedOutBranch()
	repoStatus, err := repo.Runner.Backend.RepoStatus()
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	remotes, err := repo.Runner.Backend.Remotes()
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	branchNameToShip := gitdomain.NewLocalBranchName(slice.FirstElementOr(args, branchesSnapshot.Active.String()))
	branchToShip := branchesSnapshot.Branches.FindByLocalName(branchNameToShip)
	if branchToShip != nil && branchToShip.SyncStatus == gitdomain.SyncStatusOtherWorktree {
		return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.ShipBranchOtherWorktree, branchNameToShip)
	}
	isShippingInitialBranch := branchNameToShip == branchesSnapshot.Active
	if !isShippingInitialBranch {
		if branchToShip == nil {
			return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.BranchDoesntExist, branchNameToShip)
		}
	}
	if !repo.Runner.IsFeatureBranch(branchNameToShip) {
		return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.ShipNoFeatureBranch, branchNameToShip)
	}
	err = execute.EnsureKnownBranchAncestry(branchNameToShip, execute.EnsureKnownBranchAncestryArgs{
		Config:           &repo.Runner.FullConfig,
		AllBranches:      branchesSnapshot.Branches,
		DefaultBranch:    repo.Runner.MainBranch,
		DialogTestInputs: &dialogTestInputs,
		Runner:           repo.Runner,
	})
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	err = ensureParentBranchIsMainOrPerennialBranch(branchNameToShip, &repo.Runner.FullConfig, repo.Runner.Lineage)
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	targetBranchName := repo.Runner.Lineage.Parent(branchNameToShip)
	targetBranch := branchesSnapshot.Branches.FindByLocalName(targetBranchName)
	if targetBranch == nil {
		return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.BranchDoesntExist, targetBranchName)
	}
	var proposal *hostingdomain.Proposal
	childBranches := repo.Runner.Lineage.Children(branchNameToShip)
	proposalsOfChildBranches := []hostingdomain.Proposal{}
	originURL := repo.Runner.Config.OriginURL()
	hostingService, err := repo.Runner.Config.HostingService()
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	connector, err := hosting.NewConnector(hosting.NewConnectorArgs{
		FullConfig:     &repo.Runner.FullConfig,
		HostingService: hostingService,
		OriginURL:      originURL,
		Log:            print.Logger{},
	})
	if err != nil {
		return nil, branchesSnapshot, stashSnapshot, false, err
	}
	canShipViaAPI := false
	proposalMessage := ""
	if !repo.IsOffline && connector != nil {
		if branchToShip.HasTrackingBranch() {
			proposal, err = connector.FindProposal(branchNameToShip, targetBranchName)
			if err != nil {
				return nil, branchesSnapshot, stashSnapshot, false, err
			}
			if proposal != nil {
				canShipViaAPI = true
				proposalMessage = connector.DefaultProposalMessage(*proposal)
			}
		}
		for _, childBranch := range childBranches {
			childProposal, err := connector.FindProposal(childBranch, branchNameToShip)
			if err != nil {
				return nil, branchesSnapshot, stashSnapshot, false, fmt.Errorf(messages.ProposalNotFoundForBranch, branchNameToShip, err)
			}
			if childProposal != nil {
				proposalsOfChildBranches = append(proposalsOfChildBranches, *childProposal)
			}
		}
	}
	return &shipConfig{
		FullConfig:               &repo.Runner.FullConfig,
		allBranches:              branchesSnapshot.Branches,
		connector:                connector,
		dialogTestInputs:         dialogTestInputs,
		dryRun:                   dryRun,
		initialBranch:            branchesSnapshot.Active,
		targetBranch:             *targetBranch,
		branchToShip:             *branchToShip,
		canShipViaAPI:            canShipViaAPI,
		childBranches:            childBranches,
		proposalMessage:          proposalMessage,
		hasOpenChanges:           repoStatus.OpenChanges,
		remotes:                  remotes,
		isShippingInitialBranch:  isShippingInitialBranch,
		previousBranch:           previousBranch,
		proposal:                 proposal,
		proposalsOfChildBranches: proposalsOfChildBranches,
	}, branchesSnapshot, stashSnapshot, false, nil
}

func ensureParentBranchIsMainOrPerennialBranch(branch gitdomain.LocalBranchName, config *configdomain.FullConfig, lineage configdomain.Lineage) error {
	parentBranch := lineage.Parent(branch)
	if config.IsFeatureBranch(parentBranch) {
		ancestors := lineage.Ancestors(branch)
		ancestorsWithoutMainOrPerennial := ancestors[1:]
		oldestAncestor := ancestorsWithoutMainOrPerennial[0]
		return fmt.Errorf(messages.ShipChildBranch, stringslice.Connect(ancestorsWithoutMainOrPerennial.Strings()), oldestAncestor)
	}
	return nil
}

func shipProgram(config *shipConfig, commitMessage string) program.Program {
	prog := program.Program{}
	if config.SyncBeforeShip {
		// sync the parent branch
		sync.BranchProgram(config.targetBranch, sync.BranchProgramArgs{
			Config:      config.FullConfig,
			BranchInfos: config.allBranches,
			Remotes:     config.remotes,
			Program:     &prog,
			PushBranch:  true,
		})
		// sync the branch to ship (local sync only)
		sync.BranchProgram(config.branchToShip, sync.BranchProgramArgs{
			Config:      config.FullConfig,
			BranchInfos: config.allBranches,
			Remotes:     config.remotes,
			Program:     &prog,
			PushBranch:  false,
		})
	}
	prog.Add(&opcode.EnsureHasShippableChanges{Branch: config.branchToShip.LocalName, Parent: config.MainBranch})
	prog.Add(&opcode.Checkout{Branch: config.targetBranch.LocalName})
	if config.canShipViaAPI {
		// update the proposals of child branches
		for _, childProposal := range config.proposalsOfChildBranches {
			prog.Add(&opcode.UpdateProposalTarget{
				ProposalNumber: childProposal.Number,
				NewTarget:      config.targetBranch.LocalName,
			})
		}
		prog.Add(&opcode.PushCurrentBranch{CurrentBranch: config.branchToShip.LocalName})
		prog.Add(&opcode.ConnectorMergeProposal{
			Branch:          config.branchToShip.LocalName,
			ProposalNumber:  config.proposal.Number,
			CommitMessage:   commitMessage,
			ProposalMessage: config.proposalMessage,
		})
		prog.Add(&opcode.PullCurrentBranch{})
	} else {
		prog.Add(&opcode.SquashMerge{Branch: config.branchToShip.LocalName, CommitMessage: commitMessage, Parent: config.targetBranch.LocalName})
	}
	if config.remotes.HasOrigin() && config.IsOnline() {
		prog.Add(&opcode.PushCurrentBranch{CurrentBranch: config.targetBranch.LocalName})
	}
	// NOTE: when shipping via API, we can always delete the tracking branch because:
	// - we know we have a tracking branch (otherwise there would be no PR to ship via API)
	// - we have updated the PRs of all child branches (because we have API access)
	// - we know we are online
	if config.canShipViaAPI || (config.branchToShip.HasTrackingBranch() && len(config.childBranches) == 0 && config.IsOnline()) {
		if config.ShipDeleteTrackingBranch {
			prog.Add(&opcode.DeleteTrackingBranch{Branch: config.branchToShip.RemoteName})
		}
	}
	prog.Add(&opcode.DeleteLocalBranch{Branch: config.branchToShip.LocalName, Force: false})
	if !config.dryRun {
		prog.Add(&opcode.DeleteParentBranch{Branch: config.branchToShip.LocalName})
	}
	for _, child := range config.childBranches {
		prog.Add(&opcode.ChangeParent{Branch: child, Parent: config.targetBranch.LocalName})
	}
	if !config.isShippingInitialBranch {
		prog.Add(&opcode.Checkout{Branch: config.initialBranch})
	}
	cmdhelpers.Wrap(&prog, cmdhelpers.WrapOptions{
		DryRun:                   config.dryRun,
		RunInGitRoot:             true,
		StashOpenChanges:         !config.isShippingInitialBranch && config.hasOpenChanges,
		PreviousBranchCandidates: gitdomain.LocalBranchNames{config.previousBranch},
	})
	return prog
}
