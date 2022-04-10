package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"moul.io/cosmos-snapshot/pkg/chainwalker"
	"moul.io/zapconfig"
)

func main() {
	err := run()
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		log.Fatalf("error: %+v\n", err)
	}
}

var flagOpts struct {
	chainwalker.NewRPCWalkerOpts

	debug bool
}

func run() error {
	// parse CLI flags.
	{
		fs := flag.NewFlagSet("cosmos-snapshot", flag.ContinueOnError)
		fs.Int64Var(&flagOpts.MinHeight, "min-height", 5200791, "first block to process")
		fs.Int64Var(&flagOpts.MaxHeight, "max-height", 5797010, "last block to process")
		fs.StringVar(&flagOpts.RPCAddr, "rpc-addr", "http://localhost:26657", "Cosmos RPC Address")
		fs.BoolVar(&flagOpts.debug, "debug", false, "verbose output")
		fs.BoolVar(&flagOpts.WithBlockResults, "with-block-results", false, "query block results")
		fs.BoolVar(&flagOpts.WithoutBlockTxs, "without-block-txs", false, "don't query block Txs")

		err := ff.Parse(fs, os.Args[1:])
		if err != nil {
			return fmt.Errorf("flag parse error: %w", err)
		}
	}

	// init logger
	var logger *zap.Logger
	{
		zapconf := zapconfig.New().EnableStacktrace().SetPreset("light-console")
		if flagOpts.debug {
			zapconf.SetLevel(zapcore.DebugLevel)
		} else {
			zapconf.SetLevel(zapcore.InfoLevel)
		}
		logger = zapconf.MustBuild()
		logger = logger.WithOptions(zap.WithCaller(false))
		logger.Debug("starting")
	}

	// init accountant/rules
	// init the accountant engine (example)
	accountant := &Accountant{
		Logger: logger,
	}
	accountant.init()

	// init chainwalker
	walker, err := chainwalker.NewRPCWalker(chainwalker.NewRPCWalkerOpts{
		RPCAddr:          flagOpts.RPCAddr,
		MinHeight:        flagOpts.MinHeight,
		MaxHeight:        flagOpts.MaxHeight,
		WithBlockResults: flagOpts.WithBlockResults,
		WithoutBlockTxs:  flagOpts.WithoutBlockTxs,
		Logger:           logger,
		Ctx:              context.Background(),
	})
	if err != nil {
		return fmt.Errorf("init walker: %w", err)
	}

	// run the walk
	err = walker.Run(accountant.callback)
	if err != nil {
		return fmt.Errorf("walk the chain: %w", err)
	}

	// display results
	accountant.printResults()
	return nil
}
