package zeroEx

import (
  "bytes"
  "encoding/json"
  "flag"
  "context"
  "os"
  "github.com/google/subcommands"
  "github.com/notegio/massive/utils"
  "log"
  "fmt"
  "strings"
  "net/http"
  "io"
  "io/ioutil"
)

type upload struct {
  targetURL     string
  inputFileName string
  outputFileName string
  inputFile *os.File
  outputFile *os.File
}

func (p *upload) FileNames() (string, string) {
  return p.inputFileName, p.outputFileName
}

func (p *upload) SetIOFiles(inputFile, outputFile *os.File) {
  p.inputFile, p.outputFile = inputFile, outputFile
}

func (*upload) Name() string     { return "upload" }
func (*upload) Synopsis() string { return "Set fees on incoming orders" }
func (*upload) Usage() string {
  return `msv 0x upload [--target RELAYER_URL] [--input FILE] [--output FILE]:
  Upload orders to the target relayer
`
}

func (p *upload) SetFlags(f *flag.FlagSet) {
  f.StringVar(&p.targetURL, "target", "https://api.openrelay.xyz", "Set the target 0x relayer")
  f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
  f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
}

func (p *upload) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
  utils.SetIO(p)
  return UploadMain(p.targetURL, p.inputFile, p.outputFile)
}

func UploadMain(targetURL string, inputFile io.Reader, outputFile io.Writer) subcommands.ExitStatus {
  targetURL = strings.TrimSuffix(targetURL, "/")
  for order := range orderScanner(inputFile) {
    data, err := json.Marshal(order)
    if err != nil {
      log.Printf("Error serializing order: %v", err.Error())
      return subcommands.ExitFailure
    }
    resp, err := http.Post(fmt.Sprintf("%v/v0/order", targetURL), "application/json", bytes.NewReader(data))
    if err != nil {
      log.Printf("Error uploading order to %v: %v", targetURL, err.Error())
      return subcommands.ExitFailure
    }
    if resp.StatusCode != 202 && resp.StatusCode != 200 {
      log.Printf("Got unexpected status code: %v", resp.StatusCode)
      body, _ := ioutil.ReadAll(resp.Body)
      log.Printf("Body: %v", string(body))
      return subcommands.ExitFailure
    }
  }
  return subcommands.ExitSuccess
}
