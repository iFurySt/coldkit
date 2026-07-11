package mcpinstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const (
	defaultServerName = "coldkit"
	defaultCommand    = "ck-mcp"
)

type Agent string

const (
	AgentCodex      Agent = "codex"
	AgentClaudeCode Agent = "claude-code"
)

type AgentInfo struct {
	Name        Agent
	DisplayName string
	Aliases     []string
}

type Options struct {
	Agent      Agent
	Project    bool
	ServerName string
	Command    string
	HomeDir    string
	WorkDir    string
}

type Result struct {
	Agent      Agent
	Display    string
	Path       string
	ServerName string
	Command    string
	Project    bool
}

var serverNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func SupportedAgents() []AgentInfo {
	return []AgentInfo{
		{Name: AgentCodex, DisplayName: "Codex"},
		{Name: AgentClaudeCode, DisplayName: "Claude Code", Aliases: []string{"claude", "cloud-code"}},
	}
}

func ResolveAgent(value string) (Agent, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "-")

	switch normalized {
	case "codex":
		return AgentCodex, nil
	case "claude-code", "claude", "claudecode", "cloud-code", "cloudcode":
		return AgentClaudeCode, nil
	default:
		return "", fmt.Errorf("unsupported agent %q; supported agents: %s", value, strings.Join(SupportedAgentNames(), ", "))
	}
}

func SupportedAgentNames() []string {
	var names []string
	for _, agent := range SupportedAgents() {
		names = append(names, string(agent.Name))
		names = append(names, agent.Aliases...)
	}
	sort.Strings(names)
	return names
}

func DefaultCommand() string {
	executable, err := os.Executable()
	if err == nil {
		name := defaultCommand
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		candidate := filepath.Join(filepath.Dir(executable), name)
		if fileExists(candidate) {
			return candidate
		}
	}
	return defaultCommand
}

func Install(options Options) (Result, error) {
	if options.Agent == "" {
		return Result{}, errors.New("agent is required")
	}

	serverName := options.ServerName
	if serverName == "" {
		serverName = defaultServerName
	}
	if !serverNamePattern.MatchString(serverName) {
		return Result{}, fmt.Errorf("server name %q must contain only letters, numbers, underscores, and hyphens", serverName)
	}

	command := options.Command
	if command == "" {
		command = DefaultCommand()
	}

	home, err := homeDir(options.HomeDir)
	if err != nil {
		return Result{}, err
	}
	workDir, err := workDir(options.WorkDir)
	if err != nil {
		return Result{}, err
	}

	var path string
	var display string
	switch options.Agent {
	case AgentCodex:
		display = "Codex"
		path = codexConfigPath(home, workDir, options.Project)
		err = installCodex(path, serverName, command)
	case AgentClaudeCode:
		display = "Claude Code"
		path = claudeCodeConfigPath(home, workDir, options.Project)
		err = installClaudeCode(path, serverName, command)
	default:
		err = fmt.Errorf("unsupported agent %q", options.Agent)
	}
	if err != nil {
		return Result{}, err
	}

	return Result{
		Agent:      options.Agent,
		Display:    display,
		Path:       path,
		ServerName: serverName,
		Command:    command,
		Project:    options.Project,
	}, nil
}

func homeDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return home, nil
}

func workDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	return cwd, nil
}

func codexConfigPath(home, workDir string, project bool) string {
	if project {
		return filepath.Join(workDir, ".codex", "config.toml")
	}
	if codexHome := os.Getenv("CODEX_HOME"); codexHome != "" {
		return filepath.Join(codexHome, "config.toml")
	}
	return filepath.Join(home, ".codex", "config.toml")
}

func claudeCodeConfigPath(home, workDir string, project bool) string {
	if project {
		return filepath.Join(workDir, ".mcp.json")
	}
	return filepath.Join(home, ".claude.json")
}

func installCodex(path, serverName, command string) error {
	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	tableName := "mcp_servers." + serverName
	body := fmt.Sprintf("[%s]\ncommand = %q\nargs = []\n", tableName, command)
	next := upsertTomlTable(content, tableName, body)
	return writeFile(path, []byte(next))
}

func upsertTomlTable(content, tableName, body string) string {
	lines := strings.Split(content, "\n")
	var kept []string
	skipping := false

	for _, line := range lines {
		if name, ok := tomlTableName(line); ok {
			if name == tableName || strings.HasPrefix(name, tableName+".") {
				skipping = true
				continue
			}
			skipping = false
		}
		if !skipping {
			kept = append(kept, line)
		}
	}

	next := strings.TrimRight(strings.Join(kept, "\n"), "\n")
	if strings.TrimSpace(next) != "" {
		next += "\n\n"
	}
	return next + strings.TrimRight(body, "\n") + "\n"
}

func tomlTableName(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
		return "", false
	}
	if strings.HasPrefix(trimmed, "[[") {
		return "", false
	}
	name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
	if name == "" {
		return "", false
	}
	return name, true
}

func installClaudeCode(path, serverName, command string) error {
	config := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if len(strings.TrimSpace(string(data))) > 0 {
			if err := json.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[serverName] = map[string]any{
		"command": command,
		"args":    []string{},
	}
	config["mcpServers"] = servers

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFile(path, data)
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
