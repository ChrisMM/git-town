package opcode

import "github.com/git-town/git-town/v11/src/vm/shared"

// PushTags pushes newly created Git tags to origin.
type PushTags struct {
	undeclaredOpcodeMethods
}

func (self *PushTags) Run(args shared.RunArgs) error {
	return args.Runner.Frontend.PushTags()
}
