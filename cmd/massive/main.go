package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"github.com/notegio/massive/zeroEx"
	"github.com/notegio/massive/eth"
	"os"
)

func main() {
	subcommands.Register(&zeroEx.ZeroExCmd{}, "")
	subcommands.Register(&eth.EthCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
