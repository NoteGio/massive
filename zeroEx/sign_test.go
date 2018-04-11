package zeroEx_test

import (
  "encoding/json"
  "crypto/ecdsa"
  "crypto/rand"
  "github.com/notegio/massive/zeroEx"
  "github.com/notegio/openrelay/types"
  "github.com/google/subcommands"
  "github.com/ethereum/go-ethereum/crypto"
  "testing"
  "bufio"
  "bytes"
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
  if status := zeroEx.SignOrderMain(inputFile, outputFile, key, false); status != subcommands.ExitSuccess {
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

func TestSignOrderNoReplaceSame(t *testing.T) {
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
  if status := zeroEx.SignOrderMain(inputFile, outputFile, key, true); status != subcommands.ExitSuccess {
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

func TestSignOrderNoReplaceDifferent(t *testing.T) {
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
  if status := zeroEx.SignOrderMain(inputFile, outputFile, key, true); status != subcommands.ExitFailure {
    t.Fatalf("Bad exitcode: %v", status)
  }
}
