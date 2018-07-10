package zeroEx

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/google/subcommands"
	"github.com/notegio/massive/utils"
	"github.com/notegio/openrelay/types"
	"io"
	"log"
	"math/big"
	"os"
)

type csvReader struct {
	inputFileName  string
	outputFileName string
	inputFile      *os.File
	outputFile     *os.File
}

func (p *csvReader) FileNames() (string, string) {
	return p.inputFileName, p.outputFileName
}

func (p *csvReader) SetIOFiles(inputFile, outputFile *os.File) {
	p.inputFile, p.outputFile = inputFile, outputFile
}

func (*csvReader) Name() string     { return "csv" }
func (*csvReader) Synopsis() string { return "Parse orders out of a CSV" }
func (*csvReader) Usage() string {
	return `msv 0x csv [--input FILE] [--output FILE]:
  Parse orders out of a CSV and add them to a stream
`
}

func (p *csvReader) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.inputFileName, "input", "", "Input file [stdin]")
	f.StringVar(&p.outputFileName, "output", "", "Output file [stdout]")
}

func (p *csvReader) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 0 {
		os.Stderr.WriteString(p.Usage())
	}
	utils.SetIO(p)
	return CSVMain(p.inputFile, p.outputFile)
}

func CSVMain(inputFile io.Reader, outputFile io.Writer) subcommands.ExitStatus {
	csvReader := csv.NewReader(inputFile)
	headers, err := csvReader.Read()
	if err != nil {
		log.Printf("Error getting header: %v", err.Error())
		return subcommands.ExitFailure
	}
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[header] = i
	}
	counter := 0
	for {
		counter++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Error reading CSV: %v", err.Error())
			return subcommands.ExitFailure
		}
		order := &types.Order{}
		order.Initialize()
		if idx, ok := headerMap["maker"]; ok {
			var addressBytes []byte
			if record[idx] == "" {
				addressBytes = []byte{}
			} else {
				addressBytes, err = hex.DecodeString(record[idx][2:])
				if err != nil {
					log.Printf("Error parsing maker for record %v: %v. Dropping Record.", counter, err.Error())
					continue
				}
			}
			copy(order.Maker[:], addressBytes)
		}
		if idx, ok := headerMap["makerAssetData"]; ok {
			var assetDataBytes []byte
			if record[idx] == "" {
				assetDataBytes = []byte{}
			} else {
				assetDataBytes, err = hex.DecodeString(record[idx][2:])
				if err != nil {
					log.Printf("Error parsing maker makerAssetData record %v: %v. Dropping Record.", counter, err.Error())
					continue
				}
			}
			order.MakerAssetData = make(types.AssetData, len(assetDataBytes))
			copy(order.MakerAssetData[:], assetDataBytes)
		}
		if idx, ok := headerMap["makerAssetAmount"]; ok {
			value, ok := new(big.Int).SetString(record[idx], 10)
			if !ok {
				log.Printf("Error parsing  makerTokenAmount in record %v: %v. Dropping Record.", counter, record[idx])
				continue
			}
			copy(order.MakerAssetAmount[:], abi.U256(value))
		}
		if idx, ok := headerMap["makerFee"]; ok {
			value, ok := new(big.Int).SetString(record[idx], 10)
			if !ok {
				log.Printf("Error parsing  makerFee in record %v: %v. Dropping Record.", counter, record[idx])
				continue
			}
			copy(order.MakerFee[:], abi.U256(value))
		}
		if idx, ok := headerMap["taker"]; ok {
			var addressBytes []byte
			if record[idx] == "" {
				addressBytes = []byte{}
			} else {
				addressBytes, err = hex.DecodeString(record[idx][2:])
				if err != nil {
					log.Printf("Error parsing taker for record %v: %v. Dropping Record.", counter, err.Error())
					continue
				}
			}
			copy(order.Taker[:], addressBytes)
		}
		if idx, ok := headerMap["takerAssetData"]; ok {
			var assetDataBytes []byte
			if record[idx] == "" {
				assetDataBytes = []byte{}
			} else {
				assetDataBytes, err = hex.DecodeString(record[idx][2:])
				if err != nil {
					log.Printf("Error parsing takerAssetData record %v: %v. Dropping Record.", counter, err.Error())
					continue
				}
			}
			order.TakerAssetData = make(types.AssetData, len(assetDataBytes))
			copy(order.TakerAssetData[:], assetDataBytes)
		}
		if idx, ok := headerMap["takerAssetAmount"]; ok {
			value, ok := new(big.Int).SetString(record[idx], 10)
			if !ok {
				log.Printf("Error parsing  takerAssetAmount in record %v: %v. Dropping Record.", counter, record[idx])
				continue
			}
			copy(order.TakerAssetAmount[:], abi.U256(value))
		}
		if idx, ok := headerMap["takerFee"]; ok {
			value, ok := new(big.Int).SetString(record[idx], 10)
			if !ok {
				log.Printf("Error parsing  takerFee in record %v: %v. Dropping Record.", counter, record[idx])
				continue
			}
			copy(order.TakerFee[:], abi.U256(value))
		}
		if idx, ok := headerMap["expirationUnixTimestampSec"]; ok {
			value, ok := new(big.Int).SetString(record[idx], 10)
			if !ok {
				log.Printf("Error parsing  expirationUnixTimestampSec in record %v: %v. Dropping Record.", counter, record[idx])
				continue
			}
			copy(order.ExpirationTimestampInSec[:], abi.U256(value))
		}
		if idx, ok := headerMap["feeRecipient"]; ok {
			var addressBytes []byte
			if record[idx] == "" {
				addressBytes = []byte{}
			} else {
				addressBytes, err = hex.DecodeString(record[idx][2:])
				if err != nil {
					log.Printf("Error parsing feeRecipient for record %v: %v. Dropping Record.", counter, err.Error())
					continue
				}
			}
			copy(order.FeeRecipient[:], addressBytes)
		}
		if idx, ok := headerMap["salt"]; ok {
			value, ok := new(big.Int).SetString(record[idx], 10)
			if !ok {
				log.Printf("Error parsing salt in record %v: %v. Dropping Record.", counter, record[idx])
				continue
			}
			copy(order.Salt[:], abi.U256(value))
		}
		if idx, ok := headerMap["signature"]; ok {
			dataBytes, err := hex.DecodeString(record[idx][2:])
			if err != nil {
				log.Printf("Error parsing signature in record %v: %v. Dropping Record.", counter, err.Error())
				continue
			}
			copy(order.Signature, dataBytes)
		}
		if idx, ok := headerMap["exchangeContractAddress"]; ok {
			var addressBytes []byte
			if record[idx] == "" {
				addressBytes = []byte{}
			} else {
				addressBytes, err = hex.DecodeString(record[idx][2:])
				if err != nil {
					log.Printf("Error parsing maker exchangeContractAddress record %v: %v. Dropping Record.", counter, err.Error())
					continue
				}
			}
			copy(order.ExchangeAddress[:], addressBytes)
		}
		utils.WriteRecord(order, outputFile)
	}
	return subcommands.ExitSuccess
}
