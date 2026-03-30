package main

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrNoCode = errors.New("No Code.pul was provided! A Code.pul is required to hash.")
)

func hash(opts []string, _ Arguments) error {
	file := ""
	optsLen := len(opts)
	for i := 0; i < optsLen; i++ {
		opt := opts[i]

		switch opt {
		case "-f", "--file":
			if optsLen > i+1 {
				file = opts[i+1]
				i++
			}
			// default:
			// return fmt.Errorf("Unknown option for subcommand 'crash', '%s'!\n", opt)
		}
	}

	if len(file) == 0 {
		return ErrNoCode
	}

	var bytes []byte
	var err error
	if file == "stdin" {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(file)
	}

	if err != nil {
		return err
	}

	regionSizes := []int32{}
	for i := range 4 {
		var regionSize int32
		binary.Decode(bytes[i*4:(i*4)+4], binary.BigEndian, &regionSize)
		regionSizes = append(regionSizes, regionSize)
	}

	var offset int32 = 0x10

	for i := range 4 {
		if regionSizes[i] == 0 {
			fmt.Printf("Region %s is empty\n", regionIdxToName(i))
			continue
		}

		if i > 0 {
			offset += regionSizes[i-1]
		}

		header := bytes[offset : offset+0x20]

		var size int32
		_, err := binary.Decode(header[0xc:], binary.BigEndian, &size)
		if err != nil {
			return err
		}

		data := bytes[offset+0x20 : offset+0x20+size]

		var magic uint64
		_, err = binary.Decode(header, binary.BigEndian, &magic)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", regionIdxToName(i))

		fmt.Printf("Hash: ")

		sha1 := sha1.Sum(data)
		for _, b := range sha1 {
			fmt.Printf("%02x", b)
		}

		fmt.Printf("\nMagic: %d\nOffset: %d\n\n", magic, offset)
	}

	return nil
}

func regionIdxToName(idx int) string {
	switch idx {
	case 0:
		return "PAL"
	case 1:
		return "NTSCU"
	case 2:
		return "NTSCJ"
	case 3:
		return "NTSCK"
	default:
		return "Unknown Region"
	}
}
