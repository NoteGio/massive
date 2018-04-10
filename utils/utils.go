package utils

import (
  "encoding/json"
  "os"
  "io"
)

// IOCommand represents a command that has an Input and Output file. It should
// provide a FileNames() method, returning (inputFilePath, outputFilePath), and
// a SetIOFiles(inputFile, outputFile) method to set the file objects. If
// either inputFilePath or outputFilePath from IOCommand.FileNames() is blank,
// stdin or stdout respectively will be used.
type IOCommand interface {
  FileNames() (string, string)
  SetIOFiles(*os.File, *os.File)
}


func SetIO(cmd IOCommand) error {
  inputFileName, outputFileName := cmd.FileNames()
  inputFile := os.Stdin
  outputFile := os.Stdout
  var err error
  if inputFileName != "" {
    inputFile, err = os.Open(inputFileName)
    if err != nil {
      return err
    }
  }
  if outputFileName != "" {
    outputFile, err = os.OpenFile(outputFileName, os.O_RDWR|os.O_CREATE, 644)
    if err != nil {
      return err
    }
  }
  cmd.SetIOFiles(inputFile, outputFile)
  return nil
}

func WriteRecord(item interface{}, outFile io.Writer) error {
  data, err := json.Marshal(item)
  if err != nil {
    return err
  }
  _, err = outFile.Write(append(data, []byte("\n")...))
  if err != nil {
    return err
  }
  return nil
}
