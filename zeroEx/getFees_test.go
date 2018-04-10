package zeroEx_test

import (
  "encoding/json"
  "github.com/notegio/massive/zeroEx"
  "github.com/notegio/openrelay/types"
  "github.com/google/subcommands"
  "testing"
  "bufio"
  "bytes"
  "os"
)

func TestGetFees(t *testing.T) {
  order := &types.Order{}
  order.Initialize()
  orderBytes, err := json.Marshal(order)
  if err != nil {
    t.Errorf(err.Error())
  }
  inputFile := bytes.NewReader(orderBytes[:])
  outputBuffer := &bytes.Buffer{}
  outputFile := bufio.NewWriter(outputBuffer)
  targetURL := os.Getenv("RELAY_URL")
  if targetURL == "" {
    targetURL = "https://api.openrelay.xyz"
  }
  if status := zeroEx.GetFeesMain(targetURL, inputFile, outputFile, 0.0); status != subcommands.ExitSuccess {
    t.Fatalf("Bad exitcode: %v", status)
  }
  outputFile.Flush()
  processedOrder := &types.Order{}
  err = json.Unmarshal(outputBuffer.Bytes(), processedOrder)
  if err != nil {
    t.Fatalf("Error parsing '%v': %v", string(outputBuffer.Bytes()), err.Error())
  }
  if bytes.Equal(processedOrder.FeeRecipient[:], order.FeeRecipient[:]) {
    t.Errorf("Expected FeeRecipient from API, got %v", processedOrder.FeeRecipient)
  }
  if !bytes.Equal(processedOrder.MakerFee[:], order.MakerFee[:]) {
    t.Errorf("Expected MakerFee to be 0, got %v", processedOrder.MakerFee)
  }
  if bytes.Equal(processedOrder.TakerFee[:], order.TakerFee[:]) {
    t.Errorf("Expected TakerFee to be non-zero, got %v", processedOrder.TakerFee)
  }
}
