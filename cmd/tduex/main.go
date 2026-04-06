package main

import (
	"fmt"
	"os"

	"github.com/med-000/tduscheexport/internal/tduexcli"
)

func main() {
	if err := tduexcli.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
