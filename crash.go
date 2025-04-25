package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrNoCrashdump = errors.New("No crashdump was provided! A crashdump is required to analyze.")
	ErrMagic       = errors.New("Mismatched magic bits! Crash file appears to be invalid.")
)

const (
	FmtErrTooShort = "The provided crashdump was too short (%d bytes) to analyze."
)

var (
	file = ""
)

func crash(opts []string) error {
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
		return ErrNoCrashdump
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

	err = verifyMagic(bytes)
	if err != nil {
		return err
	}

	var exFile ExceptionFile
	err = UnmarshalPulsarType(bytes, &exFile)
	if err != nil {
		return err
	}

	region := string(byte(exFile.Region))

	out := "" +
		fmt.Sprintf("Error: %s\n", oserrorString(OSERROR(exFile.Err))) +
		fmt.Sprintf("Region: %s\n", region) +
		fmt.Sprintf("SSR0: 0x%x, %s\n", exFile.Srr0.Gpr, resolveSyms(exFile.Srr0.Gpr, region)) +
		fmt.Sprintf("SSR1: 0x%x\n", exFile.Srr1.Gpr) +
		fmt.Sprintf("MSR:  0x%x\n", exFile.Msr.Gpr) +
		fmt.Sprintf("CR:   0x%x\n", exFile.Cr.Gpr) +
		fmt.Sprintf("LR:   0x%x, %s\n", exFile.Lr.Gpr, resolveSyms(exFile.Lr.Gpr, region)) +
		"\nGPRs\n"

	for i := range 8 {
		for j := range 4 {
			idx := i + j*8
			out += fmt.Sprintf("R%02d: 0x%08x ", idx, exFile.Gprs[idx].Gpr)
		}

		out += "\n"
	}

	out += fmt.Sprintf("\nFPRs FSCR: 0x%016x\n", math.Float64bits(exFile.Fpscr.Fpr))
	for i := range 8 {
		for j := range 4 {
			idx := i + j*8
			exp := padExponent(fmt.Sprintf("% 03.02e", exFile.Fprs[idx].Fpr), 3)
			out += fmt.Sprintf("F%.2d: %s ", idx, exp)
		}

		out += "\n"
	}

	out += "\nStack Frame\n"
	for i := range 10 {
		out += fmt.Sprintf(
			"SP: 0x%x, LR: 0x%08x, %s\n",
			exFile.Frames[i].Sp,
			exFile.Frames[i].Lr,
			resolveSyms(exFile.Frames[i].Lr, region),
		)
	}

	fmt.Println(out)

	return nil
}

const MAGIC = 0x50554c44

func verifyMagic(bytes []byte) error {
	if len(bytes) < 1000 {
		return fmt.Errorf(FmtErrTooShort, len(bytes))
	}

	fmagic := int(bytes[3]) | int(bytes[2])<<8 | int(bytes[1])<<16 | int(bytes[0])<<24
	if fmagic != MAGIC {
		return ErrMagic
	}

	return nil
}

func oserrorString(err OSERROR) string {
	switch err {
	case OSERROR_DSI:
		return "DSI"
	case OSERROR_ISI:
		return "ISI"
	case OSERROR_FLOATING_POINT, OSERROR_FPE:
		return "FPE"
	}

	return "Unknown"
}

// Fuck if I know what any of this is. All stolen from Pulsar Crash.xaml.cs.
// Thank god for MIT licensing or I'd've been done for.
func resolveSyms(addr uint32, region string) string {
	ret := ""
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Failed to resolve symbols for %08x! %s\n", addr, r)
			ret = fmt.Sprintf("Failed to resolve syms for %08x", addr)
		}
	}()

	lines := strings.Split(symbolMap, "\n")
	vtemp := strings.ReplaceAll(versions, "-8", " 8")
	vtemp = strings.ReplaceAll(vtemp, ":", "")
	splices := strings.Split(vtemp, "\n")

	idx := slices.Index(splices, fmt.Sprintf("[%s]", region))
	if idx == -1 {
		return fmt.Sprintf("Failed to resolve symbols for 0x%08x! Region not located in versions\n", addr)
	}

	if region != "P" {
		idx++

		for {
			if splices[idx] == "#" {
				return fmt.Sprintf("Unknown (0x%08x)", addr)
			}

			curSplice := strings.Split(splices[idx], " ")
			var isNegative int32 = 1
			if strings.Contains(curSplice[2], "-") {
				isNegative = -1
			}

			noPrefixOffset := strings.ReplaceAll(curSplice[2], "-0x", "")
			noPrefixOffset = strings.ReplaceAll(noPrefixOffset, "+0x", "")

			offset, err := strconv.ParseUint(noPrefixOffset, 16, 32)
			if err != nil {
				return fmt.Sprintf("Failed to resolve symbols for 0x%08x. Unable to parse addr offset '%s'! %s", addr, noPrefixOffset, err)
			}

			palAddr := uint32(int32(addr) - isNegative*int32(offset))

			lower, err := strconv.ParseUint(curSplice[0], 16, 32)
			if err != nil {
				return fmt.Sprintf("Failed to resolve symbols for 0x%08x. Unable to parse lower address '%s'! %s", addr, curSplice[0], err)
			}

			upper, err := strconv.ParseUint(curSplice[1], 16, 32)
			if err != nil {
				return fmt.Sprintf("Failed to resolve symbols for 0x%08x. Unable to parse upper address '%s'! %s", addr, curSplice[1], err)
			}

			if uint32(lower) <= palAddr && palAddr < uint32(upper) {
				addr = palAddr
				break
			}

			idx++
		}
	}

	for i, line := range lines {
		curLine := strings.Split(line, " ")
		nextLine := strings.Split(lines[i+1], " ")

		curNum, err := strconv.ParseUint(curLine[0], 16, 32)
		if err != nil {
			return fmt.Sprintf("Failed to resolve symbols for 0x%08x. Unable to parse curNum '%s'! %s", addr, curLine[0], err)
		}

		nextNum, err := strconv.ParseUint(nextLine[0], 16, 32)
		if err != nil {
			return fmt.Sprintf("Failed to resolve symbols for 0x%08x. Unable to parse nextNum '%s'! %s", addr, nextLine[0], err)
		}

		if curNum <= uint64(addr) && uint64(addr) < nextNum {
			if region == "P" {
				return curLine[1]
			} else {
				return fmt.Sprintf("%s (0x%x)", curLine[1], addr)
			}
		}
	}

	ret = fmt.Sprintf("Unknown (0x%08x)", addr)

	return ret
}

func padExponent(exp string, width int) string {
	var sign string

	if strings.Contains(exp, "+") {
		sign = "+"
	} else {
		sign = "-"
	}

	splits := strings.Split(exp, sign)

	if len(splits) != 2 {
		return exp
	}

	return splits[0] + sign + strings.Repeat("0", width-len(splits[1])) + splits[1]
}
