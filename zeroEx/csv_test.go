package zeroEx_test

import (
  "encoding/json"
	"math/big"
  "github.com/notegio/massive/zeroEx"
  "github.com/notegio/openrelay/types"
  "github.com/google/subcommands"
  "testing"
  "bufio"
  "bytes"
)

func TestCsvProcess(t *testing.T) {
	csvData := `makerTokenAddress,makerTokenAmount,takerTokenAddress,takerTokenAmount,exchangeContractAddress
0xa1df88ea6a08722055250ed65601872e59cddfaa,1000000000000000000,0xc778417e063141139fce010982780140aa0cd5ab,1000000000000000000,0x479cc461fecd078f766ecc58533d6f69580cf3ac
`
	inputFile := bytes.NewReader([]byte(csvData))
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.CSVMain(inputFile, outputFile); status != subcommands.ExitSuccess {
    t.Fatalf("Bad exitcode: %v", status)
  }
  outputFile.Flush()
  processedOrder := &types.Order{}
  err := json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
    t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
  }
	if processedOrder.MakerToken.String() != "0xa1df88ea6a08722055250ed65601872e59cddfaa" {
		t.Errorf("Unexpected MakerToken value: %v", processedOrder.MakerToken)
	}
	if processedOrder.TakerToken.String() != "0xc778417e063141139fce010982780140aa0cd5ab" {
		t.Errorf("Unexpected MakerToken value: %v", processedOrder.TakerToken)
	}
	if processedOrder.ExchangeAddress.String() != "0x479cc461fecd078f766ecc58533d6f69580cf3ac" {
		t.Errorf("Unexpected ExchangeAddress value: %v", processedOrder.ExchangeAddress)
	}
	if makerTokenAmount := new(big.Int).SetBytes(processedOrder.MakerTokenAmount[:]); makerTokenAmount.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Errorf("Unexpected makerTokenAmount: %v", makerTokenAmount)
	}
	if takerTokenAmount := new(big.Int).SetBytes(processedOrder.TakerTokenAmount[:]); takerTokenAmount.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Errorf("Unexpected takerTokenAmount: %v", takerTokenAmount)
	}
}
func TestCsvAllFieldsProcess(t *testing.T) {
	csvData := `maker,makerTokenAddress,makerTokenAmount,makerFee,taker,takerTokenAddress,takerTokenAmount,takerFee,expirationUnixTimestampSec,feeRecipient,salt,exchangeContractAddress,ecSignature.v,ecSignature.r,ecSignature.s
0x324454186bb728a3ea55750e0618ff1b18ce6cf8,0xa1df88ea6a08722055250ed65601872e59cddfaa,1000000000000000000,0,,0xc778417e063141139fce010982780140aa0cd5ab,1000000000000000000,0,1502841540,0x0000000000000000000000000000000000000000,11065671350908846865864045738088581419204014210814002044381812654087807531,0x479cc461fecd078f766ecc58533d6f69580cf3ac,27,0x021fe6dba378a347ea5c581adcd0e0e454e9245703d197075f5d037d0935ac2e,0x12ac107cb04be663f542394832bbcb348deda8b5aa393a97a4cc3139501007f1
`
	inputFile := bytes.NewReader([]byte(csvData))
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.CSVMain(inputFile, outputFile); status != subcommands.ExitSuccess {
    t.Fatalf("Bad exitcode: %v", status)
  }
  outputFile.Flush()
  processedOrder := &types.Order{}
  err := json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
    t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
  }
	if processedOrder.MakerToken.String() != "0xa1df88ea6a08722055250ed65601872e59cddfaa" {
		t.Errorf("Unexpected MakerToken value: %v", processedOrder.MakerToken)
	}
	if processedOrder.TakerToken.String() != "0xc778417e063141139fce010982780140aa0cd5ab" {
		t.Errorf("Unexpected MakerToken value: %v", processedOrder.TakerToken)
	}
	if processedOrder.ExchangeAddress.String() != "0x479cc461fecd078f766ecc58533d6f69580cf3ac" {
		t.Errorf("Unexpected ExchangeAddress value: %v", processedOrder.ExchangeAddress)
	}
	if makerTokenAmount := new(big.Int).SetBytes(processedOrder.MakerTokenAmount[:]); makerTokenAmount.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Errorf("Unexpected makerTokenAmount: %v", makerTokenAmount)
	}
	if takerTokenAmount := new(big.Int).SetBytes(processedOrder.TakerTokenAmount[:]); takerTokenAmount.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Errorf("Unexpected takerTokenAmount: %v", takerTokenAmount)
	}
}
