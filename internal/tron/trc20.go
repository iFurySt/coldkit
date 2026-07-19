package tron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
)

const (
	USDTTRC20Contract       = usdtTRC20Contract
	TRC20TransferSelector   = "transfer(address,uint256)"
	TRC20TransferMethodID   = "a9059cbb"
	defaultTRC20TokenSymbol = "USDT"
)

type TRC20TransferPreview struct {
	Token            string               `json:"token"`
	ContractAddress  string               `json:"contract_address"`
	ContractHex      string               `json:"contract_hex"`
	ToAddress        string               `json:"to_address"`
	ToHex            string               `json:"to_hex"`
	Amount           string               `json:"amount"`
	AmountRaw        string               `json:"amount_raw"`
	Decimals         int                  `json:"decimals"`
	FunctionSelector string               `json:"function_selector"`
	MethodID         string               `json:"method_id"`
	Parameter        string               `json:"parameter"`
	Data             string               `json:"data"`
	AddressWord      string               `json:"address_word"`
	AmountWord       string               `json:"amount_word"`
	DryRun           *TRC20TransferDryRun `json:"dry_run,omitempty"`
}

type TRC20TransferOptions struct {
	Token           string
	ContractAddress string
	ToAddress       string
	Amount          string
	Decimals        int
}

type TRC20TransferDryRun struct {
	OwnerAddress   string   `json:"owner_address"`
	OwnerHex       string   `json:"owner_hex"`
	OK             bool     `json:"ok"`
	ConstantResult []string `json:"constant_result,omitempty"`
	EnergyUsed     int64    `json:"energy_used,omitempty"`
	LogCount       int      `json:"log_count,omitempty"`
	Message        string   `json:"message,omitempty"`
}

type triggerConstantContractResponse struct {
	ConstantResult []string `json:"constant_result"`
	EnergyUsed     int64    `json:"energy_used"`
	Logs           []any    `json:"logs"`
	Result         struct {
		Result  bool   `json:"result"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"result"`
	Transaction struct {
		Ret []struct {
			Ret string `json:"ret"`
		} `json:"ret"`
	} `json:"transaction"`
}

func BuildTRC20TransferPreview(options TRC20TransferOptions) (TRC20TransferPreview, error) {
	token := strings.TrimSpace(options.Token)
	if token == "" {
		token = defaultTRC20TokenSymbol
	}
	contractAddress := strings.TrimSpace(options.ContractAddress)
	if contractAddress == "" {
		contractAddress = USDTTRC20Contract
	}
	decimals := options.Decimals
	if decimals < 0 || decimals > 30 {
		return TRC20TransferPreview{}, errors.New("decimals must be between 0 and 30")
	}

	contract, err := ValidateAddress(contractAddress)
	if err != nil {
		return TRC20TransferPreview{}, fmt.Errorf("contract address: %w", err)
	}
	to, err := ValidateAddress(options.ToAddress)
	if err != nil {
		return TRC20TransferPreview{}, fmt.Errorf("to address: %w", err)
	}
	amountRaw, err := parseDecimalTokenAmount(options.Amount, decimals)
	if err != nil {
		return TRC20TransferPreview{}, err
	}
	if amountRaw.Sign() <= 0 {
		return TRC20TransferPreview{}, errors.New("amount must be greater than 0")
	}

	addressWord := abiAddressParameter(to.AddressHex)
	amountWord := abiUint256Parameter(amountRaw)
	parameter := addressWord + amountWord
	if err := validateTRC20TransferParameter(parameter, to.AddressHex, amountRaw); err != nil {
		return TRC20TransferPreview{}, err
	}

	return TRC20TransferPreview{
		Token:            token,
		ContractAddress:  contract.AddressBase58,
		ContractHex:      contract.AddressHex,
		ToAddress:        to.AddressBase58,
		ToHex:            to.AddressHex,
		Amount:           formatBigTokenAmount(amountRaw, decimals),
		AmountRaw:        amountRaw.Text(10),
		Decimals:         decimals,
		FunctionSelector: TRC20TransferSelector,
		MethodID:         TRC20TransferMethodID,
		Parameter:        parameter,
		Data:             TRC20TransferMethodID + parameter,
		AddressWord:      addressWord,
		AmountWord:       amountWord,
	}, nil
}

func SimulateTRC20TransferOnNetwork(ctx context.Context, client *http.Client, network string, endpoints []string, ownerAddress string, preview TRC20TransferPreview) (TRC20TransferDryRun, error) {
	resolved, err := ResolveFullNodeEndpoints(network, endpoints)
	if err != nil {
		return TRC20TransferDryRun{}, err
	}
	return SimulateTRC20Transfer(ctx, client, resolved, ownerAddress, preview)
}

func SimulateTRC20Transfer(ctx context.Context, client *http.Client, endpoints []string, ownerAddress string, preview TRC20TransferPreview) (TRC20TransferDryRun, error) {
	owner, err := ValidateAddress(ownerAddress)
	if err != nil {
		return TRC20TransferDryRun{}, fmt.Errorf("owner address: %w", err)
	}
	if client == nil {
		client = &http.Client{}
	}
	endpoints = normalizedEndpoints(endpoints)

	body, err := json.Marshal(map[string]any{
		"owner_address":     owner.AddressBase58,
		"contract_address":  preview.ContractAddress,
		"function_selector": preview.FunctionSelector,
		"parameter":         preview.Parameter,
		"visible":           true,
	})
	if err != nil {
		return TRC20TransferDryRun{}, err
	}

	var payload triggerConstantContractResponse
	if err := postWithFallback(ctx, client, endpoints, "/wallet/triggerconstantcontract", body, &payload); err != nil {
		return TRC20TransferDryRun{}, err
	}

	dryRun := TRC20TransferDryRun{
		OwnerAddress:   owner.AddressBase58,
		OwnerHex:       owner.AddressHex,
		OK:             payload.Result.Result,
		ConstantResult: payload.ConstantResult,
		EnergyUsed:     payload.EnergyUsed,
		LogCount:       len(payload.Logs),
		Message:        payload.Result.Message,
	}
	if !dryRun.OK {
		if dryRun.Message == "" {
			dryRun.Message = payload.Result.Code
		}
		if dryRun.Message == "" {
			dryRun.Message = "TRON node rejected constant contract execution"
		}
		return dryRun, fmt.Errorf("dry-run failed: %s", dryRun.Message)
	}
	if strings.TrimSpace(dryRun.Message) != "" {
		dryRun.OK = false
		return dryRun, fmt.Errorf("dry-run failed: %s", dryRun.Message)
	}
	for _, ret := range payload.Transaction.Ret {
		if ret.Ret != "" && ret.Ret != "SUCCESS" {
			dryRun.OK = false
			dryRun.Message = ret.Ret
			return dryRun, fmt.Errorf("dry-run failed: %s", dryRun.Message)
		}
	}
	return dryRun, nil
}

func parseDecimalTokenAmount(amount string, decimals int) (*big.Int, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return nil, errors.New("amount is required")
	}
	if strings.HasPrefix(amount, "+") {
		amount = strings.TrimPrefix(amount, "+")
	}
	if strings.HasPrefix(amount, "-") {
		return nil, errors.New("amount must be greater than 0")
	}
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid amount %q", amount)
	}
	whole := parts[0]
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if whole == "" {
		whole = "0"
	}
	if whole == "" || !decimalDigits(whole) || !decimalDigits(fraction) {
		return nil, fmt.Errorf("invalid amount %q", amount)
	}
	if len(fraction) > decimals {
		return nil, fmt.Errorf("amount has more than %d decimal place(s)", decimals)
	}
	fraction += strings.Repeat("0", decimals-len(fraction))
	raw := strings.TrimLeft(whole+fraction, "0")
	if raw == "" {
		raw = "0"
	}
	value, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount %q", amount)
	}
	return value, nil
}

func decimalDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func abiUint256Parameter(value *big.Int) string {
	if value == nil {
		value = big.NewInt(0)
	}
	hexValue := value.Text(16)
	return strings.Repeat("0", 64-len(hexValue)) + hexValue
}

func validateTRC20TransferParameter(parameter string, toHex string, amountRaw *big.Int) error {
	if len(parameter) != 128 {
		return fmt.Errorf("TRC20 transfer parameter must be 128 hex characters, got %d", len(parameter))
	}
	if _, ok := new(big.Int).SetString(parameter, 16); !ok {
		return errors.New("TRC20 transfer parameter is not valid hex")
	}
	addressWord := parameter[:64]
	amountWord := parameter[64:]
	wantAddressWord := abiAddressParameter(toHex)
	if addressWord != wantAddressWord {
		return errors.New("TRC20 transfer address word does not round-trip to the destination address")
	}
	gotAmount, ok := new(big.Int).SetString(amountWord, 16)
	if !ok || gotAmount.Cmp(amountRaw) != 0 {
		return errors.New("TRC20 transfer amount word does not round-trip to the requested amount")
	}
	return nil
}
