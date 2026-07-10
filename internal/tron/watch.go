package tron

import (
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
	usdtTRC20Contract               = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

type Balance struct {
	Address string `json:"address"`
	Active  bool   `json:"active"`
	TRX     string `json:"trx"`
	USDT    string `json:"usdt"`
}

type tronGridResponse struct {
	Data []tronGridAccount `json:"data"`
}

type tronGridAccount struct {
	Balance json.Number         `json:"balance"`
	TRC20   []map[string]string `json:"trc20"`
}

func FetchBalance(ctx context.Context, client *http.Client, endpoint string, address string) (Balance, error) {
	validated, err := ValidateAddress(address)
	if err != nil {
		return Balance{}, err
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		endpoint = DefaultTronGridAccountsEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/"+validated.AddressBase58, nil)
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

	balance := Balance{Address: validated.AddressBase58, TRX: "0", USDT: "0"}
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
