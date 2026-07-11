package mcpinstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveAgentAliases(t *testing.T) {
	tests := map[string]Agent{
		"codex":       AgentCodex,
		"claude-code": AgentClaudeCode,
		"claude":      AgentClaudeCode,
		"cloud-code":  AgentClaudeCode,
		"cloud_code":  AgentClaudeCode,
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := ResolveAgent(input)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestInstallCodexProjectConfig(t *testing.T) {
	dir := t.TempDir()
	result, err := Install(Options{
		Agent:   AgentCodex,
		Project: true,
		WorkDir: dir,
		HomeDir: filepath.Join(dir, "home"),
		Command: "/usr/local/bin/ck-mcp",
	})
	if err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(dir, ".codex", "config.toml")
	if result.Path != wantPath {
		t.Fatalf("got path %q, want %q", result.Path, wantPath)
	}

	content := readFile(t, wantPath)
	if !strings.Contains(content, "[mcp_servers.coldkit]") {
		t.Fatalf("missing codex mcp table:\n%s", content)
	}
	if !strings.Contains(content, `command = "/usr/local/bin/ck-mcp"`) {
		t.Fatalf("missing command:\n%s", content)
	}
	if !strings.Contains(content, "args = []") {
		t.Fatalf("missing args:\n%s", content)
	}
}

func TestInstallCodexReplacesExistingColdkitTable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[mcp_servers.other]\ncommand = \"other\"\nargs = []\n\n[mcp_servers.coldkit]\ncommand = \"old\"\nargs = [\"x\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Install(Options{
		Agent:   AgentCodex,
		Project: true,
		WorkDir: dir,
		HomeDir: filepath.Join(dir, "home"),
		Command: "ck-mcp",
	})
	if err != nil {
		t.Fatal(err)
	}

	content := readFile(t, path)
	if strings.Contains(content, `command = "old"`) {
		t.Fatalf("old coldkit table remained:\n%s", content)
	}
	if !strings.Contains(content, "[mcp_servers.other]") {
		t.Fatalf("unrelated table was removed:\n%s", content)
	}
	if count := strings.Count(content, "[mcp_servers.coldkit]"); count != 1 {
		t.Fatalf("got %d coldkit tables:\n%s", count, content)
	}
}

func TestInstallClaudeCodeGlobalConfig(t *testing.T) {
	home := t.TempDir()
	result, err := Install(Options{
		Agent:   AgentClaudeCode,
		HomeDir: home,
		Command: "ck-mcp",
	})
	if err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(home, ".claude.json")
	if result.Path != wantPath {
		t.Fatalf("got path %q, want %q", result.Path, wantPath)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(readFile(t, wantPath)), &parsed); err != nil {
		t.Fatal(err)
	}
	servers := parsed["mcpServers"].(map[string]any)
	server := servers["coldkit"].(map[string]any)
	if server["command"] != "ck-mcp" {
		t.Fatalf("got command %v", server["command"])
	}
	if args, ok := server["args"].([]any); !ok || len(args) != 0 {
		t.Fatalf("got args %#v", server["args"])
	}
}

func TestInstallRejectsUnsafeServerName(t *testing.T) {
	_, err := Install(Options{
		Agent:      AgentCodex,
		HomeDir:    t.TempDir(),
		ServerName: "bad.name",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
