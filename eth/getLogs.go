package eth

import (
	"context"
	"flag"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"math/big"
	"io"
	"log"
	"os"
)

type getLogs struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
	fromBlock      int
	toBlock        int
	topics         []string
}

func (p *getLogs) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *getLogs) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*getLogs) Name() string { return "getLogs" }
func (*getLogs) Synopsis() string {
	return "Get Ethereum logs from an RPC server and pipe them to --output"
}
func (*getLogs) Usage() string {
	return `msv eth getLogs [--fromBlock NUM] [--toBlock NUM] [--output FILE] ETHEREUM_RPC_URL:
  Reads blocks from an RPC server and write them to the outputfile
`
}

func (p *getLogs) SetFlags(f *flag.FlagSet) {
	p.topics = make([]string, 5)
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.IntVar(&p.fromBlock, "fromBlock", 0, "The starting block")
	f.IntVar(&p.toBlock, "toBlock", -1, "The ending block")
	f.StringVar(&p.topics[0], "topic0", "", "topic0")
	f.StringVar(&p.topics[1], "topic1", "", "topic1")
	f.StringVar(&p.topics[2], "topic2", "", "topic2")
	f.StringVar(&p.topics[3], "topic3", "", "topic3")
	f.StringVar(&p.topics[4], "topic4", "", "topic4")
}

func (p *getLogs) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
	return GetLogsMain(p.inputFile, p.outputFile, conn, int64(p.fromBlock), int64(p.toBlock), p.topics)
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func stringToHash(hexString string) ([]common.Hash) {
	if hexString == "" {
		return nil
	} else {
		return []common.Hash{common.HexToHash(hexString)}
	}
}

func GetLogsMain(inputFile io.Reader, outputFile io.Writer, conn *ethclient.Client, fromBlock, toBlock int64, topics []string) subcommands.ExitStatus {
	for i := fromBlock; i < toBlock; i = min(i+1000, toBlock) {
		topicHashes := [][]common.Hash{}
		topicBuffer := [][]common.Hash{}
		for _, topic := range topics {
			topicBuffer = append(topicBuffer, stringToHash(topic))
			if topic != "" {
				topicHashes = append(topicHashes, topicBuffer...)
				topicBuffer = [][]common.Hash{}
			}
		}
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(i),
			ToBlock: big.NewInt(min(i+1000, toBlock)),
			Addresses: nil,
			Topics: topicHashes,
		}
		logs, err := conn.FilterLogs(context.Background(), query)
		if err != nil {
			log.Printf("Error filtering: %v", err.Error())
			return subcommands.ExitFailure
		}
		for _, itemLog := range logs {
			if err := utils.WriteRecord(itemLog, outputFile); err != nil {
				log.Printf("Error writing record: %v", err.Error())
				return subcommands.ExitFailure
			}
		}
	}
	return subcommands.ExitSuccess
}
