package keychain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ifuryst/coldkit/internal/tron"
)

const tronPrivateKeyService = "coldkit.tron.private-key"

var keyNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

type StoredTronKey struct {
	Name          string `json:"name"`
	AddressBase58 string `json:"address_base58"`
	AddressHex    string `json:"address_hex"`
	PublicKeyHex  string `json:"public_key_hex"`
}

func ImportTronPrivateKey(name string, privateKeyHex string) (StoredTronKey, error) {
	name, err := normalizeKeyName(name)
	if err != nil {
		return StoredTronKey{}, err
	}
	account, err := tron.AccountFromPrivateKey(privateKeyHex)
	if err != nil {
		return StoredTronKey{}, err
	}
	if err := storeSecret(name, account.PrivateKeyHex, keyComment(account)); err != nil {
		return StoredTronKey{}, err
	}
	return storedKey(name, account), nil
}

func LoadTronPrivateKey(name string) (string, error) {
	name, err := normalizeKeyName(name)
	if err != nil {
		return "", err
	}
	secret, err := loadSecret(name)
	if err != nil {
		return "", err
	}
	account, err := tron.AccountFromPrivateKey(secret)
	if err != nil {
		return "", fmt.Errorf("stored key %q is not a valid TRON private key: %w", name, err)
	}
	return account.PrivateKeyHex, nil
}

func DescribeTronKey(name string) (StoredTronKey, error) {
	privateKey, err := LoadTronPrivateKey(name)
	if err != nil {
		return StoredTronKey{}, err
	}
	account, err := tron.AccountFromPrivateKey(privateKey)
	if err != nil {
		return StoredTronKey{}, err
	}
	normalized, _ := normalizeKeyName(name)
	return storedKey(normalized, account), nil
}

func DeleteTronKey(name string) error {
	name, err := normalizeKeyName(name)
	if err != nil {
		return err
	}
	return deleteSecret(name)
}

func normalizeKeyName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("key name is required")
	}
	if !keyNamePattern.MatchString(name) {
		return "", errors.New("key name must start with a letter or number and contain only letters, numbers, underscores, dots, and hyphens")
	}
	return name, nil
}

func storedKey(name string, account tron.Account) StoredTronKey {
	return StoredTronKey{
		Name:          name,
		AddressBase58: account.AddressBase58,
		AddressHex:    account.AddressHex,
		PublicKeyHex:  account.PublicKeyHex,
	}
}

func keyComment(account tron.Account) string {
	return fmt.Sprintf("coldkit TRON key for %s", account.AddressBase58)
}
