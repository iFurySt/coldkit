package tron

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildTRC20TransferPreview(t *testing.T) {
	preview, err := BuildTRC20TransferPreview(TRC20TransferOptions{
		ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J",
		Amount:    "30",
		Decimals:  6,
	})
	if err != nil {
		t.Fatal(err)
	}

	wantAddressWord := "0000000000000000000000005fbdbc82fc5bf70e92ffcd07b2f94fc95770b127"
	wantAmountWord := "0000000000000000000000000000000000000000000000000000000001c9c380"
	if preview.ContractAddress != USDTTRC20Contract {
		t.Fatalf("contract = %s", preview.ContractAddress)
	}
	if preview.AmountRaw != "30000000" || preview.Amount != "30" {
		t.Fatalf("amount = %s raw = %s", preview.Amount, preview.AmountRaw)
	}
	if preview.AddressWord != wantAddressWord {
		t.Fatalf("address word = %s", preview.AddressWord)
	}
	if preview.AmountWord != wantAmountWord {
		t.Fatalf("amount word = %s", preview.AmountWord)
	}
	if preview.Parameter != wantAddressWord+wantAmountWord {
		t.Fatalf("parameter = %s", preview.Parameter)
	}
	if preview.Data != TRC20TransferMethodID+preview.Parameter {
		t.Fatalf("data = %s", preview.Data)
	}
	if len(preview.Parameter) != 128 {
		t.Fatalf("parameter length = %d", len(preview.Parameter))
	}
}

func TestBuildTRC20TransferPreviewRejectsUnsafeInputs(t *testing.T) {
	tests := []TRC20TransferOptions{
		{ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J", Amount: "0", Decimals: 6},
		{ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J", Amount: "0.0000001", Decimals: 6},
		{ToAddress: "not-an-address", Amount: "30", Decimals: 6},
		{ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J", Amount: "1e6", Decimals: 6},
	}
	for _, tt := range tests {
		if _, err := BuildTRC20TransferPreview(tt); err == nil {
			t.Fatalf("expected error for %+v", tt)
		}
	}
}

func TestBuildTRC20TransferPreviewAllowsZeroDecimals(t *testing.T) {
	preview, err := BuildTRC20TransferPreview(TRC20TransferOptions{
		ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J",
		Amount:    "30",
		Decimals:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.AmountRaw != "30" {
		t.Fatalf("amount raw = %s", preview.AmountRaw)
	}
	if preview.AmountWord != "000000000000000000000000000000000000000000000000000000000000001e" {
		t.Fatalf("amount word = %s", preview.AmountWord)
	}
}

func TestSimulateTRC20TransferPostsConstantContractDryRun(t *testing.T) {
	preview, err := BuildTRC20TransferPreview(TRC20TransferOptions{
		ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J",
		Amount:    "30",
		Decimals:  6,
	})
	if err != nil {
		t.Fatal(err)
	}

	var request map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wallet/triggerconstantcontract" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"result":true},"constant_result":["0000000000000000000000000000000000000000000000000000000000000000"],"energy_used":65005,"logs":[{"data":"01"}],"transaction":{"ret":[{}]}}`))
	}))
	defer server.Close()

	dryRun, err := SimulateTRC20Transfer(context.Background(), server.Client(), []string{server.URL}, "TZCus6mBcvEi3Jy7PQMGbSLyHhBGnL6888", preview)
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK || dryRun.EnergyUsed != 65005 || dryRun.LogCount != 1 {
		t.Fatalf("dryRun = %+v", dryRun)
	}
	if request["owner_address"] != "TZCus6mBcvEi3Jy7PQMGbSLyHhBGnL6888" {
		t.Fatalf("owner address = %v", request["owner_address"])
	}
	if request["contract_address"] != USDTTRC20Contract || request["function_selector"] != TRC20TransferSelector {
		t.Fatalf("request = %#v", request)
	}
	if request["parameter"] != preview.Parameter {
		t.Fatalf("parameter = %v", request["parameter"])
	}
	if strings.HasPrefix(preview.Parameter, "00000000000000000000000041") {
		t.Fatalf("parameter kept TRON 0x41 prefix in the address word: %s", preview.Parameter)
	}
}

func TestSimulateTRC20TransferRejectsFailedDryRun(t *testing.T) {
	preview, err := BuildTRC20TransferPreview(TRC20TransferOptions{
		ToAddress: "TJhSVPzbatkY5rLGWapBVSXpL5Ws7ZVx6J",
		Amount:    "30",
		Decimals:  6,
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"result":true,"message":"REVERT opcode executed"},"transaction":{"ret":[{"ret":"FAILED"}]}}`))
	}))
	defer server.Close()

	dryRun, err := SimulateTRC20Transfer(context.Background(), server.Client(), []string{server.URL}, "TZCus6mBcvEi3Jy7PQMGbSLyHhBGnL6888", preview)
	if err == nil {
		t.Fatal("expected dry-run error")
	}
	if dryRun.OK || !strings.Contains(err.Error(), "REVERT") {
		t.Fatalf("dryRun = %+v err = %v", dryRun, err)
	}
}
