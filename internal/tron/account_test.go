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
	var gotAccountPath string
	var gotResourcePath string
	var gotTRC20Path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/wallet/getaccount":
			gotAccountPath = r.URL.Path
			_, _ = w.Write([]byte(`{"address":"TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3","balance":638007}`))
		case "/wallet/getaccountresource":
			gotResourcePath = r.URL.Path
			_, _ = w.Write([]byte(`{"freeNetUsed":125,"freeNetLimit":600,"NetUsed":20,"NetLimit":1000,"EnergyUsed":30,"EnergyLimit":5000,"TotalNetLimit":43200000000,"TotalNetWeight":99,"TotalEnergyLimit":180000000000,"TotalEnergyWeight":88}`))
		case "/wallet/triggerconstantcontract":
			gotTRC20Path = r.URL.Path
			_, _ = w.Write([]byte(`{"result":{"result":true},"constant_result":["0000000000000000000000000000000000000000000000000000000000bc614e"]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	balance, err := FetchBalanceWithResources(context.Background(), server.Client(), []string{server.URL}, "TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3")
	if err != nil {
		t.Fatal(err)
	}
	if gotAccountPath != "/wallet/getaccount" {
		t.Fatalf("account path = %s", gotAccountPath)
	}
	if gotResourcePath != "/wallet/getaccountresource" {
		t.Fatalf("resource path = %s", gotResourcePath)
	}
	if gotTRC20Path != "/wallet/triggerconstantcontract" {
		t.Fatalf("trc20 path = %s", gotTRC20Path)
	}
	if !balance.Active || balance.TRX != "0.638007" || balance.USDT != "12.345678" {
		t.Fatalf("balance = %+v", balance)
	}
	if balance.Resources.FreeBandwidth.Remaining != 475 || balance.Resources.StakedBandwidth.Remaining != 980 || balance.Resources.Energy.Remaining != 4970 {
		t.Fatalf("resources = %+v", balance.Resources)
	}
}

func TestFetchResourcesWatchOnly(t *testing.T) {
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"freeNetUsed":1000,"freeNetLimit":600,"NetUsed":1,"NetLimit":3,"EnergyUsed":2,"EnergyLimit":5,"tronPowerUsed":4,"tronPowerLimit":9}`))
	}))
	defer server.Close()

	resources, err := FetchResources(context.Background(), server.Client(), server.URL, "TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s", gotMethod)
	}
	if resources.FreeBandwidth.Remaining != 0 || resources.StakedBandwidth.Remaining != 2 || resources.TotalBandwidth.Remaining != 0 || resources.Energy.Remaining != 3 || resources.TronPower.Remaining != 5 {
		t.Fatalf("resources = %+v", resources)
	}
}

func TestFetchBalanceFallsBackAcrossEndpoints(t *testing.T) {
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer failing.Close()
	working := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/wallet/getaccount":
			_, _ = w.Write([]byte(`{"address":"TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3","balance":1}`))
		case "/wallet/getaccountresource":
			_, _ = w.Write([]byte(`{"freeNetLimit":600}`))
		case "/wallet/triggerconstantcontract":
			_, _ = w.Write([]byte(`{"result":{"result":true},"constant_result":["0"]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer working.Close()

	balance, err := FetchBalanceWithResources(context.Background(), working.Client(), []string{failing.URL, working.URL}, "TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3")

	if err != nil {
		t.Fatal(err)
	}
	if balance.TRX != "0.000001" || balance.Resources.FreeBandwidth.Limit != 600 {
		t.Fatalf("balance = %+v", balance)
	}
}

func TestResolveFullNodeEndpointsByNetwork(t *testing.T) {
	tests := []struct {
		network string
		want    string
	}{
		{network: "", want: DefaultTronFullNodeEndpoint},
		{network: "main", want: DefaultTronFullNodeEndpoint},
		{network: "mainnet", want: DefaultTronFullNodeEndpoint},
		{network: "nile", want: "https://api.nileex.io"},
		{network: "shasta", want: "https://api.shasta.trongrid.io"},
	}
	for _, tt := range tests {
		t.Run(tt.network, func(t *testing.T) {
			endpoints, err := ResolveFullNodeEndpoints(tt.network, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(endpoints) == 0 || endpoints[0] != tt.want {
				t.Fatalf("endpoints = %#v", endpoints)
			}
		})
	}
}

func TestResolveFullNodeEndpointsCustomOverridesNetworkDefaults(t *testing.T) {
	endpoints, err := ResolveFullNodeEndpoints(NetworkNile, []string{" https://example.test/ ", ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(endpoints) != 1 || endpoints[0] != "https://example.test" {
		t.Fatalf("endpoints = %#v", endpoints)
	}
}

func TestResolveFullNodeEndpointsRejectsUnknownNetwork(t *testing.T) {
	_, err := ResolveFullNodeEndpoints("unknown", nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported TRON network") {
		t.Fatalf("err = %v", err)
	}
}
