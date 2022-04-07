package main

import (
	"context"
	"fmt"
	"log"
	"time"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"moul.io/godev"
)

func main() {
	//ctx := client.GetClientContextFromCmd(nil)
	//fmt.Println(ctx)

	if err := run(); err != nil {
		log.Fatalf("error: %+v\n", err)
	}
}

func run() error {
	ctx := context.Background()

	client, err := NewRPCClient("http://localhost:26657")
	if err != nil {
		return err
	}

	stat, err := client.Status(ctx)
	if err != nil {
		return err
	}
	fmt.Println(godev.PrettyJSON(stat))
	// curl "http://localhost:26657/block?height=5222672"

	return nil
}

func NewRPCClient(addr string) (*rpchttp.HTTP, error) {
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
