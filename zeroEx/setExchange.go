package zeroEx

import (
	"context"
	"encoding/hex"
	"flag"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"log"
	"os"
	"strings"
)

type setExchange struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
	mainnet        bool
	ropsten        bool
	kovan          bool
	rinkeby        bool
	testrpc        bool
	address        string
}

func (p *setExchange) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *setExchange) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*setExchange) Name() string { return "setExchange" }
func (*setExchange) Synopsis() string {
	return "Set the 0x Exchange Address on each order for the specified network"
}
func (*setExchange) Usage() string {
	return `msv 0x setExchange [--mainnet | --ropsten | --kovan | --rinkeby | --testrpc | --address=EXCHANGE_ADDRESS][--input FILE] [--output FILE]:
  Set the exchange address address based on the specified network or address
`
}

func (p *setExchange) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.StringVar(&p.address, "address", "", "Use the specified address")
	f.BoolVar(&p.mainnet, "mainnet", false, "Use the exchange address for mainnet")
	f.BoolVar(&p.ropsten, "ropsten", false, "Use the exchange address for ropsten")
	f.BoolVar(&p.kovan, "kovan", false, "Use the exchange address for kovan")
	f.BoolVar(&p.rinkeby, "rinkeby", false, "Use the exchange address for rinkeby")
	f.BoolVar(&p.testrpc, "testrpc", false, "Use the exchange address for testrpc")
}

func (p *setExchange) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	utils.SetIO(p)
	if p.mainnet {
		p.address = "0x12459c951127e0c374ff9105dda097662a027093"
	} else if p.ropsten {
		p.address = "0x479cc461fecd078f766ecc58533d6f69580cf3ac"
	} else if p.kovan {
		p.address = "0x90fe2af704b34e0224bf2299c838e04d4dcf1364"
	} else if p.rinkeby {
		p.address = "0x1d16ef40fac01cec8adac2ac49427b9384192c05"
	} else if p.testrpc {
		p.address = "0x48bacb9266a570d521063ef5dd96e61686dbe788"
	}
	return SetExchangeMain(p.inputFile, p.outputFile, p.address)
}
func SetExchangeMain(inputFile io.Reader, outputFile io.Writer, address string) subcommands.ExitStatus {
	address = strings.TrimPrefix(address, "0x")
	if len(address) != 40 {
		log.Printf("Address should be 40 hex characters, with optional '0x' prefix")
		return subcommands.ExitFailure
	}
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		log.Printf("Error decoding address bytes: %v", err.Error)
		return subcommands.ExitFailure
	}
	for order := range orderScanner(inputFile) {
		copy(order.ExchangeAddress[:], addressBytes)
		utils.WriteRecord(order, outputFile)
	}
	return subcommands.ExitSuccess
}
