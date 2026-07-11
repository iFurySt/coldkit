//go:build !darwin

package keychain

import "errors"

func storeSecret(account string, secret string, comment string) error {
	return unsupported()
}

func loadSecret(account string) (string, error) {
	return "", unsupported()
}

func deleteSecret(account string) error {
	return unsupported()
}

func unsupported() error {
	return errors.New("macOS Keychain backend is only available on macOS")
}
