package cmd

import (
	"fmt"
	"strings"

	"github.com/git-town/git-town/v11/src/cli/flags"
	"github.com/git-town/git-town/v11/src/config/configdomain"
	"github.com/git-town/git-town/v11/src/domain"
	"github.com/git-town/git-town/v11/src/git"
	"github.com/git-town/git-town/v11/src/messages"
	"github.com/git-town/git-town/v11/src/vm/interpreter"
	"github.com/git-town/git-town/v11/src/vm/opcode"
	"github.com/git-town/git-town/v11/src/vm/program"
	"github.com/git-town/git-town/v11/src/vm/runstate"
	"github.com/spf13/cobra"
)

const aliasesDesc = "Adds or removes default global aliases"

const aliasesHelp = `
Global aliases make Git Town commands feel like native Git commands.
When enabled, you can run "git hack" instead of "git town hack".

Does not overwrite existing aliases.

This can conflict with other tools that also define Git aliases.`

func aliasesCommand() *cobra.Command {
	addVerboseFlag, readVerboseFlag := flags.Verbose()
	cmd := cobra.Command{
		Use:     "aliases (add | remove)",
		GroupID: "setup",
		Args:    cobra.ExactArgs(1),
		Short:   aliasesDesc,
		Long:    long(aliasesDesc, aliasesHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeAliases(args[0], readVerboseFlag(cmd))
		},
	}
	addVerboseFlag(&cmd)
	return &cmd
}

func executeAliases(arg string, verbose bool) error {
	runState := runstate.RunState{
		Command:             "aliases",
		InitialActiveBranch: domain.EmptyLocalBranchName(),
		RunProgram:          aliasesProgram(arg),
	}
	return interpreter.Execute(interpreter.ExecuteArgs{
		RunState:                &runState,
		Run:                     &repo.Runner,
		Connector:               nil,
		Verbose:                 verbose,
		RootDir:                 repo.RootDir,
		InitialBranchesSnapshot: initialBranchesSnapshot,
		InitialConfigSnapshot:   repo.ConfigSnapshot,
		InitialStashSnapshot:    initialStashSnapshot,
		Lineage:                 config.lineage,
		NoPushHook:              !config.pushHook,
	})
}

func aliasesProgram(arg string) program.Program {
	prog := program.Program{}
	switch strings.ToLower(arg) {
	case "add":
		return addAliases(repo.Runner)
	case "remove":
		return removeAliases(repo.Runner)
	default:
		return fmt.Errorf(messages.InputAddOrRemove, arg)
	}
	return prog
}

func addAliasesProgram(prog *program.Program) {
	for _, alias := range configdomain.Aliases() {
		prog.Add(&opcode.AddGitAlias{Alias: alias})
	}
}

func removeAliasesProgram(prog *program.Program, run *git.ProdRunner) {
	for _, alias := range configdomain.Aliases() {
		existingAlias := run.GitTown.GitAlias(alias)
		if existingAlias == "town "+alias.String() {
			prog.Add(&opcode.RemoveGitAlias{Alias: alias})
		}
	}
}
