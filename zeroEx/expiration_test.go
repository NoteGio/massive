package zeroEx_test

import (
  "math/big"
  "encoding/json"
  "github.com/notegio/massive/zeroEx"
  "github.com/notegio/openrelay/types"
  "github.com/google/subcommands"
  "testing"
  "bufio"
  "bytes"
  "time"
)

func TestSetExpirationTimestamp(t *testing.T) {
  order := &types.Order{}
  order.Initialize()
  orderBytes, err := json.Marshal(order)
  if err != nil {
    t.Errorf(err.Error())
  }
  inputFile := bytes.NewReader(orderBytes[:])
  outputBuffer := &bytes.Buffer{}
  outputFile := bufio.NewWriter(outputBuffer)
  if status := zeroEx.SetExpirationMain(inputFile, outputFile, false, big.NewInt(10)); status != subcommands.ExitSuccess {
    t.Fatalf("Bad exitcode: %v", status)
  }
  outputFile.Flush()
  processedOrder := &types.Order{}
  err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
  if err != nil {
    t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
  }
  expiration := new(big.Int).SetBytes(processedOrder.ExpirationTimestampInSec[:])
  if expiration.Int64() != 10 {
    t.Fatalf("Unexpected expiration value: %v", expiration)
  }
}

func TestSetExpirationDuration(t *testing.T) {
  order := &types.Order{}
  order.Initialize()
  orderBytes, err := json.Marshal(order)
  if err != nil {
    t.Errorf(err.Error())
  }
  inputFile := bytes.NewReader(orderBytes[:])
  outputBuffer := &bytes.Buffer{}
  outputFile := bufio.NewWriter(outputBuffer)
  if status := zeroEx.SetExpirationMain(inputFile, outputFile, true, big.NewInt(10)); status != subcommands.ExitSuccess {
    t.Fatalf("Bad exitcode: %v", status)
  }
  outputFile.Flush()
  processedOrder := &types.Order{}
  err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
  if err != nil {
    t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
  }
  expiration := new(big.Int).SetBytes(processedOrder.ExpirationTimestampInSec[:])
  if delta := expiration.Int64() - time.Now().Unix(); delta < 10 || delta > 11 {
    t.Fatalf("Unexpected expiration value: %v", delta)
  }
}
