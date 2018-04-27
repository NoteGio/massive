package zeroEx

import (
	"context"
	"crypto/rand"
	"flag"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"math/big"
	"os"
	"time"
)

type setSalt struct {
	random         bool
	value          int64
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
}

func (p *setSalt) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *setSalt) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*setSalt) Name() string     { return "setSalt" }
func (*setSalt) Synopsis() string { return "Set the nonce on orders according to the current timestamp" }
func (*setSalt) Usage() string {
	return `msv 0x setSalt [--random] [--value=INT] [--input FILE] [--output FILE]:
  Add a salt to the order
`
}

func (p *setSalt) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.BoolVar(&p.random, "random", false, "Use a random salt instead of a time based salt")
	f.Int64Var(&p.value, "value", -1, "A specific value to set the salt to")
}

func (p *setSalt) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 0 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	return SetSaltMain(p.inputFile, p.outputFile, p.random, p.value)
}
func SetSaltMain(inputFile io.Reader, outputFile io.Writer, random bool, value int64) subcommands.ExitStatus {
	for order := range orderScanner(inputFile) {
		if value >= 0 {
			copy(order.Salt[:], abi.U256(big.NewInt(value)))
		} else if !random {
			copy(order.Salt[:], abi.U256(big.NewInt(time.Now().Unix())))
		} else {
			rand.Read(order.Salt[:])
		}
		utils.WriteRecord(order, outputFile)
	}
	return subcommands.ExitSuccess
}
