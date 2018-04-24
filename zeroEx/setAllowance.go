package zeroEx

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	orCommon "github.com/notegio/openrelay/common"
	"github.com/notegio/openrelay/config"
	"github.com/notegio/openrelay/funds"
	tokenModule "github.com/notegio/openrelay/token"
	"github.com/notegio/openrelay/types"
	"io"
	"log"
	"math/big"
	"os"
	"sync"
)

type setAllowance struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
	unlimited      bool
}

func (p *setAllowance) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *setAllowance) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*setAllowance) Name() string { return "setAllowance" }
func (*setAllowance) Synopsis() string {
	return "Set the 0x Exchange Address on each order for the specified network"
}
func (*setAllowance) Usage() string {
	return `msv 0x setAllowance [ETHEREUM_RPC_URL] [KEY_FILE] [--unlimited] [--input FILE] [--output FILE]:
  Set allowances on the target Ethereum host using the specified key for any
	orders in the input file. Replays the orders on the output file after the
	approvals have been confirmed. Orders may output in a different order than
	input, depending on when the approvals are confirmed.

	If the --unlimited flag is provided and the current allowance is below 2^255,
	allowances will be set to 2^256 - 1. Otherwise allowances will be increased
	by the amount in the order.
`
}

func (p *setAllowance) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
	f.BoolVar(&p.unlimited, "unlimited", false, "Set unlimited allowances for ")
}

func (p *setAllowance) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	utils.SetIO(p)
	conn, err := ethclient.Dial(f.Arg(0))
	if err != nil {
		log.Printf("Error establishing Ethereum connection: %v", err.Error())
		return subcommands.ExitFailure
	}
	tokenProxyCfg, err := config.NewRpcTokenProxy(f.Arg(0))
	if err != nil {
		log.Printf("Error setting up TokenProxy config: %v", err.Error())
		return subcommands.ExitFailure
	}
	feeTokenCfg, err := config.NewRpcFeeToken(f.Arg(0))
	if err != nil {
		log.Printf("Error setting up FeeToken config: %v", err.Error())
		return subcommands.ExitFailure
	}
	privKey, err := crypto.LoadECDSA(f.Arg(1))
	if err != nil {
		log.Printf("Error loading key: %v", err.Error())
		return subcommands.ExitFailure
	}
	balanceChecker, err := funds.NewRpcBalanceChecker(f.Arg(0))
	if err != nil {
		log.Printf("Error initializing balanceChecker: %v", err.Error())
		return subcommands.ExitFailure
	}
	return SetAllowanceMain(p.inputFile, p.outputFile, conn, privKey, p.unlimited, tokenProxyCfg, feeTokenCfg, balanceChecker)
}

func SetAllowanceMain(inputFile io.Reader, outputFile io.Writer, conn *ethclient.Client, key *ecdsa.PrivateKey, unlimited bool, tokenProxyCfg config.TokenProxy, feeTokenCfg config.FeeToken, balanceChecker funds.BalanceChecker) subcommands.ExitStatus {
	if !unlimited {
		log.Printf("Currently only unlimited allowances are supported. Add the '--unlimited' to use this tool.")
		return subcommands.ExitFailure
	}
	allowanceFutures := make(map[types.Address]map[types.Address]*allowanceFuture)
	orderChannel := make(chan *types.Order)
	exitStatusChannel := make(chan subcommands.ExitStatus)
	var wg sync.WaitGroup
	go func() {
		for order := range orderChannel {
			if order == nil {
				exitStatusChannel <- subcommands.ExitFailure
				return
			}
			utils.WriteRecord(order, outputFile)
		}
		exitStatusChannel <- subcommands.ExitSuccess
	}()
	for order := range orderScanner(inputFile) {
		feeTokenAddress, err := feeTokenCfg.Get(order)
		if err != nil {
			log.Printf("Error getting fee token for exchange %v: %v", order.ExchangeAddress, err.Error())
			return subcommands.ExitFailure
		}
		_, ok := allowanceFutures[*order.Maker]
		if !ok {
			allowanceFutures[*order.Maker] = make(map[types.Address]*allowanceFuture)
		}
		_, ok = allowanceFutures[*order.Maker][*order.MakerToken]
		if !ok {
			allowanceFutures[*order.Maker][*order.MakerToken] = &allowanceFuture{
				nil,
				nil,
				make(chan bool),
			}
			go allowanceFutures[*order.Maker][*order.MakerToken].Populate(order.Maker, order.MakerToken, order, balanceChecker, tokenProxyCfg, conn, key)
		}
		_, ok = allowanceFutures[*order.Maker][*feeTokenAddress]
		if !ok {
			allowanceFutures[*order.Maker][*feeTokenAddress] = &allowanceFuture{
				nil,
				nil,
				make(chan bool),
			}
			go allowanceFutures[*order.Maker][*feeTokenAddress].Populate(order.Maker, feeTokenAddress, order, balanceChecker, tokenProxyCfg, conn, key)
		}
		wg.Add(1)
		go func(order *types.Order) {
			defer wg.Done()
			_, err := allowanceFutures[*order.Maker][*order.MakerToken].Get()
			if err != nil {
				orderChannel <- nil
				log.Printf("Error getting / setting allowance %v", err.Error())
			}
			_, err = allowanceFutures[*order.Maker][*feeTokenAddress].Get()
			if err != nil {
				orderChannel <- nil
				log.Printf("Error getting / setting allowance %v", err.Error())
			}
			orderChannel <- order
		}(order)
	}
	go func() {
		wg.Wait()
		close(orderChannel)
	}()
	return <-exitStatusChannel
}

type allowanceFuture struct {
	allowance *big.Int
	err       error
	channel   chan bool
}

func (future *allowanceFuture) Get() (*big.Int, error) {
	for _ = range future.channel {
	}
	return future.allowance, future.err
}

func (future *allowanceFuture) Populate(makerAddress, tokenAddress *types.Address, order *types.Order, balanceChecker funds.BalanceChecker, tokenProxyCfg config.TokenProxy, conn *ethclient.Client, key *ecdsa.PrivateKey) {
	unlimitedAllowance := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
	tokenProxyAddress, err := tokenProxyCfg.Get(order)
	allowance, err := balanceChecker.GetAllowance(tokenAddress, makerAddress, tokenProxyAddress)
	if err != nil {
		future.err = err
		close(future.channel)
		return
	}
	if new(big.Int).Rsh(unlimitedAllowance, 2).Cmp(allowance) > 0 {
		token, err := tokenModule.NewToken(orCommon.ToGethAddress(tokenAddress), conn)
		if err != nil {
			future.err = err
			close(future.channel)
			return
		}
		transactOpt := bind.NewKeyedTransactor(key)
		// TODO: Allow configuration of gas price
		transaction, err := token.TokenTransactor.Approve(transactOpt, orCommon.ToGethAddress(tokenProxyAddress), unlimitedAllowance)
		if err != nil {
			future.err = err
			close(future.channel)
			return
		}
		_, err = bind.WaitMined(context.Background(), conn, transaction)
		if err != nil {
			future.err = err
			close(future.channel)
			return
		}
		future.allowance = unlimitedAllowance
	}
	future.allowance = allowance
	close(future.channel)

}
