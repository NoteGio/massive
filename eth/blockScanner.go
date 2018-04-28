
package eth

import (
	"bufio"
	"io"
	"log"

)

func blockScanner(fd io.Reader) chan *blockWithHeader {
	channel := make(chan *blockWithHeader)
	go func() {
		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			line := scanner.Bytes()
			block, err := getBlock(line)
			if err != nil {
				log.Fatalf("Error parsing record: %v", err.Error())
			}
			channel <- block
		}
		close(channel)
	}()
	return channel
}
