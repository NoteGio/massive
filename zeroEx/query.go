package zeroEx

import (
	"os"
	"flag"
	"context"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"strings"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type queryOrders struct {
	targetURL      string
	inputFileName  string
	outputFileName string
	makerToken     string
	takerToken		 string
	exchange        string
	inputFile      *os.File
	outputFile     *os.File
}

type ecSignature struct {
	V    	int64 `json:"v"`
	R			string `json:"r"`
	S    	string `json:"s"`
}

type OrderResponse struct {
	Exchange       		string `json:"exchangeContractAddress"`
	Maker					 		string `json:"maker"`
	Taker					 		string `json:"taker"`
	MakerToken	   		string `json:"makerTokenAddress"`
	TakerToken	   		string `json:"takerTokenAddress"`
	FeeRecipient   		string `json:"feeRecipient"`
	MakerTokenAmount 	string `json:"makerTokenAmount"`
	TakerTokenAmount 	string `json:"takerTokenAmount"`
	MakerFee       		string `json:"makerFee"`
	TakerFee       		string `json:"takerFee"`
	Expiration		 		string `json:"expirationUnixTimestampSec"`
	Salt		       		string `json:"salt"`
	Signature					ecSignature `json:"ecSignature"`
}

type Orders struct {
	orders []OrderResponse
}



func (p *queryOrders) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *queryOrders) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*queryOrders) Name() string     { return "query" }
func (*queryOrders) Synopsis() string { return "Query the relayer api for orders" }
func (*queryOrders) Usage() string {
	return `msv 0x query [--target RELAYER_URL] [--maker-token  0x..] [--taker-token 0x..] [--exchange testrpc] [--input FILE] [--output FILE]:
  Get fees from the target relayer and set them on the order
`
}

func (p *queryOrders) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.targetURL, "target", "https://api.openrelay.xyz", "Set the target 0x relayer")
	f.StringVar(&p.makerToken, "maker-token", "0x871dd7c2b4b25e1aa18728e9d5f2af4c4e431f5c", "The maker token contract address")
	f.StringVar(&p.takerToken, "taker-token", "0x1d7022f5b17d2f8b695918fb48fa1089c9f85401", "The taker token contract address")
	f.StringVar(&p.exchange, "exchange", "0x48bacb9266a570d521063ef5dd96e61686dbe788", "The exchange contract address")
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
}


func (p *queryOrders) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 0 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	return QueryMain(p.targetURL, p.inputFile, p.outputFile, p.makerToken, p.takerToken, p.exchange)
}

func QueryMain(targetURL string, inputFile io.Reader, outputFile io.Writer, makerToken string, takerToken string, exchange string) subcommands.ExitStatus {

	targetURL = strings.TrimSuffix(targetURL, "/")
	resp, err := http.Get(fmt.Sprintf("%v/v0/orders?makerTokenAddress=%v&takerTokenAddress=%v&exchangeContractAddress=%v", targetURL, makerToken, takerToken, exchange))
	if err != nil {
		log.Printf("Getting orders from %v: %v", targetURL, err.Error())
		return subcommands.ExitFailure
	}
	orderBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error getting response body: %v", err.Error())
		return subcommands.ExitFailure
	}
	if resp.StatusCode != 200 {
		log.Printf("Got unexpected status code: %v", resp.StatusCode)
		log.Printf("Body: %v", string(orderBytes))
		return subcommands.ExitFailure
	}
	orders := make([]OrderResponse,0)
	if err := json.Unmarshal(orderBytes, &orders); err != nil {
		log.Printf("Error parsing response body: %v - '%v'", err.Error(), string(orderBytes))
		return subcommands.ExitFailure
	}
	for _,order := range orders {
		utils.WriteRecord(order, outputFile)
	}
	return subcommands.ExitSuccess
}
