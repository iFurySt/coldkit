package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ifuryst/coldkit/internal/mcp"
)

func main() {
	var enableSecretTools bool
	flag.BoolVar(&enableSecretTools, "enable-secret-tools", false, "enable MCP tools that can return private keys")
	flag.Parse()

	server := mcp.NewServer(mcp.Config{EnableSecretTools: enableSecretTools})
	if err := server.Serve(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
