package zeroEx

import (
  "context"
  "flag"
  "path"
  "os"
  "fmt"
  "github.com/google/subcommands"
)

type ZeroExCmd struct {}

func (*ZeroExCmd) Name() string     { return "0x" }
func (*ZeroExCmd) Synopsis() string { return "Subcommands relating to 0x" }
func (*ZeroExCmd) Usage() string {
  return `msv 0x [subcommand]:
  Call 0x subcommands
`
}

func (p *ZeroExCmd) SetFlags(f *flag.FlagSet) {}

func (p *ZeroExCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
  commander := subcommands.NewCommander(f, fmt.Sprintf("%v %v", path.Base(os.Args[0]), os.Args[1]))
  commander.Register(&getFees{}, "")
  commander.Register(&setSalt{}, "")
  commander.Register(&expiration{}, "")
  commander.Register(&signOrder{}, "")
  commander.Register(&upload{}, "")
  commander.Register(&csvReader{}, "")
  commander.Register(&setExchange{}, "")
  commander.Register(commander.HelpCommand(), "")
  commander.Register(commander.FlagsCommand(), "")
  commander.Register(commander.CommandsCommand(), "")
  ctx := context.Background()
  return commander.Execute(ctx)
}
