package cmd

import (
	"github.com/git-town/git-town/v7/src/git"
	"github.com/git-town/git-town/v7/src/runstate"
	"github.com/git-town/git-town/v7/src/steps"
	"github.com/git-town/git-town/v7/src/validate"
	"github.com/spf13/cobra"
)

const appendDesc = "Creates a new feature branch as a child of the current branch"

const appendHelp = `
Syncs the current branch,
forks a new feature branch with the given name off the current branch,
makes the new branch a child of the current branch,
pushes the new feature branch to the origin repository
(if and only if "push-new-branches" is true),
and brings over all uncommitted changes to the new feature branch.

See "sync" for information regarding upstream remotes.`

func appendCmd(repo *git.ProdRepo) *cobra.Command {
	return &cobra.Command{
		Use:     "append <branch>",
		GroupID: "lineage",
		Args:    cobra.ExactArgs(1),
		PreRunE: ensure(repo, hasGitVersion, isRepository, isConfigured),
		Short:   appendDesc,
		Long:    long(appendDesc, appendHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAppend(args, repo)
		},
	}
}

func runAppend(args []string, repo *git.ProdRepo) error {
	config, err := determineAppendConfig(args, repo)
	if err != nil {
		return err
	}
	stepList, err := appendStepList(config, repo)
	if err != nil {
		return err
	}
	runState := runstate.New("append", stepList)
	return runstate.Execute(runState, repo, nil)
}

type appendConfig struct {
	ancestorBranches    []string
	hasOrigin           bool
	isOffline           bool
	noPushHook          bool
	parentBranch        string
	shouldNewBranchPush bool
	targetBranch        string
}

func determineAppendConfig(args []string, repo *git.ProdRepo) (*appendConfig, error) {
	ec := runstate.ErrorChecker{}
	parentBranch := ec.String(repo.Silent.CurrentBranch())
	hasOrigin := ec.Bool(repo.Silent.HasOrigin())
	isOffline := ec.Bool(repo.Config.IsOffline())
	pushHook := ec.Bool(repo.Config.PushHook())
	shouldNewBranchPush := ec.Bool(repo.Config.ShouldNewBranchPush())
	targetBranch := args[0]
	if ec.Err != nil {
		return nil, ec.Err
	}
	if hasOrigin && !isOffline {
		ec.Check(repo.Logging.Fetch())
	}
	hasTargetBranch := ec.Bool(repo.Silent.HasLocalOrOriginBranch(targetBranch))
	if hasTargetBranch {
		ec.Fail("a branch named %q already exists", targetBranch)
	}
	ec.Check(validate.KnowsBranchAncestry(parentBranch, repo.Config.MainBranch(), repo))
	ancestorBranches := repo.Config.AncestorBranches(parentBranch)
	return &appendConfig{
		ancestorBranches:    ancestorBranches,
		isOffline:           isOffline,
		hasOrigin:           hasOrigin,
		noPushHook:          !pushHook,
		parentBranch:        parentBranch,
		shouldNewBranchPush: shouldNewBranchPush,
		targetBranch:        targetBranch,
	}, ec.Err
}

func appendStepList(config *appendConfig, repo *git.ProdRepo) (runstate.StepList, error) {
	list := runstate.StepListBuilder{}
	for _, branch := range append(config.ancestorBranches, config.parentBranch) {
		updateBranchSteps(&list, branch, true, repo)
	}
	list.Add(&steps.CreateBranchStep{Branch: config.targetBranch, StartingPoint: config.parentBranch})
	list.Add(&steps.SetParentStep{Branch: config.targetBranch, ParentBranch: config.parentBranch})
	list.Add(&steps.CheckoutStep{Branch: config.targetBranch})
	if config.hasOrigin && config.shouldNewBranchPush && !config.isOffline {
		list.Add(&steps.CreateTrackingBranchStep{Branch: config.targetBranch, NoPushHook: config.noPushHook})
	}
	list.Wrap(runstate.WrapOptions{RunInGitRoot: true, StashOpenChanges: true}, repo)
	return list.Result()
}
