
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
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024*5)
		for scanner.Scan() {
			line := scanner.Bytes()
			block, err := getBlock(line)
			if err != nil {
				log.Fatalf("Error parsing record: %v", err.Error())
			}
			channel <- block
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error from scanner: %v", err.Error())
		}
		close(channel)
	}()
	return channel
}
