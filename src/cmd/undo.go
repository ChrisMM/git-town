package cmd

import (
	"fmt"

	"github.com/git-town/git-town/v7/src/git"
	"github.com/git-town/git-town/v7/src/runstate"
	"github.com/spf13/cobra"
)

const undoDesc = "Undoes the last run git-town command"

func undoCmd(repo *git.ProdRepo) *cobra.Command {
	return &cobra.Command{
		Use:     "undo",
		GroupID: "errors",
		Args:    cobra.NoArgs,
		PreRunE: ensure(repo, hasGitVersion, isRepository, isConfigured),
		Short:   undoDesc,
		Long:    long(undoDesc),
		RunE: func(cmd *cobra.Command, args []string) error {
			return undo(repo)
		},
	}
}

func undo(repo *git.ProdRepo) error {
	runState, err := runstate.Load(repo)
	if err != nil {
		return fmt.Errorf("cannot load previous run state: %w", err)
	}
	if runState == nil || runState.IsUnfinished() {
		return fmt.Errorf("nothing to undo")
	}
	undoRunState := runState.CreateUndoRunState()
	return runstate.Execute(&undoRunState, repo, nil)
}
