package main

import (
	_ "embed"
	"fmt"
	"os"
)

const (
	FmtErrUnknownOpt = "Unknown option for subcommand '%s': '%s'!"
)

var (
	//go:embed resources/map
	symbolMap string

	//go:embed resources/versions
	versions string
)

func print_help() {
	fmt.Printf(`
ppeb's pulsar tools!!!

Usage: ./pulsar-tools [subcommand] [options]
 crash            Analyze crashdump
     -f|--file        Crashdump to analyze. Pass 'stdin' to accept a crashdump over stdin
 import-config    Analyze a config.pul file
     -f|--file        Config to analyze. Pass 'stdin' to accept a crashdump over stdin
`)

	os.Exit(1)
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "No arguments provided! A subcommand is required to continue.")
		print_help()
	}

	subcommand := os.Args[1]
	var opts []string

	if len(subcommand) > 2 {
		opts = os.Args[2:]
	} else {
		opts = []string{}
	}

	var err error

	switch subcommand {
	case "-h", "--help", "help":
		print_help()
	case "crash":
		err = crash(opts)
	case "import-config":
		err = import_config(opts)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand %s!\n", subcommand)
		print_help()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while running subcommand '%s': %s\n", subcommand, err.Error())
		os.Exit(1)
	}
}
