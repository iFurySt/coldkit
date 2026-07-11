package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ifuryst/coldkit/internal/keychain"
	"github.com/ifuryst/coldkit/internal/mcpinstall"
	"github.com/ifuryst/coldkit/internal/tron"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:          "ck",
		Short:        "coldkit wallet safety tools",
		SilenceUsage: true,
	}
	root.AddCommand(newTronCommand())
	root.AddCommand(newAddMCPCommand())
	root.AddCommand(newKeychainCommand())
	root.AddCommand(newSelfCommand())
	return root
}

func newAddMCPCommand() *cobra.Command {
	var project bool
	var command string
	var name string
	cmd := &cobra.Command{
		Use:     "add-mcp AGENT",
		Aliases: []string{"install-mcp"},
		Short:   "install the coldkit MCP server for an agent",
		Example: "  ck add-mcp codex\n  ck add-mcp claude-code\n  ck add-mcp claude-code --project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, err := mcpinstall.ResolveAgent(args[0])
			if err != nil {
				return err
			}
			result, err := mcpinstall.Install(mcpinstall.Options{
				Agent:      agent,
				Project:    project,
				ServerName: name,
				Command:    command,
			})
			if err != nil {
				return err
			}

			scope := "global"
			if result.Project {
				scope = "project"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed %s MCP server for %s (%s)\n", result.ServerName, result.Display, scope)
			fmt.Fprintf(cmd.OutOrStdout(), "Config:  %s\n", result.Path)
			fmt.Fprintf(cmd.OutOrStdout(), "Command: %s\n", result.Command)
			return nil
		},
	}
	cmd.Flags().BoolVar(&project, "project", false, "install to project-level config instead of user config")
	cmd.Flags().StringVar(&command, "command", "", "MCP server command to write; defaults to a sibling ck-mcp binary when available")
	cmd.Flags().StringVar(&name, "name", "coldkit", "MCP server name")
	return cmd
}

func newTronCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tron",
		Short:   "TRON and USDT/TRC20 tools",
		Aliases: []string{"trx"},
	}
	cmd.AddCommand(newTronGenCommand())
	cmd.AddCommand(newTronFromPrivateCommand())
	cmd.AddCommand(newTronValidateCommand())
	cmd.AddCommand(newTronBalanceCommand())
	cmd.AddCommand(newTronSignHashCommand())
	cmd.AddCommand(newTronSelfCommand())
	return cmd
}

func newKeychainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keychain",
		Short: "store local signing keys in macOS Keychain",
	}
	cmd.AddCommand(newKeychainImportTronCommand())
	cmd.AddCommand(newKeychainShowTronCommand())
	cmd.AddCommand(newKeychainDeleteCommand())
	return cmd
}

func newKeychainImportTronCommand() *cobra.Command {
	var asJSON bool
	var privateKeyStdin bool
	cmd := &cobra.Command{
		Use:   "import-tron NAME",
		Short: "import a TRON private key into macOS Keychain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey, err := readPrivateKey(cmd, privateKeyStdin)
			if err != nil {
				return err
			}
			key, err := keychain.ImportTronPrivateKey(args[0], privateKey)
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSONTo(cmd.OutOrStdout(), key)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Stored TRON key:   %s\n", key.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "TRON address:      %s\n", key.AddressBase58)
			fmt.Fprintf(cmd.OutOrStdout(), "TRON hex address:  %s\n", key.AddressHex)
			fmt.Fprintf(cmd.OutOrStdout(), "Public key hex:    %s\n", key.PublicKeyHex)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	cmd.Flags().BoolVar(&privateKeyStdin, "private-key-stdin", false, "read the private key from stdin instead of prompting")
	return cmd
}

func newKeychainShowTronCommand() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show-tron NAME",
		Short: "show public metadata for a TRON key stored in macOS Keychain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := keychain.DescribeTronKey(args[0])
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSONTo(cmd.OutOrStdout(), key)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Stored TRON key:   %s\n", key.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "TRON address:      %s\n", key.AddressBase58)
			fmt.Fprintf(cmd.OutOrStdout(), "TRON hex address:  %s\n", key.AddressHex)
			fmt.Fprintf(cmd.OutOrStdout(), "Public key hex:    %s\n", key.PublicKeyHex)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	return cmd
}

func newKeychainDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "delete a coldkit key from macOS Keychain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := keychain.DeleteTronKey(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted key: %s\n", args[0])
			return nil
		},
	}
}

func newTronGenCommand() *cobra.Command {
	var prefixes []string
	var suffixes []string
	var count uint64
	var maxAttempts uint64
	var asJSON bool
	var publicOnly bool
	cmd := &cobra.Command{
		Use:     "gen",
		Aliases: []string{"g"},
		Short:   "generate TRON addresses offline",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := tron.GenerateAccounts(tron.GenerateOptions{
				Prefixes:    prefixes,
				Suffixes:    suffixes,
				Count:       count,
				MaxAttempts: maxAttempts,
			})
			if err != nil {
				return err
			}
			if publicOnly {
				return printValue(tron.PublicResults(results), asJSON)
			}
			return printValue(results, asJSON)
		},
	}
	cmd.Flags().StringArrayVarP(&prefixes, "prefix", "p", nil, "require generated Base58 addresses to start with this prefix; repeat for multiple prefixes")
	cmd.Flags().StringArrayVarP(&suffixes, "suffix", "s", nil, "require generated Base58 addresses to end with this suffix; repeat for multiple suffixes")
	cmd.Flags().Uint64VarP(&count, "count", "n", 1, "number of matching addresses to generate")
	cmd.Flags().Uint64Var(&maxAttempts, "max", 0, "maximum total random keys to try; 0 means unlimited")
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	cmd.Flags().BoolVar(&publicOnly, "pub", false, "omit private keys from output")
	return cmd
}

func newTronFromPrivateCommand() *cobra.Command {
	var asJSON bool
	var publicOnly bool
	cmd := &cobra.Command{
		Use:     "from-private PRIVATE_KEY_HEX",
		Aliases: []string{"derive"},
		Short:   "derive a TRON address from a private key offline",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			account, err := tron.AccountFromPrivateKey(args[0])
			if err != nil {
				return err
			}
			if publicOnly {
				return printValue(tron.PublicResults([]tron.GeneratedAccount{{Account: account}})[0].PublicAccount, asJSON)
			}
			return printValue(account, asJSON)
		},
	}
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	cmd.Flags().BoolVar(&publicOnly, "pub", false, "omit private key from output")
	return cmd
}

func newTronValidateCommand() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "val ADDRESS",
		Aliases: []string{"validate", "v"},
		Short:   "validate a public TRON address offline",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := tron.ValidateAddress(args[0])
			if err != nil {
				return err
			}
			return printValue(address, asJSON)
		},
	}
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	return cmd
}

func newTronBalanceCommand() *cobra.Command {
	var asJSON bool
	var endpoint string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:     "bal ADDRESS",
		Aliases: []string{"balance", "b"},
		Short:   "check TRX and USDT/TRC20 balances for a public address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			balance, err := tron.FetchBalance(ctx, &http.Client{Timeout: timeout}, endpoint, args[0])
			if err != nil {
				return err
			}
			return printValue(balance, asJSON)
		},
	}
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	cmd.Flags().StringVar(&endpoint, "endpoint", tron.DefaultTronGridAccountsEndpoint, "TRON accounts API endpoint")
	cmd.Flags().DurationVar(&timeout, "timeout", 20*time.Second, "HTTP timeout")
	return cmd
}

func newTronSignHashCommand() *cobra.Command {
	var keyName string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "sign-hash DIGEST_HEX",
		Short: "sign a 32-byte digest with a TRON key from macOS Keychain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if keyName == "" {
				return fmt.Errorf("--key is required")
			}
			privateKey, err := keychain.LoadTronPrivateKey(keyName)
			if err != nil {
				return err
			}
			signature, err := tron.SignDigest(privateKey, args[0])
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSONTo(cmd.OutOrStdout(), signature)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "TRON address:      %s\n", signature.AddressBase58)
			fmt.Fprintf(cmd.OutOrStdout(), "Digest hex:        %s\n", signature.DigestHex)
			fmt.Fprintf(cmd.OutOrStdout(), "Signature hex:     %s\n", signature.SignatureHex)
			fmt.Fprintf(cmd.OutOrStdout(), "Recovery ID:       %d\n", signature.RecoveryID)
			fmt.Fprintf(cmd.OutOrStdout(), "DER signature hex: %s\n", signature.DERSignatureHex)
			return nil
		},
	}
	cmd.Flags().StringVar(&keyName, "key", "", "macOS Keychain key name")
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "print JSON")
	return cmd
}

func newTronSelfCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "self",
		Aliases: []string{"self-test"},
		Short:   "run TRON deterministic test vectors",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tron.SelfTest(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "tron self-test ok")
			return nil
		},
	}
}

func newSelfCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "self",
		Short: "run coldkit deterministic test vectors",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tron.SelfTest(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "self-test ok")
			return nil
		},
	}
}

func printValue(value any, asJSON bool) error {
	if asJSON {
		return writeJSON(value)
	}
	switch v := value.(type) {
	case []tron.GeneratedAccount:
		for index, result := range v {
			if index > 0 {
				fmt.Println()
			}
			printAccount(result.Account)
			if result.MatchedPrefix != "" {
				fmt.Printf("Matched prefix:    %s\n", result.MatchedPrefix)
			}
			if result.MatchedSuffix != "" {
				fmt.Printf("Matched suffix:    %s\n", result.MatchedSuffix)
			}
			fmt.Printf("Attempts:          %d\n", result.Attempts)
		}
	case []tron.GeneratedPublicAccount:
		for index, result := range v {
			if index > 0 {
				fmt.Println()
			}
			printPublicAccount(result.PublicAccount)
			if result.MatchedPrefix != "" {
				fmt.Printf("Matched prefix:    %s\n", result.MatchedPrefix)
			}
			if result.MatchedSuffix != "" {
				fmt.Printf("Matched suffix:    %s\n", result.MatchedSuffix)
			}
			fmt.Printf("Attempts:          %d\n", result.Attempts)
		}
	case tron.Account:
		printAccount(v)
	case tron.PublicAccount:
		printPublicAccount(v)
	case tron.Address:
		fmt.Printf("TRON address:      %s\n", v.AddressBase58)
		fmt.Printf("TRON hex address:  %s\n", v.AddressHex)
	case tron.Balance:
		status := "not activated or no account data returned"
		if v.Active {
			status = "active"
		}
		fmt.Printf("Address:           %s\n", v.Address)
		fmt.Printf("Status:            %s\n", status)
		fmt.Printf("TRX:               %s\n", v.TRX)
		fmt.Printf("USDT:              %s\n", v.USDT)
	default:
		return writeJSON(value)
	}
	return nil
}

func printAccount(account tron.Account) {
	fmt.Printf("TRON address:      %s\n", account.AddressBase58)
	fmt.Printf("TRON hex address:  %s\n", account.AddressHex)
	fmt.Printf("Private key hex:   %s\n", account.PrivateKeyHex)
	fmt.Printf("Public key hex:    %s\n", account.PublicKeyHex)
}

func printPublicAccount(account tron.PublicAccount) {
	fmt.Printf("TRON address:      %s\n", account.AddressBase58)
	fmt.Printf("TRON hex address:  %s\n", account.AddressHex)
	fmt.Printf("Public key hex:    %s\n", account.PublicKeyHex)
}

func writeJSON(value any) error {
	return writeJSONTo(os.Stdout, value)
}

func writeJSONTo(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func readPrivateKey(cmd *cobra.Command, fromStdin bool) (string, error) {
	if fromStdin {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return "", fmt.Errorf("read private key from stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	stdin, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(stdin.Fd())) {
		return "", fmt.Errorf("refusing to read a private key from non-interactive stdin; pass --private-key-stdin explicitly")
	}
	fmt.Fprint(cmd.ErrOrStderr(), "Private key hex: ")
	data, err := term.ReadPassword(int(stdin.Fd()))
	fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", fmt.Errorf("read private key: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
