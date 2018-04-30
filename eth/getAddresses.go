package eth

import (
	"context"
	"flag"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"os"
	"fmt"
	"log"
)

type getAddresses struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
	noDuplicates   bool
}

func (p *getAddresses) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *getAddresses) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*getAddresses) Name() string { return "getAddresses" }
func (*getAddresses) Synopsis() string {
	return "Find all addresses in transactions and blocks"
}
func (*getAddresses) Usage() string {
	return `msv eth getAddresses [--input FILE] [--output FILE] [--no-duplicates]:
  Read through provided blocks, writing out any addresses in transactions or
	block coinbases.
`
}

func (p *getAddresses) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.BoolVar(&p.noDuplicates, "no-duplicates", false, "Exclude duplicate address")
}

func (p *getAddresses) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 0 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	return GetAddressesMain(p.inputFile, p.outputFile, p.noDuplicates)
}

func GetAddressesMain(inputFile io.Reader, outputFile io.Writer, noDuplicates bool) subcommands.ExitStatus {
	knownAddresses := make(map[common.Address]struct{})
	outputAddress := func (address common.Address) {
		if noDuplicates {
			if _, ok := knownAddresses[address]; ok {
				return
			}
			knownAddresses[address] = struct{}{}
		}
		io.WriteString(outputFile, fmt.Sprintf("%v\n", address.String()))
	}
	for block := range blockScanner(inputFile) {
		outputAddress(block.Coinbase)
		for _, tx := range block.Transactions {
			outputAddress(tx.From)
			if tx.To != nil {
				outputAddress(*tx.To)
			}
		}
	}
	return subcommands.ExitSuccess
}
