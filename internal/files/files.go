package files

import (
	"io"
	"log"
	"os"
)

func Open(path string) *os.File {
	if file, err := os.Open(path); err != nil {
		log.Printf("Error open file: %v", err)
		return nil
	} else {
		return file
	}
}

func Read(file *os.File) []byte {
	if bytes, err := io.ReadAll(file); err != nil {
		log.Printf("Error read file: %v", err)
		return nil
	} else {
		return bytes
	}
}
