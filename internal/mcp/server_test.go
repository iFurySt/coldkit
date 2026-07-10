package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestToolsListHidesSecretToolsByDefault(t *testing.T) {
	output := runServer(t, NewServer(Config{}), `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`+"\n")
	if strings.Contains(output, "tron_generate_secret") {
		t.Fatalf("secret tool should be hidden by default: %s", output)
	}
	if !strings.Contains(output, "tron_generate_preview") {
		t.Fatalf("preview tool missing: %s", output)
	}
}

func TestToolsListCanExposeSecretTools(t *testing.T) {
	output := runServer(t, NewServer(Config{EnableSecretTools: true}), `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`+"\n")
	if !strings.Contains(output, "tron_generate_secret") {
		t.Fatalf("secret tool missing: %s", output)
	}
}

func TestValidateTool(t *testing.T) {
	output := runServer(t, NewServer(Config{}), `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"tron_validate","arguments":{"address":"TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3"}}}`+"\n")
	var resp response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	if !strings.Contains(output, "4162f94e9ac9349bccc61bfe66ddade6292702ecb6") {
		t.Fatalf("hex address missing: %s", output)
	}
}

func runServer(t *testing.T, server *Server, input string) string {
	t.Helper()
	var out strings.Builder
	if err := server.Serve(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}
