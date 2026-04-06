package tduexcli

import (
	"fmt"
	"os"
)

func Run(args []string) error {
	if len(args) < 2 {
		return runInteractive()
	}

	switch args[1] {
	case "classes":
		return runClasses(args[2:])
	case "full":
		return runFull(args[2:])
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return nil
	default:
		printUsage(os.Stderr)
		return fmt.Errorf("unknown subcommand: %s", args[1])
	}
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "usage: tduex [classes|full] [options]")
	fmt.Fprintln(out, "example: tduex full -year 2025 -term 1 -format json,csv,ics")
}
