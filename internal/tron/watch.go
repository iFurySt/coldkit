package tron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultTronFullNodeEndpoint = "https://api.trongrid.io"
	fullNodeAttemptTimeout      = 5 * time.Second
	usdtTRC20Contract           = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

const (
	NetworkMainnet = "mainnet"
	NetworkNile    = "nile"
	NetworkShasta  = "shasta"
)

var DefaultTronFullNodeEndpoints = []string{
	DefaultTronFullNodeEndpoint,
	"http://3.225.171.164:8090",
	"http://18.133.82.227:8090",
	"http://15.207.144.3:8090",
	"http://15.222.19.181:8090",
}

var NileFullNodeEndpoints = []string{
	"https://api.nileex.io",
	"https://nile.trongrid.io",
}

var ShastaFullNodeEndpoints = []string{
	"https://api.shasta.trongrid.io",
}

type Balance struct {
	Address   string    `json:"address"`
	Active    bool      `json:"active"`
	TRX       string    `json:"trx"`
	USDT      string    `json:"usdt"`
	Resources Resources `json:"resources"`
}

type Resources struct {
	Address          string        `json:"address"`
	FreeBandwidth    ResourceUsage `json:"free_bandwidth"`
	StakedBandwidth  ResourceUsage `json:"staked_bandwidth"`
	TotalBandwidth   ResourceUsage `json:"total_bandwidth"`
	Energy           ResourceUsage `json:"energy"`
	TronPower        ResourceUsage `json:"tron_power"`
	NetworkBandwidth NetworkTotal  `json:"network_bandwidth"`
	NetworkEnergy    NetworkTotal  `json:"network_energy"`
}

type ResourceUsage struct {
	Used      int64 `json:"used"`
	Limit     int64 `json:"limit"`
	Remaining int64 `json:"remaining"`
}

type NetworkTotal struct {
	Limit  int64 `json:"limit"`
	Weight int64 `json:"weight"`
}

type accountResponse struct {
	Address string      `json:"address"`
	Balance json.Number `json:"balance"`
}

type resourceResponse struct {
	FreeNetUsed          int64 `json:"freeNetUsed"`
	FreeNetLimit         int64 `json:"freeNetLimit"`
	NetUsed              int64 `json:"NetUsed"`
	NetLimit             int64 `json:"NetLimit"`
	TotalNetLimit        int64 `json:"TotalNetLimit"`
	TotalNetWeight       int64 `json:"TotalNetWeight"`
	EnergyUsed           int64 `json:"EnergyUsed"`
	EnergyLimit          int64 `json:"EnergyLimit"`
	TotalEnergyLimit     int64 `json:"TotalEnergyLimit"`
	TotalEnergyWeight    int64 `json:"TotalEnergyWeight"`
	TronPowerUsed        int64 `json:"tronPowerUsed"`
	TronPowerLimit       int64 `json:"tronPowerLimit"`
	TotalTronPowerWeight int64 `json:"totalTronPowerWeight"`
}

type triggerConstantResponse struct {
	ConstantResult []string `json:"constant_result"`
	Result         struct {
		Result bool `json:"result"`
	} `json:"result"`
}

func FetchBalance(ctx context.Context, client *http.Client, endpoint string, address string) (Balance, error) {
	return FetchBalanceWithResources(ctx, client, endpointList(endpoint), address)
}

func FetchBalanceOnNetwork(ctx context.Context, client *http.Client, network string, endpoints []string, address string) (Balance, error) {
	resolved, err := ResolveFullNodeEndpoints(network, endpoints)
	if err != nil {
		return Balance{}, err
	}
	return FetchBalanceWithResources(ctx, client, resolved, address)
}

func FetchBalanceWithResources(ctx context.Context, client *http.Client, endpoints []string, address string) (Balance, error) {
	validated, err := ValidateAddress(address)
	if err != nil {
		return Balance{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	endpoints = normalizedEndpoints(endpoints)
	resources, err := FetchResourcesWithEndpoints(ctx, client, endpoints, validated.AddressBase58)
	if err != nil {
		return Balance{}, err
	}
	balance := Balance{Address: validated.AddressBase58, TRX: "0", USDT: "0", Resources: resources}

	account, active, err := fetchAccount(ctx, client, endpoints, validated.AddressBase58)
	if err != nil {
		return Balance{}, err
	}
	balance.Active = active
	if active && account.Balance != "" {
		balance.TRX = formatTokenAmount(account.Balance.String(), 6)
	}
	usdt, err := fetchTRC20Balance(ctx, client, endpoints, validated.AddressBase58, usdtTRC20Contract, 6)
	if err != nil {
		return Balance{}, err
	}
	balance.USDT = usdt
	return balance, nil
}

func FetchResources(ctx context.Context, client *http.Client, endpoint string, address string) (Resources, error) {
	return FetchResourcesWithEndpoints(ctx, client, endpointList(endpoint), address)
}

func FetchResourcesOnNetwork(ctx context.Context, client *http.Client, network string, endpoints []string, address string) (Resources, error) {
	resolved, err := ResolveFullNodeEndpoints(network, endpoints)
	if err != nil {
		return Resources{}, err
	}
	return FetchResourcesWithEndpoints(ctx, client, resolved, address)
}

func FetchResourcesWithEndpoints(ctx context.Context, client *http.Client, endpoints []string, address string) (Resources, error) {
	validated, err := ValidateAddress(address)
	if err != nil {
		return Resources{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	endpoints = normalizedEndpoints(endpoints)

	body, err := json.Marshal(map[string]any{
		"address": validated.AddressBase58,
		"visible": true,
	})
	if err != nil {
		return Resources{}, err
	}

	var payload resourceResponse
	if err := postWithFallback(ctx, client, endpoints, "/wallet/getaccountresource", body, &payload); err != nil {
		return Resources{}, err
	}

	freeBandwidth := usage(payload.FreeNetUsed, payload.FreeNetLimit)
	stakedBandwidth := usage(payload.NetUsed, payload.NetLimit)
	return Resources{
		Address:          validated.AddressBase58,
		FreeBandwidth:    freeBandwidth,
		StakedBandwidth:  stakedBandwidth,
		TotalBandwidth:   usage(freeBandwidth.Used+stakedBandwidth.Used, freeBandwidth.Limit+stakedBandwidth.Limit),
		Energy:           usage(payload.EnergyUsed, payload.EnergyLimit),
		TronPower:        usage(payload.TronPowerUsed, payload.TronPowerLimit),
		NetworkBandwidth: NetworkTotal{Limit: payload.TotalNetLimit, Weight: payload.TotalNetWeight},
		NetworkEnergy:    NetworkTotal{Limit: payload.TotalEnergyLimit, Weight: payload.TotalEnergyWeight},
	}, nil
}

func fetchAccount(ctx context.Context, client *http.Client, endpoints []string, address string) (accountResponse, bool, error) {
	body, err := json.Marshal(map[string]any{
		"address": address,
		"visible": true,
	})
	if err != nil {
		return accountResponse{}, false, err
	}

	var payload accountResponse
	if err := postWithFallback(ctx, client, endpoints, "/wallet/getaccount", body, &payload); err != nil {
		return accountResponse{}, false, err
	}
	return payload, payload.Address != "" || payload.Balance != "", nil
}

func fetchTRC20Balance(ctx context.Context, client *http.Client, endpoints []string, ownerAddress string, contractAddress string, decimals int) (string, error) {
	owner, err := ValidateAddress(ownerAddress)
	if err != nil {
		return "", err
	}
	contract, err := ValidateAddress(contractAddress)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(map[string]any{
		"owner_address":     owner.AddressBase58,
		"contract_address":  contract.AddressBase58,
		"function_selector": "balanceOf(address)",
		"parameter":         abiAddressParameter(owner.AddressHex),
		"visible":           true,
	})
	if err != nil {
		return "", err
	}

	var payload triggerConstantResponse
	if err := postWithFallback(ctx, client, endpoints, "/wallet/triggerconstantcontract", body, &payload); err != nil {
		return "", err
	}
	if len(payload.ConstantResult) == 0 || payload.ConstantResult[0] == "" {
		return "0", nil
	}
	return formatHexTokenAmount(payload.ConstantResult[0], decimals), nil
}

func postWithFallback(ctx context.Context, client *http.Client, endpoints []string, path string, body []byte, out any) error {
	var failures []string
	for _, endpoint := range endpoints {
		err := postJSON(ctx, client, endpoint, path, body, out)
		if err == nil {
			return nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", endpoint, err))
		if !isFallbackError(err) {
			return err
		}
	}
	return fmt.Errorf("all TRON full node endpoints failed: %s", strings.Join(failures, "; "))
}

func postJSON(ctx context.Context, client *http.Client, endpoint string, path string, body []byte, out any) error {
	attemptCtx, cancel := context.WithTimeout(ctx, fullNodeAttemptTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, strings.TrimRight(endpoint, "/")+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "coldkit-watch-only/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return httpStatusError(resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	return decoder.Decode(out)
}

type httpStatusError int

func (e httpStatusError) Error() string {
	return fmt.Sprintf("HTTP %d", int(e))
}

func isFallbackError(err error) bool {
	status, ok := err.(httpStatusError)
	if !ok {
		return true
	}
	return status == http.StatusTooManyRequests || status >= 500
}

func endpointList(endpoint string) []string {
	if strings.TrimSpace(endpoint) == "" {
		return nil
	}
	return []string{endpoint}
}

func normalizedEndpoints(endpoints []string) []string {
	var normalized []string
	for _, endpoint := range endpoints {
		endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
		if endpoint != "" {
			normalized = append(normalized, endpoint)
		}
	}
	if len(normalized) == 0 {
		return append([]string(nil), DefaultTronFullNodeEndpoints...)
	}
	return normalized
}

func ResolveFullNodeEndpoints(network string, endpoints []string) ([]string, error) {
	network = strings.ToLower(strings.TrimSpace(network))
	if network == "" || network == "main" {
		network = NetworkMainnet
	}
	if custom := normalizedCustomEndpoints(endpoints); len(custom) > 0 {
		if _, err := defaultEndpointsForNetwork(network); err != nil {
			return nil, err
		}
		return custom, nil
	}
	defaults, err := defaultEndpointsForNetwork(network)
	if err != nil {
		return nil, err
	}
	return append([]string(nil), defaults...), nil
}

func defaultEndpointsForNetwork(network string) ([]string, error) {
	switch network {
	case NetworkMainnet:
		return DefaultTronFullNodeEndpoints, nil
	case NetworkNile:
		return NileFullNodeEndpoints, nil
	case NetworkShasta:
		return ShastaFullNodeEndpoints, nil
	default:
		return nil, fmt.Errorf("unsupported TRON network %q; expected mainnet, nile, or shasta", network)
	}
}

func normalizedCustomEndpoints(endpoints []string) []string {
	var normalized []string
	for _, endpoint := range endpoints {
		endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
		if endpoint != "" {
			normalized = append(normalized, endpoint)
		}
	}
	return normalized
}

func abiAddressParameter(addressHex string) string {
	addressHex = strings.TrimPrefix(strings.ToLower(addressHex), "41")
	return strings.Repeat("0", 64-len(addressHex)) + addressHex
}

func usage(used int64, limit int64) ResourceUsage {
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	return ResourceUsage{Used: used, Limit: limit, Remaining: remaining}
}

func formatTokenAmount(raw string, decimals int) string {
	value, ok := new(big.Int).SetString(raw, 10)
	if !ok || value.Sign() == 0 {
		return "0"
	}
	return formatBigTokenAmount(value, decimals)
}

func formatHexTokenAmount(raw string, decimals int) string {
	value, ok := new(big.Int).SetString(strings.TrimPrefix(raw, "0x"), 16)
	if !ok || value.Sign() == 0 {
		return "0"
	}
	return formatBigTokenAmount(value, decimals)
}

func formatBigTokenAmount(value *big.Int, decimals int) string {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	whole := new(big.Int).Div(new(big.Int).Set(value), divisor)
	fraction := new(big.Int).Mod(value, divisor).Text(10)
	for len(fraction) < decimals {
		fraction = "0" + fraction
	}
	fraction = strings.TrimRight(fraction, "0")
	if fraction == "" {
		return whole.Text(10)
	}
	return whole.Text(10) + "." + fraction
}
