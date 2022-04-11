package main

/*
 *
 * WARNING: WIP, all the code will be cleaned and refactored as soon as the PoC is finished.
 *
 */

// https://github.com/cosmos/ibc-go/blob/main/docs/ibc/events.md

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
	"moul.io/cosmos-snapshot/pkg/chainwalker"
	"moul.io/u"
)

// TODO:
// write helpers to compose a custom filter
//   whale-cap
//   min-invest
//   min-duration
//   has-voted-any-on-n-votes
//   has-voted-yes-on-this
//   has-voted-no-on-this
//   first-transaction-before-specific-date
//   any-activity-since-1y
//   regularly-active
//   has-stacked
//   has-not-stacked-on-blacklist
//   in-a-whitelist
//   not-in-a-blacklist
//   exception
//   ...

// Accountant contains the storage and implementation of a snapshot engine based on rules (in code).
type Accountant struct {
	Accounts AccountList
	Stats    struct {
		StartedAt        time.Time
		Duration         string
		TotalCalls       int
		TotalByKind      map[string]uint
		TotalByEventKind map[string]uint
	}
	Logger *zap.Logger
}

func (accountant *Accountant) init() {
	// temporarily disabled while we improve the code organization
	// bar = progressbar.NewOptions(int(flagOpts.MaxHeight-flagOpts.MinHeight), progressbar.OptionSetWriter(os.Stdout))

	if accountant.Logger == nil {
		accountant.Logger = zap.NewNop()
	}
	accountant.Accounts = make(AccountList)
	accountant.Stats.TotalByKind = make(map[string]uint)
	accountant.Stats.TotalByEventKind = make(map[string]uint)
	accountant.Stats.StartedAt = time.Now()
}

func (accountant *Accountant) callback(entry chainwalker.Entry) error {
	var (
		logger = accountant.Logger
	)

	accountant.Stats.TotalCalls++
	accountant.Stats.TotalByKind[entry.Kind.String()]++

	switch entry.Kind {
	case chainwalker.EntryBlock:
		logger.Debug(" block", zap.Int64("height", entry.Height))
	case chainwalker.EntryTx:
		event := entry.Tx
		accountant.Stats.TotalByEventKind["tx:"+event.Type]++
		logEntry := logger.With(zap.String("type", event.Type))
		attrs := map[string]string{}
		for _, v := range event.GetAttributes() {
			// tendermint 0.34.x
			key := bytes.NewBuffer(v.GetKey()).String()
			value := bytes.NewBuffer(v.GetValue()).String()
			// tendermint 0.35.x
			// key := v.GetKey()
			// value := v.GetValue()

			attrs[key] = value
			logEntry = logEntry.With(zap.String(key, value))
		}
		switch event.Type {
		case "transfer":
			// accountant.Accounts.ByID()
			recipient := accountant.Accounts.ByID(attrs["recipient"])
			sender := accountant.Accounts.ByID(attrs["sender"])
			// FIXME: parse amount value and token
			// amount := attrs["amount"]
			// recipient.Balance += ...
			// sender.Balance -= ...
			recipient.Txs -= 1
			sender.Txs += 1
		case "message":
		case "unbond":
		case "withdraw_commission":
		case "withdraw_rewards":
		case "delegate":
		case "redelegate":
		case "set_withdraw_address":
		case "edit_validator":
		case "create_client":
		case "proposal_vote":
		case "update_client":
		case "update_client_proposal":
		case "client_misbehaviour":
		case "send_packet":
		case "ibc_transfer":
		case "acknowledge_packet":
		case "fungible_token_packet":
		case "recv_packet":
		case "denomination_trace":
		case "write_acknowledgement":
		case "connection_open_try":
		case "connection_open_confirm":
		case "channel_open_try":
		case "channel_open_confirm":
		case "channel_open_init":
		case "channel_open_ack":
		case "channel_close_confirm":
		case "channel_close_init":
		case "timeout_packet":
		default:
			log.Fatalf("unknown TX event type: %q", event.Type)
		}
		logEntry.Debug("   tx event")
	case chainwalker.EntryBeginBlock:
		event := *entry.BeginBlock
		accountant.Stats.TotalByEventKind["bbegin:"+event.Type]++
		logEntry := logger.With(zap.String("type", event.Type))
		switch event.Type {
		case "liveness":
		case "commission":
		case "rewards":
		case "transfer":
		case "message":
		case "mint":
		case "proposer_reward":
		case "slash":
		default:
			log.Fatalf("unknown begin event type: %q", event.Type)
		}
		for _, v := range event.GetAttributes() {
			// tendermint 0.34.x
			key := bytes.NewBuffer(v.GetKey()).String()
			value := bytes.NewBuffer(v.GetValue()).String()
			// tendermint 0.35.x
			// key := v.GetKey()
			// value := v.GetValue()

			logEntry = logEntry.With(zap.String(key, value))
		}
		logEntry.Debug("  begin block event")
	case chainwalker.EntryEndBlock:
		event := entry.EndBlock
		accountant.Stats.TotalByEventKind["bend:"+event.Type]++
		logEntry := logger.With(zap.String("type", event.Type))
		switch event.Type {
		case "complete_unbonding":
		case "complete_redelegation":
		case "transfer":
		case "message":
		default:
			log.Fatalf("unknown end block event type: %q", event.Type)
		}
		for _, v := range event.GetAttributes() {
			// tendermint 0.34.x
			key := bytes.NewBuffer(v.GetKey()).String()
			value := bytes.NewBuffer(v.GetValue()).String()
			// tendermint 0.35.x
			// key := v.GetKey()
			// value := v.GetValue()

			logEntry = logEntry.With(zap.String(key, value))
		}
		logEntry.Debug("  end block event")
	default:
		return fmt.Errorf("Unsupported chainwalker.Entry kind: %q", entry.Kind)
	}
	return nil
}

func (accountant *Accountant) printResults() {
	fmt.Println("# Results:")
	fmt.Println(u.PrettyJSON(accountant.Accounts))
	fmt.Println("# Stats:")
	accountant.Stats.Duration = fmt.Sprintf("%v", time.Since(accountant.Stats.StartedAt))
	fmt.Println(u.PrettyJSON(accountant.Stats))
}

// FIXME: reuse something official here
type Account struct {
	Balance int64
	Txs     int64
}

type AccountList map[string]*Account

func (list *AccountList) ByID(id string) *Account {
	_, found := (*list)[id]
	if !found {
		(*list)[id] = &Account{}
	}
	return (*list)[id]
}
