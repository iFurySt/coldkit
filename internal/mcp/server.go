package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ifuryst/coldkit/internal/keychain"
	"github.com/ifuryst/coldkit/internal/tron"
)

type Config struct {
	EnableSecretTools bool
}

type Server struct {
	config Config
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Result  any            `json:"result,omitempty"`
	Error   *responseError `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type genArgs struct {
	Prefixes    []string `json:"prefixes"`
	Suffixes    []string `json:"suffixes"`
	Count       uint64   `json:"count"`
	MaxAttempts uint64   `json:"max_attempts"`
}

type addressArgs struct {
	Address string `json:"address"`
}

type signHashArgs struct {
	KeyName   string `json:"key_name"`
	DigestHex string `json:"digest_hex"`
}

func NewServer(config Config) *Server {
	return &Server{config: config}
}

func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			if err := encoder.Encode(errorResponse(nil, -32700, "parse error")); err != nil {
				return err
			}
			continue
		}
		resp, ok := s.handle(ctx, req)
		if !ok {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, req request) (response, bool) {
	switch req.Method {
	case "initialize":
		return okResponse(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{"listChanged": false},
			},
			"serverInfo": map[string]any{
				"name":    "ck-mcp",
				"version": "0.1.0",
			},
		}), true
	case "notifications/initialized":
		return response{}, false
	case "tools/list":
		return okResponse(req.ID, map[string]any{"tools": s.tools()}), true
	case "tools/call":
		result, err := s.callTool(ctx, req.Params)
		if err != nil {
			return errorResponse(req.ID, -32602, err.Error()), true
		}
		return okResponse(req.ID, map[string]any{
			"content": []map[string]string{{
				"type": "text",
				"text": result,
			}},
		}), true
	default:
		return errorResponse(req.ID, -32601, "method not found"), true
	}
}

func (s *Server) tools() []tool {
	tools := []tool{
		{
			Name:        "tron_validate",
			Description: "Validate a public TRON Base58Check address offline.",
			InputSchema: objectSchema(map[string]any{
				"address": map[string]any{"type": "string"},
			}, []string{"address"}),
		},
		{
			Name:        "tron_balance",
			Description: "Check public TRX, USDT/TRC20, energy, and bandwidth for a TRON address. This tool performs network I/O and never accepts private keys.",
			InputSchema: objectSchema(map[string]any{
				"address": map[string]any{"type": "string"},
			}, []string{"address"}),
		},
		{
			Name:        "tron_resource",
			Description: "Check public TRON energy and bandwidth resources for a TRON address. This tool performs network I/O and never accepts private keys.",
			InputSchema: objectSchema(map[string]any{
				"address": map[string]any{"type": "string"},
			}, []string{"address"}),
		},
		{
			Name:        "tron_generate_preview",
			Description: "Generate TRON address previews without returning private keys. Generated addresses are not usable for funds unless secret tools are enabled and the matching private key is exported locally.",
			InputSchema: genSchema(),
		},
		{
			Name:        "tron_sign_hash",
			Description: "Sign a 32-byte digest with a TRON key stored in macOS Keychain. The private key is not returned; macOS may prompt the user to authorize access.",
			InputSchema: objectSchema(map[string]any{
				"key_name":   map[string]any{"type": "string"},
				"digest_hex": map[string]any{"type": "string", "description": "32-byte digest encoded as 64 hex characters"},
			}, []string{"key_name", "digest_hex"}),
		},
	}
	if s.config.EnableSecretTools {
		tools = append(tools, tool{
			Name:        "tron_generate_secret",
			Description: "DANGEROUS: generate TRON addresses and return private keys. Enable only on an offline machine.",
			InputSchema: genSchema(),
		})
	}
	return tools
}

func (s *Server) callTool(ctx context.Context, raw json.RawMessage) (string, error) {
	var params callParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return "", err
	}
	switch params.Name {
	case "tron_validate":
		var args addressArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return "", err
		}
		return jsonText(tron.ValidateAddress(args.Address))
	case "tron_balance":
		var args addressArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return "", err
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		return jsonText(tron.FetchBalance(timeoutCtx, &http.Client{Timeout: 20 * time.Second}, tron.DefaultTronGridAccountsEndpoint, args.Address))
	case "tron_resource":
		var args addressArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return "", err
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		return jsonText(tron.FetchResources(timeoutCtx, &http.Client{Timeout: 20 * time.Second}, tron.DefaultTronGridResourceEndpoint, args.Address))
	case "tron_generate_preview":
		var args genArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return "", err
		}
		results, err := generate(args)
		if err != nil {
			return "", err
		}
		return marshalText(tron.PublicResults(results))
	case "tron_generate_secret":
		if !s.config.EnableSecretTools {
			return "", fmt.Errorf("secret tools are disabled; restart ck-mcp with --enable-secret-tools on an offline machine")
		}
		var args genArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return "", err
		}
		results, err := generate(args)
		if err != nil {
			return "", err
		}
		return marshalText(results)
	case "tron_sign_hash":
		var args signHashArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			return "", err
		}
		privateKey, err := keychain.LoadTronPrivateKey(args.KeyName)
		if err != nil {
			return "", err
		}
		return jsonText(tron.SignDigest(privateKey, args.DigestHex))
	default:
		return "", fmt.Errorf("unknown tool %q", params.Name)
	}
}

func generate(args genArgs) ([]tron.GeneratedAccount, error) {
	count := args.Count
	if count == 0 {
		count = 1
	}
	return tron.GenerateAccounts(tron.GenerateOptions{
		Prefixes:    args.Prefixes,
		Suffixes:    args.Suffixes,
		Count:       count,
		MaxAttempts: args.MaxAttempts,
	})
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func genSchema() map[string]any {
	return objectSchema(map[string]any{
		"prefixes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"suffixes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"count": map[string]any{
			"type":        "integer",
			"minimum":     1,
			"description": "number of matching addresses to generate; defaults to 1",
		},
		"max_attempts": map[string]any{
			"type":        "integer",
			"minimum":     0,
			"description": "maximum total random keys to try; 0 means unlimited",
		},
	}, nil)
}

func jsonText[T any](value T, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return marshalText(value)
}

func marshalText(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func okResponse(id any, result any) response {
	return response{JSONRPC: "2.0", ID: id, Result: result}
}

func errorResponse(id any, code int, message string) response {
	return response{JSONRPC: "2.0", ID: id, Error: &responseError{Code: code, Message: message}}
}
