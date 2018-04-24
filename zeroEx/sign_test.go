package zeroEx_test

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/subcommands"
	"github.com/notegio/massive/zeroEx"
	"github.com/notegio/openrelay/types"
	"testing"
)

func TestSignOrderReplace(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	if order.Signature.Verify(order.Maker) {
		t.Errorf("Initial order unexpectedly has valid signature")
	}
	key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	address := crypto.PubkeyToAddress(key.PublicKey)
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SignOrderMain(inputFile, outputFile, key, false, true); status != subcommands.ExitSuccess {
		t.Fatalf("Bad exitcode: %v", status)
	}
	outputFile.Flush()
	processedOrder := &types.Order{}
	err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
		t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
	}
	if !processedOrder.Signature.Verify(processedOrder.Maker) {
		t.Errorf("Signature should be valid")
	}
	if !bytes.Equal(processedOrder.Maker[:], address[:]) {
		t.Errorf("Address mismatch: %v != %v", processedOrder.Maker, address)
	}
}

func TestSignOrderNoReplace(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	if order.Signature.Verify(order.Maker) {
		t.Errorf("Initial order unexpectedly has valid signature")
	}
	key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	address := crypto.PubkeyToAddress(key.PublicKey)
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SignOrderMain(inputFile, outputFile, key, false, false); status != subcommands.ExitSuccess {
		t.Fatalf("Bad exitcode: %v", status)
	}
	outputFile.Flush()
	processedOrder := &types.Order{}
	err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
		t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
	}
	if processedOrder.Signature.Verify(processedOrder.Maker) {
		t.Errorf("Signature should not be valid")
	}
	if bytes.Equal(processedOrder.Maker[:], address[:]) {
		t.Errorf("Address mismatch: %v == %v", processedOrder.Maker, address)
	}
}

func TestSignOrderErrReplaceSame(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	if order.Signature.Verify(order.Maker) {
		t.Errorf("Initial order unexpectedly has valid signature")
	}
	key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	address := crypto.PubkeyToAddress(key.PublicKey)
	copy(order.Maker[:], address[:])
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SignOrderMain(inputFile, outputFile, key, true, false); status != subcommands.ExitSuccess {
		t.Fatalf("Bad exitcode: %v", status)
	}
	outputFile.Flush()
	processedOrder := &types.Order{}
	err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
	if err != nil {
		t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
	}
	if !processedOrder.Signature.Verify(order.Maker) {
		t.Errorf("Signature should be valid")
	}
	if !bytes.Equal(order.Maker[:], address[:]) {
		t.Errorf("Address mismatch: %v != %v", order.Maker, address)
	}
}

func TestSignOrderErrReplaceDifferent(t *testing.T) {
	order := &types.Order{}
	order.Initialize()
	if order.Signature.Verify(order.Maker) {
		t.Errorf("Initial order unexpectedly has valid signature")
	}
	key, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	orderBytes, err := json.Marshal(order)
	if err != nil {
		t.Errorf(err.Error())
	}
	inputFile := bytes.NewReader(orderBytes[:])
	outputBuffer := &bytes.Buffer{}
	outputFile := bufio.NewWriter(outputBuffer)
	if status := zeroEx.SignOrderMain(inputFile, outputFile, key, true, false); status != subcommands.ExitFailure {
		t.Fatalf("Bad exitcode: %v", status)
	}
}
