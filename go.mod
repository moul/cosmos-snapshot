module moul.io/gno-bounty-7

go 1.16

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/gogo/protobuf v1.3.3 // indirect
	github.com/peterbourgon/ff v1.7.1
	github.com/tendermint/tendermint v0.34.16
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb // indirect
	moul.io/godev v1.7.0
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2

)
