// This file is partially adapted from Pulsar. All logic for processing config
// files was created referencing Pulsar's Pack Creator

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

var (
	ErrNoConfig      = errors.New("No config.pul was provided! A config.pul is required to analyze.")
	ErrConfigMagic   = errors.New("Mismatched magic bits! config.pul appears to be invalid.")
	ErrInfoMagic     = errors.New("Mismatched Info magic bits! config.pul appears to be invalid.")
	ErrCupsMagic     = errors.New("Mismatched Cups magic bits! config.pul appears to be invalid.")
	ErrBMGMagic      = errors.New("Mismatched BMG magic bits! config.pul's BMG section appears to be invalid.")
	ErrFileMagic     = errors.New("Mismatched FILE magic bits! config.pul's FILE section appears to be invalid.")
	ErrConfigVersion = errors.New("Your config.pul file is too old! Please convert it to V3 using PulsarPackCreator.")
)

const (
	CONFIGMAGIC = 0x50554C53
	INFOMAGIC   = 0x494E464F
	CUPSMAGIC   = 0x43555053
	TEXTMAGIC   = 0x54455854
	FILEMAGIC   = 0x46494C45
	BMGMAGIC    = 0x4D455347626D6731
)

const (
	BMG_CUPS    = 0x10000
	BMG_TRACKS  = 0x20000
	BMG_AUTHORS = 0x30000
)

func import_config(opts []string) error {
	var file string
	optsLen := len(opts)
	for i := 0; i < optsLen; i++ {
		opt := opts[i]

		switch opt {
		case "-f", "--file":
			if optsLen > i+1 {
				file = opts[i+1]
				i++
			}
		default:
			return fmt.Errorf("Unknown option for subcommand 'crash', '%s'!\n", opt)
		}
	}

	if len(file) == 0 {
		return ErrNoConfig
	}

	err := os.Mkdir("temp", 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	wbmgt, err := os.OpenFile("temp/wbmgt", os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	_, err = wbmgt.Write(wbmgtBytes)
	err = wbmgt.Close()
	if err != nil {
		return err
	}

	var bytes []byte
	if file == "stdin" {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(file)
	}

	if err != nil {
		return err
	}

	var config Config
	_, err = UnmarshalPulsarType(bytes, &config.Header)
	if err != nil {
		return err
	}

	if config.Header.Magic != CONFIGMAGIC {
		return ErrConfigMagic
	}

	// Far too lazy to support legacy versions. Force version 3
	if config.Header.Version != 3 {
		return ErrConfigVersion
	}

	config.Header.ModFolderName = strings.TrimLeft(config.Header.ModFolderName, "/\\")

	_, err = UnmarshalPulsarType(bytes[config.Header.InfoOffset:], &config.Info)
	if err != nil {
		return err
	}

	if config.Info.Header.Magic != INFOMAGIC {
		return ErrInfoMagic
	}

	config.Cups = []CupHolder{}

	_, err = UnmarshalPulsarType(bytes[config.Header.CupsOffset:], &config.CupInfo)
	if err != nil {
		return err
	}

	if config.CupInfo.Header.Magic != CUPSMAGIC {
		return ErrCupsMagic
	}

	tvcField, _ := reflect.TypeOf(config.CupInfo).FieldByName("TotalVariantCount")
	trackOffset := config.Header.CupsOffset + int32(tvcField.Offset) + int32(unsafe.Sizeof(int32(0)))
	trackSize := int32(reflect.TypeOf((*TrackV3)(nil)).Elem().Size())

	variantOffset := trackOffset + int32(trackSize)*int32(config.CupInfo.CtsCupCount)*4

	for range int(config.CupInfo.CtsCupCount) {
		tracks := []TrackV3{}
		variants := []Variant{} // Max 8 variants

		for j := range 4 {
			track := TrackV3{}
			nTrack, err := UnmarshalPulsarType(bytes[trackOffset:], &track)
			if err != nil {
				return err
			}

			for range int(tracks[j].VariantCount) {
				variant := Variant{}
				nVariant, err := UnmarshalPulsarType(bytes[variantOffset:], &variant)
				if err != nil {
					return err
				}

				variants = append(variants, variant)
				variantOffset += int32(nVariant)
			}

			tracks = append(tracks, track)
			trackOffset += int32(nTrack)
		}

		config.Cups = append(config.Cups, CupHolder{
			Tracks:   tracks,
			Variants: variants,
		})
	}

	// Should be at the bmg header now!
	bmgSize, _, err := readBMG(bytes, int(config.Header.BMGOffset))
	if err != nil {
		return err
	}

	_, err = readFile(bytes, int(config.Header.BMGOffset+int32(bmgSize)))
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", config)

	return nil
}

// Reads and decodes the bmg section
func readBMG(bytes []byte, offset int) (int, []byte, error) {
	magic := binary.BigEndian.Uint64(bytes[offset : offset+8])

	if magic != BMGMAGIC {
		return 8, nil, ErrBMGMagic
	}

	size := int(binary.BigEndian.Uint32(bytes[offset+8 : offset+12]))

	decoded, err := DecodeBMG(bytes[offset : offset+size])
	if err != nil {
		return size, nil, err
	}

	return size, decoded, nil
}

// FILE section, not arbitrary file...
func readFile(bytes []byte, offset int) ([]byte, error) {
	magic := binary.BigEndian.Uint32(bytes[offset : offset+4])

	if magic != FILEMAGIC {
		return nil, ErrBMGMagic
	}

	return bytes[offset:], nil
}

// Parses a decoded BMG
func (config *Config) parseBMG(bmgBytes []byte) error {
	s := bufio.NewScanner(bytes.NewReader(bmgBytes))

	for s.Scan() {
		curline := s.Text()
		var bmgID uint64

		if len(curline) < 5 {
			continue
		}

		// Parse correctly mfw
		bmgID, err := strconv.ParseUint(curline[1:5], 16, 32)
		if err != nil {
			continue
		}

		if bmgID == 0x2847 {
			lsplit := strings.Split(curline, " ")
			// TODO: assign to date. fuck if I know what to do with it
			_ = lsplit[len(lsplit)-1]
			continue
		}

		if bmgID < 0x10000 || bmgID >= 0x60000 {
			continue
		}

		lsplit := strings.Split(curline, "=")
		if len(lsplit) < 2 {
			continue
		}

		content := lsplit[1]

		typ := bmgID & 0xFFFF0000
		rest := bmgID & 0x0000FFFF
		varIdx := (rest & 0x0000F000) >> 12
		cupIdx := (rest & 0x0000FFF0) / 4
		trackIdx := (rest & 0x00000FFF) % 4

		switch typ {
		case BMG_CUPS:
			if rest >= uint64(config.CupInfo.CtsCupCount) {
				break
			}

			config.Cups[rest].Name = content
		case BMG_TRACKS, BMG_AUTHORS:
			track := &config.Cups[cupIdx].Tracks[trackIdx]

			if varIdx >= 8 {
				track.Name = content
				break
			}

			var variant Variant
			if varIdx == 0 {
				variant = track.Main
			} else {
				variant = track.
			}
		}
	}

	return nil
}
