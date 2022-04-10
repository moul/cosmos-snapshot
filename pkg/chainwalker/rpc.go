package chainwalker

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"go.uber.org/zap"
)

type NewRPCWalkerOpts struct {
	RPCAddr          string
	MinHeight        int64
	MaxHeight        int64
	Logger           *zap.Logger
	Ctx              context.Context
	Timeout          time.Duration
	WithBlockResults bool
	WithoutBlockTxs  bool
}

func (opts *NewRPCWalkerOpts) applyDefaults() {
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	if opts.Ctx == nil {
		opts.Ctx = context.Background()
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}
}

// NewRPCWalker returns a `Walker` implementation that uses an RPC connection to perform the walk.
func NewRPCWalker(opts NewRPCWalkerOpts) (*rpcWalker, error) {
	if opts.RPCAddr == "" {
		return nil, errors.New("missing RPCAddr variable")
	}
	opts.applyDefaults()

	var (
		rpcAddr = opts.RPCAddr
		timeout = opts.Timeout
		logger  = opts.Logger
		ctx     = opts.Ctx
	)

	// init the RPC client
	httpClient, err := libclient.DefaultHTTPClient(rpcAddr)
	if err != nil {
		return nil, fmt.Errorf("new HTTP client: %w", err)
	}
	httpClient.Timeout = timeout
	rpcClient, err := rpchttp.NewWithClient(rpcAddr, "/websocket", httpClient)
	if err != nil {
		return nil, fmt.Errorf("new RPC client: %w", err)
	}
	logger.Debug("connected to remote RPC", zap.String("addr", rpcAddr))

	// check status and constraints
	status, err := rpcClient.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("get RPC Status: %w", err)
	}
	// fmt.Println(u.PrettyJSON(status))
	logger.Debug("fetched status",
		zap.String("network", status.NodeInfo.Network),
		zap.Int64("earliest-height", status.SyncInfo.EarliestBlockHeight),
		zap.Int64("latest-height", status.SyncInfo.LatestBlockHeight),
	)
	if opts.MinHeight <= 0 {
		opts.MinHeight = status.SyncInfo.EarliestBlockHeight
	}
	if opts.MaxHeight <= 0 {
		opts.MaxHeight = status.SyncInfo.LatestBlockHeight
	}
	if opts.MinHeight < status.SyncInfo.EarliestBlockHeight {
		return nil, fmt.Errorf("specified min-height is smaller than earliest chain block")
	}
	if opts.MaxHeight > status.SyncInfo.LatestBlockHeight {
		return nil, fmt.Errorf("specified max-height is larger than latest chain block")
	}

	walker := rpcWalker{
		client: rpcClient,
		opts:   opts,
	}
	return &walker, nil
}

type rpcWalker struct {
	client *rpchttp.HTTP
	opts   NewRPCWalkerOpts
}

func (walker *rpcWalker) Run(callback Callback) error {
	// FIXME: speedup with a channel queue?

	var (
		client           = walker.client
		logger           = walker.opts.Logger
		minHeight        = walker.opts.MinHeight
		maxHeight        = walker.opts.MaxHeight
		withBlockResults = walker.opts.WithBlockResults
		withBlockTxs     = !walker.opts.WithoutBlockTxs
		ctx              = walker.opts.Ctx
	)

	for height := minHeight; height <= maxHeight; height++ {
		// trigger a new height event
		err := callback(Entry{Height: height, Kind: EntryBlock})
		if err != nil {
			return fmt.Errorf("height block handler error: %w", err)
		}

		if withBlockResults {
			results, err := client.BlockResults(ctx, &height)
			if err != nil {
				// FIXME: retry policy, ignore?
				return fmt.Errorf("call BlockResults: %w", err)
			}
			for _, event := range results.BeginBlockEvents {
				err := callback(Entry{Height: height, BeginBlock: &event, Kind: EntryBeginBlock})
				if err != nil {
					return fmt.Errorf("begin block handler error: %w", err)
				}
			}
			for _, event := range results.EndBlockEvents {
				err := callback(Entry{Height: height, EndBlock: &event, Kind: EntryEndBlock})
				if err != nil {
					return fmt.Errorf("end block handler error: %w", err)
				}
			}
		}

		if withBlockTxs {
			block, err := walker.client.Block(ctx, &height)
			if err != nil {
				// FIXME: retry policy, ignore?
				return fmt.Errorf("call Block: %w", err)
			}
			// temporarily disabled while working on the incompatibilty between tendermint v0.34.13 and v0.35.x
			// trigger EntryBlock
			// err = callback(Entry{Height: height, Block: block, Kind: EntryBlock})
			// if err != nil {
			//	return fmt.Errorf("tx handler error: %w", err)
			// }

			for _, tx := range block.Block.Txs {
				if block.Block.Txs != nil {
					logger.Debug("  tx", zap.String("hash", fmt.Sprintf("%x", tx.Hash())))
				}
				/*
					txBytes, err := hex.DecodeString(tx.Hash())
					if err != nil {
						return fmt.Errorf("decode tx: %w", err)
					}
				*/
				_ = hex.Decode
				res, err := walker.client.Tx(ctx, tx.Hash(), false)
				if err != nil {
					return fmt.Errorf("call Tx: %w", err)
				}

				for _, event := range res.TxResult.Events {
					err := callback(Entry{Height: height, Tx: &event, Kind: EntryTx})
					if err != nil {
						return fmt.Errorf("tx handler error: %w", err)
					}
				}
			}
		}
	}
	return nil
}
