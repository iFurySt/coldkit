package keychain

import "testing"

func TestNormalizeKeyName(t *testing.T) {
	tests := []string{"main", "wallet_1", "wallet.1", "wallet-1"}
	for _, input := range tests {
		if _, err := normalizeKeyName(input); err != nil {
			t.Fatalf("%q: %v", input, err)
		}
	}
}

func TestNormalizeKeyNameRejectsUnsafeNames(t *testing.T) {
	tests := []string{"", "../key", "-key", "bad/key", "bad key"}
	for _, input := range tests {
		if _, err := normalizeKeyName(input); err == nil {
			t.Fatalf("%q: expected error", input)
		}
	}
}
