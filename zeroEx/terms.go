package zeroEx

import (
	"bytes"
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"github.com/notegio/openrelay/types"
	"github.com/notegio/openrelay/common"
	termsModule "github.com/notegio/openrelay/terms"
	"math/rand"
	"math/big"
	"time"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"net/http"
)

type terms struct {
	inputFileName     string
	outputFileName    string
	keyFile           *os.File
	inputFile         *os.File
	outputFile        *os.File
	targetURL         string
}

func (p *terms) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *terms) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*terms) Name() string     { return "terms" }
func (*terms) Synopsis() string { return "View and sign the Terms of Service" }
func (*terms) Usage() string {
	return `msv 0x terms [--input FILE] [--output FILE] KEYFILE:
  View and sign the terms of
`
}

func (p *terms) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.StringVar(&p.targetURL, "target", "https://api.openrelay.xyz", "Set the target 0x relayer")
}

// CheckMask verifies that a given hash matches the provided hashmask
func CheckMask(mask, hash []byte) bool {
	maskInt := new(big.Int).SetBytes(mask)
	hashInt := new(big.Int).SetBytes(hash)
	return new(big.Int).And(hashInt, maskInt).Cmp(maskInt) == 0
}

func FindValidNonce(text, timestamp string, mask []byte) (<-chan []byte) {
	ch := make(chan []byte)
	go func(text, timestamp string, mask []byte, ch chan []byte) {
		rand.Seed(time.Now().UTC().UnixNano())
		hash := []byte{}
		nonce := make([]byte, 32)
		for !CheckMask(mask, hash) {
			nonce = make([]byte, 32)
			rand.Read(nonce[:])
			termsSha := sha3.NewKeccak256()
			termsSha.Write([]byte(fmt.Sprintf("%v\n%v\n%#x", text, timestamp, nonce)))
			hash = termsSha.Sum(nil)
		}
		ch <- nonce
	}(text, timestamp, mask, ch)
	return ch
}

func (p *terms) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
	return TermsMain(p.inputFile, p.outputFile, privKey, p.targetURL)
}
func TermsMain(inputFile io.Reader, outputFile io.Writer, key *ecdsa.PrivateKey, targetURL string) subcommands.ExitStatus {
	targetURL = strings.TrimSuffix(targetURL, "/")

	resp, err := http.Get(fmt.Sprintf("%v/v2/_tos", targetURL))
	if err != nil {
		log.Printf("Error getting terms from %v: %v", targetURL, err.Error())
		return subcommands.ExitFailure
	}
	if resp.StatusCode != 200 {
		log.Printf("Got unexpected status code: %v", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Body: %v", string(body))
		return subcommands.ExitFailure
	}
	termsPayload := &termsModule.TermsFormat{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, termsPayload); err != nil {
		log.Printf("Error processing payload: %v", err.Error())
		return subcommands.ExitFailure
	}

	fmt.Fprintf(os.Stderr, "%v\nPress Enter to Agree...", termsPayload.Text)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	now := time.Now()
	nonce := <-FindValidNonce(termsPayload.Text, fmt.Sprintf("%v", now.Unix()), termsPayload.Mask[:])


	termsSha := sha3.NewKeccak256()
	termsSha.Write([]byte(fmt.Sprintf("%v\n%v\n%#x", termsPayload.Text, now.Unix(), nonce)))
	hash := termsSha.Sum(nil)

	log.Printf("\nHash: %#x\n Mask: %#x\nNonce: %#x\n", hash, termsPayload.Mask[:], nonce)

	address := common.BytesToOrAddress(crypto.PubkeyToAddress(key.PublicKey))
	hashedBytes := append([]byte("\x19Ethereum Signed Message:\n32"), hash...)
	signedBytes := crypto.Keccak256(hashedBytes)

	sig, _ := crypto.Sign(signedBytes, key)
	Signature := make(types.Signature, 66)
	Signature[0] = sig[64] + 27
	copy(Signature[1:33], sig[0:32])
	copy(Signature[33:65], sig[32:64])
	Signature[65] = types.SigTypeEthSign

	response := &termsModule.TermsSigPayload{
		TermsID: termsPayload.ID,
		MaskID: termsPayload.MaskID,
		Signature: &Signature,
		Address: address,
		Timestamp: fmt.Sprintf("%v", now.Unix()),
		Nonce: fmt.Sprintf("%#x", nonce[:]),
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling signature: %v", err.Error())
		return subcommands.ExitFailure
	}
	resp, err = http.Post(fmt.Sprintf("%v/v2/_tos", targetURL), "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("Error signing terms %v: %v", targetURL, err.Error())
		return subcommands.ExitFailure
	}
	if resp.StatusCode != 202 {
		bodyContent, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error submitting terms: %v", string(bodyContent))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
