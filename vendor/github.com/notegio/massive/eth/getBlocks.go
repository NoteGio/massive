package eth

import (
	"context"
	"flag"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"log"
	"math/big"
	"os"
)

type getBlocks struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
	fromBlock      int
	toBlock        int
}

func (p *getBlocks) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *getBlocks) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*getBlocks) Name() string { return "getBlocks" }
func (*getBlocks) Synopsis() string {
	return "Get Ethereum blocks from an RPC server and pipe them to --output"
}
func (*getBlocks) Usage() string {
	return `msv eth getBlocks [--fromBlock NUM] [--toBlock NUM] [--output FILE] ETHEREUM_RPC_URL:
  Reads blocks from an RPC server and write them to the outputfile
`
}

func (p *getBlocks) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.IntVar(&p.fromBlock, "fromBlock", 0, "The starting block")
	f.IntVar(&p.toBlock, "toBlock", -1, "The ending block")
}

func (p *getBlocks) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	conn, err := ethclient.Dial(f.Arg(0))
	if err != nil {
		log.Printf("Error establishing Ethereum connection: %v", err.Error())
		return subcommands.ExitFailure
	}
	return GetBlocksMain(p.inputFile, p.outputFile, conn, int64(p.fromBlock), int64(p.toBlock))
}

func GetBlocksMain(inputFile io.Reader, outputFile io.Writer, conn *ethclient.Client, fromBlock, toBlock int64) subcommands.ExitStatus {
	for blockNumber := fromBlock; blockNumber < toBlock || toBlock == -1; blockNumber++ {
		header, err := conn.BlockByNumber(context.Background(), big.NewInt(blockNumber))
		if err == ethereum.NotFound && toBlock == -1 {
			return subcommands.ExitSuccess
		} else if err != nil {
			log.Printf("Error getting block %v: %v", blockNumber, err.Error())
			return subcommands.ExitFailure
		}
		blockMap, err := serializeBlock(header, true, true)
		if err != nil {
			log.Printf("Error serializing block %v: %v", blockNumber, err.Error())
			return subcommands.ExitFailure
		}
		err = utils.WriteRecord(blockMap, outputFile)
		if err != nil {
			log.Printf("Error writing block %v: %v", blockNumber, err.Error())
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}
