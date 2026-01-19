package main

import (
	"fmt"
	"os"

	"github.com/asjdf/p2p-playground-lite/cmd/controller/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
