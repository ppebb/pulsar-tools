package main

import (
	"io"
	"os"
)

func readFile(file string) ([]byte, error) {
	var bytes []byte
	var err error
	if file == "stdin" {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(file)
	}

	return bytes, err
}
