package opcode

import "github.com/git-town/git-town/v11/src/vm/shared"

// DiscardOpenChanges resets the branch to the last commit, discarding uncommitted changes.
type DiscardOpenChanges struct {
	undeclaredOpcodeMethods
}

func (self *DiscardOpenChanges) Run(args shared.RunArgs) error {
	return args.Runner.Frontend.DiscardOpenChanges()
}
