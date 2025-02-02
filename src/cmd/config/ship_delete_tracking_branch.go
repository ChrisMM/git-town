package config

import (
	"fmt"

	"github.com/git-town/git-town/v11/src/cli/flags"
	"github.com/git-town/git-town/v11/src/cli/format"
	"github.com/git-town/git-town/v11/src/cli/io"
	"github.com/git-town/git-town/v11/src/cmd/cmdhelpers"
	"github.com/git-town/git-town/v11/src/config/configdomain"
	"github.com/git-town/git-town/v11/src/execute"
	"github.com/git-town/git-town/v11/src/git"
	"github.com/git-town/git-town/v11/src/gohacks"
	"github.com/git-town/git-town/v11/src/messages"
	"github.com/spf13/cobra"
)

const shipDeleteTrackingBranchDesc = `Displays or changes whether "git ship" deletes tracking branches`

const shipDeleteTrackingBranchHelp = `
If "ship-delete-tracking-branches" is enabled, the "git ship" command deletes the tracking branch of the branch it ships.`

func shipDeleteTrackingBranchCommand() *cobra.Command {
	addVerboseFlag, readVerboseFlag := flags.Verbose()
	addGlobalFlag, readGlobalFlag := flags.Bool("global", "g", "If set, reads or updates the ship-delete-tracking-branch strategy for all repositories on this machine", flags.FlagTypeNonPersistent)
	cmd := cobra.Command{
		Use:   "ship-delete-tracking-branch [--global] [(yes | no)]",
		Args:  cobra.MaximumNArgs(1),
		Short: shipDeleteTrackingBranchDesc,
		Long:  cmdhelpers.Long(shipDeleteTrackingBranchDesc, shipDeleteTrackingBranchHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeShipDeleteTrackingBranchanches(args, readGlobalFlag(cmd), readVerboseFlag(cmd))
		},
	}
	addVerboseFlag(&cmd)
	addGlobalFlag(&cmd)
	return &cmd
}

func executeShipDeleteTrackingBranchanches(args []string, global, verbose bool) error {
	repo, err := execute.OpenRepo(execute.OpenRepoArgs{
		Verbose:          verbose,
		DryRun:           false,
		OmitBranchNames:  true,
		PrintCommands:    true,
		ValidateIsOnline: false,
		ValidateGitRepo:  false,
	})
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return setShipDeleteTrackingBranch(args[0], global, repo.Runner)
	}
	return printShipDeleteTrackingBranch(global, repo.Runner)
}

func printShipDeleteTrackingBranch(globalFlag bool, run *git.ProdRunner) error {
	var setting *configdomain.ShipDeleteTrackingBranch
	if globalFlag {
		setting = run.Config.GlobalGitConfig.ShipDeleteTrackingBranch
		if setting == nil {
			defaults := configdomain.DefaultConfig()
			setting = &defaults.ShipDeleteTrackingBranch
		}
	} else {
		setting = &run.Config.ShipDeleteTrackingBranch
	}
	io.Println(format.Bool(setting.Bool()))
	return nil
}

func setShipDeleteTrackingBranch(text string, globalFlag bool, run *git.ProdRunner) error {
	boolValue, err := gohacks.ParseBool(text)
	if err != nil {
		return fmt.Errorf(messages.InputYesOrNo, text)
	}
	return run.Config.SetShipDeleteTrackingBranch(configdomain.ShipDeleteTrackingBranch(boolValue), globalFlag)
}
