package chainwalker

import (
	"fmt"

	"github.com/tendermint/tendermint/abci/types"
)

type Walker interface {
	Run(Callback) error
}

type Callback func(Entry) error

type Entry struct {
	Height     int64
	BeginBlock *types.Event
	EndBlock   *types.Event
	Tx         *types.Event
	Kind       EntryKind
	// Block      *coretypes.ResultBlock
}

type EntryKind uint

const (
	EntryZero       EntryKind = iota
	EntryHeight     EntryKind = iota
	EntryTx         EntryKind = iota
	EntryBlock      EntryKind = iota
	EntryBeginBlock EntryKind = iota
	EntryEndBlock   EntryKind = iota
)

func (k EntryKind) String() string {
	mapping := map[EntryKind]string{
		EntryHeight:     "height",
		EntryTx:         "tx",
		EntryBlock:      "block",
		EntryBeginBlock: "bbegin",
		EntryEndBlock:   "bend",
	}
	val, found := mapping[k]
	if !found {
		return fmt.Sprintf("%d", k)
	}
	return val
}
