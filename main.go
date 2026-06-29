package main

import (
	"fmt"
	"os"

	"github.com/gofuego/fuego-dotclaude/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "fuego-dotclaude: %v\n", err)
		os.Exit(1)
	}
}
