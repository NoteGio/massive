package main

import (
  "context"
  "flag"
  "os"
  "github.com/google/subcommands"
  "github.com/notegio/massive/zeroEx"
)

func main() {
  subcommands.Register(&zeroEx.ZeroExCmd{}, "")

  flag.Parse()
  ctx := context.Background()
  os.Exit(int(subcommands.Execute(ctx)))
}
