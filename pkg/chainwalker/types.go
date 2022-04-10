package chainwalker

import (
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
