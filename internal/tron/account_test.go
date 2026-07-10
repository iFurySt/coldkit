package tron

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSelfTest(t *testing.T) {
	if err := SelfTest(); err != nil {
		t.Fatal(err)
	}
}

func TestAccountFromPrivateKey(t *testing.T) {
	account, err := AccountFromPrivateKey(strings.Repeat("88", 32))
	if err != nil {
		t.Fatal(err)
	}
	if account.AddressBase58 != "TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3" {
		t.Fatalf("address = %s", account.AddressBase58)
	}
	if account.AddressHex != "4162f94e9ac9349bccc61bfe66ddade6292702ecb6" {
		t.Fatalf("hex address = %s", account.AddressHex)
	}
}

func TestValidateAddress(t *testing.T) {
	address, err := ValidateAddress("TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3")
	if err != nil {
		t.Fatal(err)
	}
	if address.AddressHex != "4162f94e9ac9349bccc61bfe66ddade6292702ecb6" {
		t.Fatalf("hex address = %s", address.AddressHex)
	}
}

func TestGenerateAccountsWithMultipleSuffixes(t *testing.T) {
	results, err := GenerateAccounts(GenerateOptions{Suffixes: []string{"2", "3"}, Count: 2, MaxAttempts: 10_000})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d", len(results))
	}
	for _, result := range results {
		if result.MatchedSuffix != "2" && result.MatchedSuffix != "3" {
			t.Fatalf("matched suffix = %q", result.MatchedSuffix)
		}
		if !strings.HasSuffix(result.AddressBase58, result.MatchedSuffix) {
			t.Fatalf("address %s does not have suffix %s", result.AddressBase58, result.MatchedSuffix)
		}
	}
}

func TestGenerateAccountsWithPrefixAndSuffix(t *testing.T) {
	results, err := GenerateAccounts(GenerateOptions{Prefixes: []string{"T"}, Suffixes: []string{"2", "3"}, Count: 1, MaxAttempts: 10_000})
	if err != nil {
		t.Fatal(err)
	}
	result := results[0]
	if result.MatchedPrefix != "T" {
		t.Fatalf("matched prefix = %q", result.MatchedPrefix)
	}
	if result.MatchedSuffix != "2" && result.MatchedSuffix != "3" {
		t.Fatalf("matched suffix = %q", result.MatchedSuffix)
	}
}

func TestGenerateAccountsRejectsInvalidBase58(t *testing.T) {
	_, err := GenerateAccounts(GenerateOptions{Suffixes: []string{"888", "0"}, Count: 1, MaxAttempts: 1})
	if err == nil {
		t.Fatal("expected invalid suffix error")
	}
}

func TestGenerateAccountsRejectsZeroCount(t *testing.T) {
	_, err := GenerateAccounts(GenerateOptions{})
	if err == nil {
		t.Fatal("expected count error")
	}
}

func TestFetchBalanceWatchOnly(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"balance":638007,"trc20":[{"TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t":"12345678"}]}]}`))
	}))
	defer server.Close()

	balance, err := FetchBalance(context.Background(), server.Client(), server.URL, "TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3" {
		t.Fatalf("path = %s", gotPath)
	}
	if !balance.Active || balance.TRX != "0.638007" || balance.USDT != "12.345678" {
		t.Fatalf("balance = %+v", balance)
	}
}
