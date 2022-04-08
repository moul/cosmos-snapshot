package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/peterbourgon/ff"
	"github.com/schollz/progressbar/v3"
	"github.com/tendermint/tendermint/rpc/client/http"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"moul.io/zapconfig"
)

// TODO:
// support lambda style custom filter
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

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %+v\n", err)
	}
}

func run() error {
	var config struct {
		minHeight int64
		maxHeight int64
		rpcAddr   string
		debug     bool
	}
	var (
		ctx = context.Background()
	)

	// parse CLI flags.
	{
		fs := flag.NewFlagSet("cosmos-snapshot", flag.ContinueOnError)
		fs.Int64Var(&config.minHeight, "min-height", 5200791, "first block to process")
		fs.Int64Var(&config.maxHeight, "max-height", 5797010, "last block to process")
		fs.StringVar(&config.rpcAddr, "rpc-addr", "http://localhost:26657", "Cosmos RPC Address")
		fs.BoolVar(&config.debug, "debug", false, "verbose output")
		err := ff.Parse(fs, os.Args[1:])
		if err != nil {
			return fmt.Errorf("flag parse error: %w", err)
		}
	}

	// init logger
	var logger *zap.Logger
	{
		zapconf := zapconfig.New().EnableStacktrace().SetPreset("light-console")
		if config.debug {
			zapconf.SetLevel(zapcore.DebugLevel)
		} else {
			zapconf.SetLevel(zapcore.InfoLevel)
		}
		logger = zapconf.MustBuild()
		logger = logger.WithOptions(zap.WithCaller(false))
		logger.Debug("starting")
	}

	// init client
	var client *http.HTTP
	{

		var err error
		client, err = newRPCClient("http://localhost:26657")
		if err != nil {
			return fmt.Errorf("new RPC client: %w", err)
		}
		logger.Debug("connected to remote RPC", zap.String("addr", config.rpcAddr))
	}

	// check status
	{
		status, err := client.Status(ctx)
		if err != nil {
			return fmt.Errorf("get RPC Status: %w", err)
		}
		logger.Debug("fetched status",
			zap.String("network", status.NodeInfo.Network),
			zap.Int64("earliest-height", status.SyncInfo.EarliestBlockHeight),
			zap.Int64("latest-height", status.SyncInfo.LatestBlockHeight),
		)
		if config.minHeight < status.SyncInfo.EarliestBlockHeight {
			return fmt.Errorf("specified min-height is smaller than earliest chain block")
		}
		if config.maxHeight > status.SyncInfo.LatestBlockHeight {
			return fmt.Errorf("specified max-height is larger than latest chain block")
		}
		// fmt.Println(u.PrettyJSON(status))
	}

	// iterate over blocks
	{
		var (
			start            = time.Now()
			bar              = progressbar.Default(config.maxHeight - config.minHeight)
			eventsByType     = make(map[string]int)
			totalBlocks      = 0
			totalBlockEvents = 0
			totalTxs         = 0
			totalTxEvents    = 0
		)

		// FIXME: speedup with a channel queue
		for height := config.minHeight; height <= config.maxHeight; height++ {
			logger.Debug(" block", zap.Int64("height", height))
			results, err := client.BlockResults(ctx, &height)
			if err != nil {
				// FIXME: retry policy, ignore?
				panic(err)
			}
			for _, event := range results.BeginBlockEvents {
				continue
				logEntry := logger.With(zap.String("type", event.Type))
				eventsByType["bbegin:"+event.Type]++
				switch event.Type {
				case "liveness":
				case "commission":
				case "rewards":
				case "transfer":
				case "message":
				case "mint":
				case "proposer_reward":
				default:
					log.Fatalf("unknown event type: %q", event.Type)
				}
				for _, v := range event.GetAttributes() {
					key := bytes.NewBuffer(v.GetKey()).String()
					value := bytes.NewBuffer(v.GetValue()).String()
					logEntry = logEntry.With(zap.String(key, value))
				}
				logEntry.Debug("  begin block event")
				totalBlockEvents++
			}
			for _, event := range results.EndBlockEvents {
				logEntry := logger.With(zap.String("type", event.Type))
				eventsByType["bend:"+event.Type]++
				switch event.Type {
				case "complete_unbonding":
				default:
					log.Fatalf("unknown event type: %q", event.Type)
				}
				for _, v := range event.GetAttributes() {
					key := bytes.NewBuffer(v.GetKey()).String()
					value := bytes.NewBuffer(v.GetValue()).String()
					logEntry = logEntry.With(zap.String(key, value))
				}
				logEntry.Debug("  end block event")
				totalBlockEvents++
			}

			block, err := client.Block(ctx, &height)
			if err != nil {
				// FIXME: retry policy, ignore?
				panic(err)
			}
			for _, tx := range block.Block.Txs {
				if block.Block.Txs != nil {
					logger.Debug("  tx", zap.String("hash", fmt.Sprintf("%x", tx.Hash())))
				}
				res, err := client.Tx(ctx, tx.Hash(), false)
				if err != nil {
					panic(err)
				}

				for _, event := range res.TxResult.Events {
					logEntry := logger.With(zap.String("type", event.Type))
					switch event.Type {
					case "transfer":
					case "message":
					case "unbond":
					case "withdraw_commission":
					case "withdraw_rewards":
					case "delegate":
					default:
						log.Fatalf("unknown event type: %q", event.Type)
					}
					for _, v := range event.GetAttributes() {
						key := bytes.NewBuffer(v.GetKey()).String()
						value := bytes.NewBuffer(v.GetValue()).String()

						logEntry = logEntry.With(zap.String(key, value))
						logEntry.Debug("   tx event")
					}
					totalTxEvents++
					eventsByType["tx:"+event.Type]++
				}

				// fmt.Println("  ", u.PrettyJSON(res))
				totalTxs++
			}
			if !config.debug {
				bar.Add(1)
			}
			totalBlocks++
		}
		logger.Info("finished",
			zap.Int("blocks", totalBlocks),
			zap.Int("block events", totalBlockEvents),
			zap.Int("txs", totalTxs),
			zap.Int("tx events", totalTxEvents),
			zap.Duration("duration", time.Since(start)),
		)
		logEntry := logger
		for key, value := range eventsByType {
			logEntry = logEntry.With(zap.Int(key, value))
		}
		logEntry.Debug("by events")
	}

	return nil
}

func newRPCClient(addr string) (*rpchttp.HTTP, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = 5 * time.Second
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}
