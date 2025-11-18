// Package main provides the entry point for the kira CLI application.
package main

import (
	"fmt"
	"os"

	"kira/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
