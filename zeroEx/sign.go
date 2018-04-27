package zeroEx

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"flag"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"io"
	"log"
	"os"
)

type signOrder struct {
	inputFileName     string
	outputFileName    string
	keyFile           *os.File
	inputFile         *os.File
	outputFile        *os.File
	errOnMismatch     bool
	replaceOnMismatch bool
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
	return `msv 0x sign [--input FILE] [--output FILE] KEYFILE:
  Sign the 0x order
`
}

func (p *signOrder) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.BoolVar(&p.errOnMismatch, "err-on-mismatch", false, "Exit with a non-zero exit code if an order's Maker does not match the provided key")
	f.BoolVar(&p.replaceOnMismatch, "replace-on-mismatch", false, "Replace the maker address if an order's maker does not match the provided key. If the maker does not match the provided key and this flag is not set, the order will pass through unsigned.")
}

func (p *signOrder) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	privKey, err := crypto.LoadECDSA(f.Arg(0))
	if err != nil {
		log.Printf("Error loading key: %v", err.Error())
		return subcommands.ExitFailure
	}
	return SignOrderMain(p.inputFile, p.outputFile, privKey, p.errOnMismatch, p.replaceOnMismatch)
}
func SignOrderMain(inputFile io.Reader, outputFile io.Writer, key *ecdsa.PrivateKey, errOnMismatch, replaceOnMismatch bool) subcommands.ExitStatus {
	if errOnMismatch && replaceOnMismatch {
		log.Printf("Specify at most one of --err-on-mismatch or --replace-on-mismatch")
	}
	address := crypto.PubkeyToAddress(key.PublicKey)
	for order := range orderScanner(inputFile) {
		if !bytes.Equal(address[:], order.Maker[:]) {
			if errOnMismatch {
				log.Printf("Private key address does not match maker address")
				return subcommands.ExitFailure
			}
			if !replaceOnMismatch {
				utils.WriteRecord(order, outputFile)
				continue
			}
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
