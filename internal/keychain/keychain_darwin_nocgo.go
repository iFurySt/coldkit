//go:build darwin && !cgo

package keychain

import "errors"

func storeSecret(account string, secret string, comment string) error {
	return unsupportedNative()
}

func loadSecret(account string) (string, error) {
	return "", unsupportedNative()
}

func deleteSecret(account string) error {
	return unsupportedNative()
}

func unsupportedNative() error {
	return errors.New("native macOS Keychain backend requires a cgo-enabled macOS build")
}
