package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ifuryst/coldkit/internal/mcpinstall"
	"github.com/ifuryst/coldkit/internal/tron"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:          "ck",
		Short:        "coldkit wallet safety tools",
		SilenceUsage: true,
	}
	root.AddCommand(newTronCommand())
	root.AddCommand(newAddMCPCommand())
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
		Example: "  ck add-mcp codex\n  ck add-mcp claude-code\n  ck add-mcp cloud-code --project",
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
	cmd.AddCommand(newTronSelfCommand())
	return cmd
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
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
