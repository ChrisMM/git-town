package undostash

import (
	"github.com/git-town/git-town/v11/src/git/gitdomain"
	"github.com/git-town/git-town/v11/src/vm/opcode"
	"github.com/git-town/git-town/v11/src/vm/program"
)

// StashDiff describes the changes made to the Git stash.
type StashDiff struct {
	// the number of entries added to the Git stash (positive = entries added, negative = entries removed, 0 = nothing added)
	EntriesAdded int
}

func NewStashDiff(before, after gitdomain.StashSize) StashDiff {
	return StashDiff{
		EntriesAdded: int(after) - int(before),
	}
}

func (self StashDiff) Program() program.Program {
	result := program.Program{}
	for ; self.EntriesAdded > 0; self.EntriesAdded-- {
		result.Add(&opcode.RestoreOpenChanges{})
	}
	return result
}
