package zeroEx

import (
	"context"
	"encoding/hex"
	"flag"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	orCommon "github.com/notegio/openrelay/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/notegio/openrelay/exchangecontract"
)

type fillOrder struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
}

func (p *fillOrder) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *fillOrder) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*fillOrder) Name() string { return "fill" }
func (*fillOrder) Synopsis() string {
	return "Fills the given order"
}
func (*fillOrder) Usage() string {
	return `msv 0x fill [--input FILE] [--output FILE] ETHEREUM_RPC_URL KEY_FILE AMOUNT_TO_FILL:`
}

func (p *fillOrder) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
}

func (p *fillOrder) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 3 {
		os.Stderr.WriteString(p.Usage())
		return subcommands.ExitUsageError
	}
	utils.SetIO(p)
	conn, err := ethclient.Dial(f.Arg(0))
	if err != nil {
		log.Printf("Error establishing Ethereum connection: %v", err.Error())
		return subcommands.ExitFailure
	}
	privKey, err := crypto.LoadECDSA(f.Arg(1))
	if err != nil {
		log.Printf("Error loading key: %v", err.Error())
		return subcommands.ExitFailure
	}
	takerValue, ok := new(big.Int).SetString(f.Arg(2), 10)
	if !ok {
		log.Printf("Error processing argument: %v\n", f.Arg(2))
		return subcommands.ExitFailure
	}

	for order := range orderScanner(p.inputFile) {

		exchange, err := exchangecontract.NewExchange(orCommon.ToGethAddress(order.ExchangeAddress), conn)
		if err != nil {
			log.Printf("Error intializing exchange contract '%v': '%v'", hex.EncodeToString(order.ExchangeAddress[:]), err.Error())
			return subcommands.ExitFailure
		}
		takerAddress := crypto.PubkeyToAddress(privKey.PublicKey)

		// Set the taker address
		transactOpt := bind.NewKeyedTransactor(privKey)
		transactOpt.From = takerAddress

		// Set the order Addresses
		orderAddresses := [5]common.Address{
			orCommon.ToGethAddress(order.Maker),
			orCommon.ToGethAddress(order.Taker),
			orCommon.ToGethAddress(order.MakerToken),
			orCommon.ToGethAddress(order.TakerToken),
			orCommon.ToGethAddress(order.FeeRecipient)}

		// Set the order values
		makerTokenAmount, _ := new(big.Int).SetString(order.MakerTokenAmount.String(), 10)
		takerTokenAmount, _ := new(big.Int).SetString(order.TakerTokenAmount.String(), 10)
		makerFee, _ := new(big.Int).SetString(order.MakerFee.String(), 10)
		takerFee, _ := new(big.Int).SetString(order.TakerFee.String(), 10)
		expirationTime, _ := new(big.Int).SetString(order.ExpirationTimestampInSec.String(), 10)
		salt, _ := new(big.Int).SetString(order.Salt.String(), 10)
		orderValues := [6]*big.Int{
			makerTokenAmount,
			takerTokenAmount,
			makerFee,
			takerFee,
			expirationTime,
			salt}

		// Fill the order
		fillTransaction, err := exchange.FillOrder(
			transactOpt,
			orderAddresses,
			orderValues,
			takerValue,
			true,
			order.Signature.V,
			order.Signature.R,
			order.Signature.S)
		if err != nil {
			log.Printf(err.Error())
			return subcommands.ExitFailure
		}
		_, err = bind.WaitMined(context.Background(), conn, fillTransaction)
		if err != nil {
			log.Printf(err.Error())
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}
