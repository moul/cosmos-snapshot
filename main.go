package main

import (
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
	"moul.io/godev"
)

// TODO:
// custom handler -> pseudo-lambda
// helpers
//   whale-cap
//   min-investp
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

var config struct {
	minHeight int64
	maxHeight int64
	debug     bool
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %+v\n", err)
	}
}

func run() error {
	var (
		ctx = context.Background()
	)

	// parse CLI flags.
	{
		fs := flag.NewFlagSet("gno-bounty-7", flag.ContinueOnError)
		fs.Int64Var(&config.minHeight, "min-height", 5200791, "first block to process")
		fs.Int64Var(&config.maxHeight, "max-height", 5797010, "last block to process")
		fs.BoolVar(&config.debug, "debug", false, "verbose output")
		err := ff.Parse(fs, os.Args[1:])
		if err != nil {
			return fmt.Errorf("flag parse error: %w", err)
		}
	}

	// init client
	var client *http.HTTP
	{

		var err error
		client, err = newRPCClient("http://localhost:26657")
		if err != nil {
			return fmt.Errorf("new RPC client: %w", err)
		}
	}

	// check status
	{
		status, err := client.Status(ctx)
		if err != nil {
			return fmt.Errorf("get RPC Status: %w", err)
		}
		if config.debug {
			fmt.Println(godev.PrettyJSON(status))
		}
		// FIXME: perform checks + actionable error message
	}

	// iterate over blocks
	{
		bar := progressbar.Default(config.maxHeight - config.minHeight)
		for height := config.minHeight; height <= config.maxHeight; height++ {
			block, err := client.Block(ctx, &height)
			if config.debug {
				fmt.Println(block, err)
				fmt.Println(height)
			}
			bar.Add(1)
		}
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
