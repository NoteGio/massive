package zeroEx

import (
	"bufio"
	"encoding/json"
	"github.com/notegio/openrelay/types"
	"io"
	"log"
)

func orderScanner(fd io.Reader) chan *types.Order {
	channel := make(chan *types.Order)
	go func() {
		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			line := scanner.Bytes()
			order := &types.Order{}
			err := json.Unmarshal(line, order)
			if err != nil {
				log.Fatalf("Error parsing record: %v", err.Error())
			}
			channel <- order
		}
		close(channel)
	}()
	return channel
}
