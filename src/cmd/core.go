// Package cmd defines the Git Town commands.
package cmd

import (
	"github.com/git-town/git-town/v11/src/cmd/config"
	"github.com/git-town/git-town/v11/src/cmd/debug"
)

// Execute runs the Cobra stack.
func Execute() error {
	rootCmd := rootCmd()
	rootCmd.AddCommand(aliasesCommand())
	rootCmd.AddCommand(appendCmd())
	rootCmd.AddCommand(completionsCmd(&rootCmd))
	rootCmd.AddCommand(config.RootCmd())
	rootCmd.AddCommand(continueCmd())
	rootCmd.AddCommand(debug.RootCmd())
	rootCmd.AddCommand(diffParentCommand())
	rootCmd.AddCommand(hackCmd())
	rootCmd.AddCommand(killCommand())
	rootCmd.AddCommand(newPullRequestCommand())
	rootCmd.AddCommand(proposeCommand())
	rootCmd.AddCommand(prependCommand())
	rootCmd.AddCommand(renameBranchCommand())
	rootCmd.AddCommand(repoCommand())
	rootCmd.AddCommand(statusCommand())
	rootCmd.AddCommand(setParentCommand())
	rootCmd.AddCommand(shipCmd())
	rootCmd.AddCommand(skipCmd())
	rootCmd.AddCommand(switchCmd())
	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(undoCmd())
	return rootCmd.Execute()
}
