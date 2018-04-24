package zeroEx_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/google/subcommands"
	"github.com/notegio/massive/zeroEx"
	"github.com/notegio/openrelay/types"
	"math/big"
	"testing"
	"time"
)

func TestSetSaltTimestamp(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SetSaltMain(inputFile, outputFile, false, -1); status != subcommands.ExitSuccess {
		t.Fatalf("Bad exitcode: %v", status)
	}
	outputFile.Flush()
	processedOrder := &types.Order{}
	err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
		t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
	}
	salt := new(big.Int).SetBytes(processedOrder.Salt[:])
	timeStamp := big.NewInt(time.Now().Unix())
	if new(big.Int).Abs(new(big.Int).Sub(timeStamp, salt)).Int64() > 1 {
		t.Fatalf("Unexpected salt value: %v", salt)
	}
}

func TestSetSaltValue(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SetSaltMain(inputFile, outputFile, false, 1); status != subcommands.ExitSuccess {
		t.Fatalf("Bad exitcode: %v", status)
	}
	outputFile.Flush()
	processedOrder := &types.Order{}
	err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
		t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
	}
	salt := new(big.Int).SetBytes(processedOrder.Salt[:])
	if salt.Int64() != 1 {
		t.Errorf("Unexpected Salt value: %v", salt)
	}
}

func TestSetSaltRandom(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SetSaltMain(inputFile, outputFile, true, -1); status != subcommands.ExitSuccess {
		t.Fatalf("Bad exitcode: %v", status)
	}
	outputFile.Flush()
	processedOrder := &types.Order{}
	err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
		t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
	}
	salt := new(big.Int).SetBytes(processedOrder.Salt[:])
	if salt.Int64() == 0 {
		t.Errorf("Unexpected Salt value: %v", salt)
	}
}
