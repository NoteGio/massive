package zeroEx

import (
  "flag"
  "context"
  "os"
  "github.com/google/subcommands"
  "github.com/notegio/massive/utils"
  "github.com/ethereum/go-ethereum/accounts/abi"
  "math/big"
  "io"
  "time"
  "log"
)

type expiration struct {
  duration bool
  inputFileName string
  outputFileName string
  inputFile *os.File
  outputFile *os.File
}

func (p *expiration) FileNames() (string, string) {
  return p.inputFileName, p.outputFileName
}

func (p *expiration) SetIOFiles(inputFile, outputFile *os.File) {
  p.inputFile, p.outputFile = inputFile, outputFile
}

func (*expiration) Name() string     { return "expiration" }
func (*expiration) Synopsis() string { return "Set the expiration timestamp" }
func (*expiration) Usage() string {
  return `msv 0x expiration [--duration] [--input FILE] [--output FILE] TIME :
  Set the expiration timestamp. TIME is a unix timestamp, unless --duration is
  specified, in which case it will be the number of seconds from now that the
  order will be valid.
`
}

func (p *expiration) SetFlags(f *flag.FlagSet) {
  f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
  f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
  f.BoolVar(&p.duration, "duration", false, "Specify the lifespan of the order, instead of the expiration time")
}

func (p *expiration) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
  utils.SetIO(p)
  value, ok := new(big.Int).SetString(f.Arg(0), 10)
  if !ok {
    log.Printf("Error processing argument: %v\n", f.Arg(0))
    return subcommands.ExitFailure
  }
  return SetExpirationMain(p.inputFile, p.outputFile, p.duration, value)
}
func SetExpirationMain(inputFile io.Reader, outputFile io.Writer, duration bool, value *big.Int) subcommands.ExitStatus {
  for order := range orderScanner(inputFile) {
    if duration {
      copy(order.ExpirationTimestampInSec[:], abi.U256(new(big.Int).Add(value, big.NewInt(time.Now().Unix()))))
    } else {
      copy(order.ExpirationTimestampInSec[:], abi.U256(value))
    }
    utils.WriteRecord(order, outputFile)
  }
  return subcommands.ExitSuccess
}
