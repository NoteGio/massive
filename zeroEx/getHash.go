package zeroEx

import (
	"encoding/hex"
	"context"
	"flag"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"os"
)

type getHash struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
}

func (p *getHash) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *getHash) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*getHash) Name() string     { return "getHash" }
func (*getHash) Synopsis() string { return "Set the nonce on orders according to the current timestamp" }
func (*getHash) Usage() string {
	return `msv 0x getHash [--input FILE] [--output FILE]:
  Output the hash of an order
`
}

func (p *getHash) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
}

func (p *getHash) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 0 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	return GetHashMain(p.inputFile, p.outputFile)
}
func GetHashMain(inputFile io.Reader, outputFile io.Writer) subcommands.ExitStatus {
	for order := range orderScanner(inputFile) {
		utils.WriteRecord(hex.EncodeToString(order.Hash()), outputFile)
	}
	return subcommands.ExitSuccess
}
