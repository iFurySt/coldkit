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
	DefaultTronGridAccountsEndpoint = "https://api.trongrid.io/v1/accounts"
	DefaultTronGridResourceEndpoint = "https://api.trongrid.io/wallet/getaccountresource"
	usdtTRC20Contract               = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

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

type tronGridResponse struct {
	Data []tronGridAccount `json:"data"`
}

type tronGridAccount struct {
	Balance json.Number         `json:"balance"`
	TRC20   []map[string]string `json:"trc20"`
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

func FetchBalance(ctx context.Context, client *http.Client, endpoint string, address string) (Balance, error) {
	return FetchBalanceWithResources(ctx, client, endpoint, DefaultTronGridResourceEndpoint, address)
}

func FetchBalanceWithResources(ctx context.Context, client *http.Client, accountsEndpoint string, resourceEndpoint string, address string) (Balance, error) {
	validated, err := ValidateAddress(address)
	if err != nil {
		return Balance{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	accountsEndpoint = strings.TrimRight(accountsEndpoint, "/")
	if accountsEndpoint == "" {
		accountsEndpoint = DefaultTronGridAccountsEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accountsEndpoint+"/"+validated.AddressBase58, nil)
	if err != nil {
		return Balance{}, err
	}
	req.Header.Set("User-Agent", "coldkit-watch-only/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return Balance{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Balance{}, fmt.Errorf("TronGrid returned HTTP %d", resp.StatusCode)
	}

	var payload tronGridResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return Balance{}, err
	}

	resources, err := FetchResources(ctx, client, resourceEndpoint, validated.AddressBase58)
	if err != nil {
		return Balance{}, err
	}

	balance := Balance{Address: validated.AddressBase58, TRX: "0", USDT: "0", Resources: resources}
	if len(payload.Data) == 0 {
		return balance, nil
	}
	account := payload.Data[0]
	balance.Active = true
	if account.Balance != "" {
		balance.TRX = formatTokenAmount(account.Balance.String(), 6)
	}
	for _, token := range account.TRC20 {
		if value, ok := token[usdtTRC20Contract]; ok {
			balance.USDT = formatTokenAmount(value, 6)
			break
		}
	}
	return balance, nil
}

func FetchResources(ctx context.Context, client *http.Client, endpoint string, address string) (Resources, error) {
	validated, err := ValidateAddress(address)
	if err != nil {
		return Resources{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		endpoint = DefaultTronGridResourceEndpoint
	}

	body, err := json.Marshal(map[string]any{
		"address": validated.AddressBase58,
		"visible": true,
	})
	if err != nil {
		return Resources{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Resources{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "coldkit-watch-only/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return Resources{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Resources{}, fmt.Errorf("TronGrid returned HTTP %d", resp.StatusCode)
	}

	var payload resourceResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
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
