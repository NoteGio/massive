package zeroEx

import (
  "bytes"
  "encoding/json"
  "encoding/hex"
  "flag"
  "context"
  "os"
  "github.com/google/subcommands"
  "github.com/notegio/massive/utils"
  "github.com/notegio/openrelay/ingest"
  "github.com/notegio/openrelay/types"
  "github.com/ethereum/go-ethereum/accounts/abi"
  "math/big"
  "log"
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
  "io"
)

type getFees struct {
  targetURL     string
  inputFileName string
  outputFileName string
  makerShare float64
  inputFile *os.File
  outputFile *os.File
}

func (p *getFees) FileNames() (string, string) {
  return p.inputFileName, p.outputFileName
}

func (p *getFees) SetIOFiles(inputFile, outputFile *os.File) {
  p.inputFile, p.outputFile = inputFile, outputFile
}

func (*getFees) Name() string     { return "getFees" }
func (*getFees) Synopsis() string { return "Set fees on incoming orders" }
func (*getFees) Usage() string {
  return `msv 0x getFees [--target RELAYER_URL] [--maker-share 1] [--input FILE] [--output FILE]:
  Call 0x subcommands
`
}

func (p *getFees) SetFlags(f *flag.FlagSet) {
  f.StringVar(&p.targetURL, "target", "https://api.openrelay.xyz", "Set the target 0x relayer")
  f.Float64Var(&p.makerShare, "maker-share", -1, "What share of fees")
  f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
  f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
}

func (p *getFees) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
  utils.SetIO(p)
  return GetFeesMain(p.targetURL, p.inputFile, p.outputFile, p.makerShare)
}

func GetFeesMain(targetURL string, inputFile io.Reader, outputFile io.Writer, makerShare float64) subcommands.ExitStatus {
  targetURL = strings.TrimSuffix(targetURL, "/")
  for order := range orderScanner(inputFile) {
    feeInput := &ingest.FeeInputPayload{}
    feeInput.Maker = order.Maker.String()
    emptyAddress := &types.Address{}
    if bytes.Equal(order.FeeRecipient[:], emptyAddress[:]) {
      feeInput.FeeRecipient = ""
    } else {
      feeInput.FeeRecipient = order.FeeRecipient.String()
    }
    data, err := json.Marshal(feeInput)
    if err != nil {
      log.Printf("Error serializing order: %v", err.Error())
      return subcommands.ExitFailure
    }
    resp, err := http.Post(fmt.Sprintf("%v/v0/fees", targetURL), "application/json", bytes.NewReader(data))
    if err != nil {
      log.Printf("Getting fees from %v: %v", targetURL, err.Error())
      return subcommands.ExitFailure
    }
    feeBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
      log.Printf("Error getting response body: %v", err.Error())
      return subcommands.ExitFailure
    }
    if resp.StatusCode != 200 {
      log.Printf("Got unexpected status code: %v", resp.StatusCode)
      log.Printf("Body: %v", string(feeBytes))
      return subcommands.ExitFailure
    }
    fees := &ingest.FeeResponse{}
    if err := json.Unmarshal(feeBytes, fees); err != nil {
      log.Printf("Error parsing response body: %v - '%v'", err.Error(), string(feeBytes))
      return subcommands.ExitFailure
    }
    makerFee, ok := new(big.Int).SetString(fees.MakerFee, 10)
    if !ok {
      log.Printf("MakerFee not a valid integer: '%v'", fees.MakerFee)
      return subcommands.ExitFailure
    }
    takerFee, ok := new(big.Int).SetString(fees.TakerFee, 10)
    if !ok {
      log.Printf("TakerFee not a valid integer: '%v'", fees.TakerFee)
      return subcommands.ExitFailure
    }
    if makerShare >= 0 && makerShare < 1 {
      totalFee := new(big.Int).Add(makerFee, takerFee)
      makerPercent := big.NewInt(int64(makerShare * 100))
      makerFee = new(big.Int).Div(new(big.Int).Mul(totalFee, makerPercent), big.NewInt(100))
      takerFee = new(big.Int).Sub(totalFee, makerFee)
    }
    copy(order.MakerFee[:], abi.U256(makerFee))
    copy(order.TakerFee[:], abi.U256(takerFee))
    _, err = hex.Decode(order.FeeRecipient[:], []byte(fees.FeeRecipient)[2:])
    if err != nil {
      log.Printf("FeeRecipient not a valid hex string: %v", err.Error())
      return subcommands.ExitFailure
    }
    _, err = hex.Decode(order.Taker[:], []byte(fees.TakerToSpecify)[2:])
    if err != nil {
      log.Printf("FeeRecipient not a valid hex string: %v", err.Error())
      return subcommands.ExitFailure
    }

    utils.WriteRecord(order, outputFile)
  }
  return subcommands.ExitSuccess
}
