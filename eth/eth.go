package eth

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"os"
	"path"
)

type EthCmd struct{}

func (*EthCmd) Name() string     { return "eth" }
func (*EthCmd) Synopsis() string { return "Subcommands relating to Ethereum" }
func (*EthCmd) Usage() string {
	return `msv eth [subcommand]:
  Call eth subcommands
`
}

func (p *EthCmd) SetFlags(f *flag.FlagSet) {}

func (p *EthCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	commander := subcommands.NewCommander(f, fmt.Sprintf("%v %v", path.Base(os.Args[0]), os.Args[1]))
	commander.Register(&getBlocks{}, "")
	commander.Register(commander.HelpCommand(), "")
	commander.Register(commander.FlagsCommand(), "")
	commander.Register(commander.CommandsCommand(), "")
	ctx := context.Background()
	return commander.Execute(ctx)
}
