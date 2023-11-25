package opcode

import (
	"github.com/git-town/git-town/v10/src/config"
	"github.com/git-town/git-town/v10/src/vm/shared"
)

// AbortRebase represents aborting on ongoing merge conflict.
// This opcode is used in the abort scripts for Git Town commands.
type RemoveGitAlias struct {
	Alias config.Alias
	undeclaredOpcodeMethods
}

func (self *RemoveGitAlias) Run(args shared.RunArgs) error {
	return args.Runner.Frontend.RemoveGitAlias(self.Alias)
}
