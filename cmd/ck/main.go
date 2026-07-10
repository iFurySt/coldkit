package main

import (
	"os"

	"github.com/ifuryst/coldkit/internal/cli"
)

func main() {
	if err := cli.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
