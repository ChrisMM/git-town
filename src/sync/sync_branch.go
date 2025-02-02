package sync

import (
	"github.com/git-town/git-town/v11/src/config/configdomain"
	"github.com/git-town/git-town/v11/src/git/gitdomain"
	"github.com/git-town/git-town/v11/src/vm/opcode"
	"github.com/git-town/git-town/v11/src/vm/program"
)

// BranchProgram syncs the given branch.
func BranchProgram(branch gitdomain.BranchInfo, args BranchProgramArgs) {
	parentBranchInfo := args.BranchInfos.FindByLocalName(args.Config.Lineage.Parent(branch.LocalName))
	parentOtherWorktree := parentBranchInfo != nil && parentBranchInfo.SyncStatus == gitdomain.SyncStatusOtherWorktree
	switch {
	case branch.SyncStatus == gitdomain.SyncStatusDeletedAtRemote:
		syncDeletedBranchProgram(args.Program, branch, parentOtherWorktree, args)
	case branch.SyncStatus == gitdomain.SyncStatusOtherWorktree:
		// Git Town doesn't sync branches that are active in another worktree
	default:
		ExistingBranchProgram(args.Program, branch, parentOtherWorktree, args)
	}
}

type BranchProgramArgs struct {
	Config      *configdomain.FullConfig
	BranchInfos gitdomain.BranchInfos
	Program     *program.Program
	PushBranch  bool
	Remotes     gitdomain.Remotes
}

// ExistingBranchProgram provides the opcode to sync a particular branch.
func ExistingBranchProgram(list *program.Program, branch gitdomain.BranchInfo, parentOtherWorktree bool, args BranchProgramArgs) {
	isFeatureBranch := args.Config.IsFeatureBranch(branch.LocalName)
	if !isFeatureBranch && !args.Remotes.HasOrigin() {
		// perennial branch but no remote --> this branch cannot be synced
		return
	}
	list.Add(&opcode.Checkout{Branch: branch.LocalName})
	if isFeatureBranch {
		FeatureBranchProgram(list, branch, parentOtherWorktree, args.Config.SyncFeatureStrategy)
	} else {
		PerennialBranchProgram(branch, args)
	}
	if args.PushBranch && args.Remotes.HasOrigin() && args.Config.IsOnline() {
		switch {
		case !branch.HasTrackingBranch():
			list.Add(&opcode.CreateTrackingBranch{Branch: branch.LocalName})
		case !isFeatureBranch:
			list.Add(&opcode.PushCurrentBranch{CurrentBranch: branch.LocalName})
		default:
			pushFeatureBranchProgram(list, branch.LocalName, args.Config.SyncFeatureStrategy)
		}
	}
}

// pullParentBranchOfCurrentFeatureBranchOpcode adds the opcode to pull updates from the parent branch of the current feature branch into the current feature branch.
func pullParentBranchOfCurrentFeatureBranchOpcode(list *program.Program, currentBranch gitdomain.LocalBranchName, parentOtherWorktree bool, strategy configdomain.SyncFeatureStrategy) {
	switch strategy {
	case configdomain.SyncFeatureStrategyMerge:
		list.Add(&opcode.MergeParent{CurrentBranch: currentBranch, ParentActiveInOtherWorktree: parentOtherWorktree})
	case configdomain.SyncFeatureStrategyRebase:
		list.Add(&opcode.RebaseParent{CurrentBranch: currentBranch, ParentActiveInOtherWorktree: parentOtherWorktree})
	}
}

// pullTrackingBranchOfCurrentFeatureBranchOpcode adds the opcode to pull updates from the remote branch of the current feature branch into the current feature branch.
func pullTrackingBranchOfCurrentFeatureBranchOpcode(list *program.Program, trackingBranch gitdomain.RemoteBranchName, strategy configdomain.SyncFeatureStrategy) {
	switch strategy {
	case configdomain.SyncFeatureStrategyMerge:
		list.Add(&opcode.Merge{Branch: trackingBranch.BranchName()})
	case configdomain.SyncFeatureStrategyRebase:
		list.Add(&opcode.RebaseBranch{Branch: trackingBranch.BranchName()})
	}
}

func pushFeatureBranchProgram(list *program.Program, branch gitdomain.LocalBranchName, syncFeatureStrategy configdomain.SyncFeatureStrategy) {
	switch syncFeatureStrategy {
	case configdomain.SyncFeatureStrategyMerge:
		list.Add(&opcode.PushCurrentBranch{CurrentBranch: branch})
	case configdomain.SyncFeatureStrategyRebase:
		list.Add(&opcode.ForcePushCurrentBranch{})
	}
}

// updateCurrentPerennialBranchOpcode provides the opcode to update the current perennial branch with changes from the given other branch.
func updateCurrentPerennialBranchOpcode(list *program.Program, otherBranch gitdomain.RemoteBranchName, strategy configdomain.SyncPerennialStrategy) {
	switch strategy {
	case configdomain.SyncPerennialStrategyMerge:
		list.Add(&opcode.Merge{Branch: otherBranch.BranchName()})
	case configdomain.SyncPerennialStrategyRebase:
		list.Add(&opcode.RebaseBranch{Branch: otherBranch.BranchName()})
	}
}
