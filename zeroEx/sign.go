package zeroEx

import (
  "bytes"
  "flag"
  "context"
  "os"
  "crypto/ecdsa"
  "github.com/google/subcommands"
  "github.com/notegio/massive/utils"
  "github.com/ethereum/go-ethereum/crypto"
  "io"
  "log"
)

type signOrder struct {
  inputFileName string
  outputFileName string
  keyFile *os.File
  inputFile *os.File
  outputFile *os.File
  errOnMismatch bool
}

func (p *signOrder) FileNames() (string, string) {
  return p.inputFileName, p.outputFileName
}

func (p *signOrder) SetIOFiles(inputFile, outputFile *os.File) {
  p.inputFile, p.outputFile = inputFile, outputFile
}

func (*signOrder) Name() string     { return "sign" }
func (*signOrder) Synopsis() string { return "Add a signature to an order" }
func (*signOrder) Usage() string {
  return `msv 0x sign KEYFILE [--input FILE] [--output FILE]:
  Add a salt to the order
`
}

func (p *signOrder) SetFlags(f *flag.FlagSet) {
  f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
  f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
  f.BoolVar(&p.errOnMismatch, "err-on-mismatch", false, "Exit with a non-zero exit code if an order's Maker does not match the provided key")
}

func (p *signOrder) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
  utils.SetIO(p)
  privKey, err := crypto.LoadECDSA(f.Arg(0))
  if err != nil {
    log.Printf("Error loading key: %v", err.Error())
    return subcommands.ExitFailure
  }
  return SignOrderMain(p.inputFile, p.outputFile, privKey, p.errOnMismatch)
}
func SignOrderMain(inputFile io.Reader, outputFile io.Writer, key *ecdsa.PrivateKey, errOnMismatch bool) subcommands.ExitStatus {
  address := crypto.PubkeyToAddress(key.PublicKey)
  for order := range orderScanner(inputFile) {
    if errOnMismatch && !bytes.Equal(address[:], order.Maker[:]) {
      log.Printf("Private key address does not match maker address")
      return subcommands.ExitFailure
    }
    copy(order.Maker[:], address[:])
    copy(order.Signature.Hash[:], order.Hash())

    hashedBytes := append([]byte("\x19Ethereum Signed Message:\n32"), order.Signature.Hash[:]...)
    signedBytes := crypto.Keccak256(hashedBytes)

    sig, _ := crypto.Sign(signedBytes, key)
    copy(order.Signature.R[:], sig[0:32])
    copy(order.Signature.S[:], sig[32:64])
    order.Signature.V = sig[64] + 27
    utils.WriteRecord(order, outputFile)
  }
  return subcommands.ExitSuccess
}
