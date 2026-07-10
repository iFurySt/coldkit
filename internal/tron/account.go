package tron

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

const Base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	curveP, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	curveN, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	curveGx, _ = new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	curveGy, _ = new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
	bigTwo     = big.NewInt(2)
	bigThree   = big.NewInt(3)
)

type point struct {
	x   *big.Int
	y   *big.Int
	inf bool
}

type Account struct {
	AddressBase58 string `json:"address_base58"`
	AddressHex    string `json:"address_hex"`
	PrivateKeyHex string `json:"private_key_hex,omitempty"`
	PublicKeyHex  string `json:"public_key_hex"`
}

type PublicAccount struct {
	AddressBase58 string `json:"address_base58"`
	AddressHex    string `json:"address_hex"`
	PublicKeyHex  string `json:"public_key_hex"`
}

type Address struct {
	AddressBase58 string `json:"address_base58"`
	AddressHex    string `json:"address_hex"`
}

type GeneratedAccount struct {
	Account
	MatchedPrefix string `json:"matched_prefix,omitempty"`
	MatchedSuffix string `json:"matched_suffix,omitempty"`
	Attempts      uint64 `json:"attempts"`
}

type GeneratedPublicAccount struct {
	PublicAccount
	MatchedPrefix string `json:"matched_prefix,omitempty"`
	MatchedSuffix string `json:"matched_suffix,omitempty"`
	Attempts      uint64 `json:"attempts"`
}

type GenerateOptions struct {
	Prefixes    []string
	Suffixes    []string
	Count       uint64
	MaxAttempts uint64
}

func GenerateAccount() (Account, error) {
	for {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return Account{}, err
		}
		scalar := new(big.Int).SetBytes(key)
		if scalar.Sign() > 0 && scalar.Cmp(curveN) < 0 {
			return AccountFromPrivateKey(hex.EncodeToString(key))
		}
	}
}

func GenerateAccounts(options GenerateOptions) ([]GeneratedAccount, error) {
	if options.Count == 0 {
		return nil, errors.New("count must be greater than 0")
	}
	prefixes := normalizePatterns(options.Prefixes)
	suffixes := normalizePatterns(options.Suffixes)
	for _, prefix := range prefixes {
		if err := validateBase58Pattern(prefix); err != nil {
			return nil, fmt.Errorf("prefix %q: %w", prefix, err)
		}
	}
	for _, suffix := range suffixes {
		if err := validateBase58Pattern(suffix); err != nil {
			return nil, fmt.Errorf("suffix %q: %w", suffix, err)
		}
	}

	results := make([]GeneratedAccount, 0, options.Count)
	var totalAttempts uint64
	var attemptsSinceMatch uint64
	for uint64(len(results)) < options.Count {
		if options.MaxAttempts > 0 && totalAttempts >= options.MaxAttempts {
			return nil, fmt.Errorf("found %d address(es), stopped after %d attempts before reaching count %d", len(results), totalAttempts, options.Count)
		}

		account, err := GenerateAccount()
		if err != nil {
			return nil, err
		}
		totalAttempts++
		attemptsSinceMatch++

		matchedPrefix, prefixOK := matchPrefix(account.AddressBase58, prefixes)
		matchedSuffix, suffixOK := matchSuffix(account.AddressBase58, suffixes)
		if prefixOK && suffixOK {
			results = append(results, GeneratedAccount{
				Account:       account,
				MatchedPrefix: matchedPrefix,
				MatchedSuffix: matchedSuffix,
				Attempts:      attemptsSinceMatch,
			})
			attemptsSinceMatch = 0
		}
	}
	return results, nil
}

func PublicResults(results []GeneratedAccount) []GeneratedPublicAccount {
	public := make([]GeneratedPublicAccount, 0, len(results))
	for _, result := range results {
		public = append(public, GeneratedPublicAccount{
			PublicAccount: PublicAccount{
				AddressBase58: result.AddressBase58,
				AddressHex:    result.AddressHex,
				PublicKeyHex:  result.PublicKeyHex,
			},
			MatchedPrefix: result.MatchedPrefix,
			MatchedSuffix: result.MatchedSuffix,
			Attempts:      result.Attempts,
		})
	}
	return public
}

func AccountFromPrivateKey(privateKeyHex string) (Account, error) {
	key := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(privateKeyHex)), "0x")
	if len(key) != 64 {
		return Account{}, errors.New("private key must be 64 hex characters")
	}
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return Account{}, fmt.Errorf("decode private key: %w", err)
	}
	scalar := new(big.Int).SetBytes(keyBytes)
	if scalar.Sign() <= 0 || scalar.Cmp(curveN) >= 0 {
		return Account{}, errors.New("private key is outside the secp256k1 range")
	}

	pub, err := scalarBaseMult(scalar)
	if err != nil {
		return Account{}, err
	}
	publicKey := make([]byte, 65)
	publicKey[0] = 0x04
	copy(publicKey[1+32-len(pub.x.Bytes()):33], pub.x.Bytes())
	copy(publicKey[33+32-len(pub.y.Bytes()):], pub.y.Bytes())

	payload := publicKeyToTronPayload(publicKey)
	return Account{
		AddressBase58: base58CheckEncode(payload),
		AddressHex:    hex.EncodeToString(payload),
		PrivateKeyHex: key,
		PublicKeyHex:  hex.EncodeToString(publicKey),
	}, nil
}

func ValidateAddress(address string) (Address, error) {
	payload, err := base58CheckDecode(strings.TrimSpace(address))
	if err != nil {
		return Address{}, err
	}
	if len(payload) != 21 || payload[0] != 0x41 {
		return Address{}, errors.New("not a mainnet TRON address payload")
	}
	return Address{
		AddressBase58: strings.TrimSpace(address),
		AddressHex:    hex.EncodeToString(payload),
	}, nil
}

func SelfTest() error {
	hash := legacyKeccak256(nil)
	if hex.EncodeToString(hash) != "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" {
		return errors.New("Keccak-256 vector failed")
	}
	account, err := AccountFromPrivateKey(strings.Repeat("88", 32))
	if err != nil {
		return err
	}
	if account.AddressBase58 != "TJzXt1sZautjqXnpjQT4xSCBHNSYgBkDr3" {
		return fmt.Errorf("TRON address vector failed: got %s", account.AddressBase58)
	}
	return nil
}

func validateBase58Pattern(pattern string) error {
	for _, char := range pattern {
		if strings.IndexRune(Base58Alphabet, char) < 0 {
			return fmt.Errorf("contains invalid Base58 character %q", char)
		}
	}
	return nil
}

func normalizePatterns(patterns []string) []string {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern != "" {
			normalized = append(normalized, pattern)
		}
	}
	return normalized
}

func matchPrefix(address string, prefixes []string) (string, bool) {
	if len(prefixes) == 0 {
		return "", true
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(address, prefix) {
			return prefix, true
		}
	}
	return "", false
}

func matchSuffix(address string, suffixes []string) (string, bool) {
	if len(suffixes) == 0 {
		return "", true
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(address, suffix) {
			return suffix, true
		}
	}
	return "", false
}

func publicKeyToTronPayload(publicKey []byte) []byte {
	hash := legacyKeccak256(publicKey[1:])
	payload := make([]byte, 21)
	payload[0] = 0x41
	copy(payload[1:], hash[len(hash)-20:])
	return payload
}

func legacyKeccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}

func scalarBaseMult(scalar *big.Int) (point, error) {
	result := point{inf: true}
	addend := point{x: new(big.Int).Set(curveGx), y: new(big.Int).Set(curveGy)}

	for i := scalar.BitLen() - 1; i >= 0; i-- {
		result = pointDouble(result)
		if scalar.Bit(i) == 1 {
			result = pointAdd(result, addend)
		}
	}
	if result.inf {
		return point{}, errors.New("invalid private key")
	}
	return result, nil
}

func pointAdd(left, right point) point {
	if left.inf {
		return clonePoint(right)
	}
	if right.inf {
		return clonePoint(left)
	}
	if left.x.Cmp(right.x) == 0 {
		sumY := modAdd(left.y, right.y)
		if sumY.Sign() == 0 {
			return point{inf: true}
		}
		return pointDouble(left)
	}

	numerator := modSub(right.y, left.y)
	denominator := modSub(right.x, left.x)
	slope := modMul(numerator, modInverse(denominator))
	x3 := modSub(modSub(modMul(slope, slope), left.x), right.x)
	y3 := modSub(modMul(slope, modSub(left.x, x3)), left.y)
	return point{x: x3, y: y3}
}

func pointDouble(value point) point {
	if value.inf || value.y.Sign() == 0 {
		return point{inf: true}
	}
	numerator := modMul(bigThree, modMul(value.x, value.x))
	denominator := modMul(bigTwo, value.y)
	slope := modMul(numerator, modInverse(denominator))
	x3 := modSub(modMul(slope, slope), modMul(bigTwo, value.x))
	y3 := modSub(modMul(slope, modSub(value.x, x3)), value.y)
	return point{x: x3, y: y3}
}

func clonePoint(value point) point {
	if value.inf {
		return point{inf: true}
	}
	return point{x: new(big.Int).Set(value.x), y: new(big.Int).Set(value.y)}
}

func modAdd(left, right *big.Int) *big.Int {
	return new(big.Int).Mod(new(big.Int).Add(left, right), curveP)
}

func modSub(left, right *big.Int) *big.Int {
	result := new(big.Int).Sub(left, right)
	result.Mod(result, curveP)
	if result.Sign() < 0 {
		result.Add(result, curveP)
	}
	return result
}

func modMul(left, right *big.Int) *big.Int {
	return new(big.Int).Mod(new(big.Int).Mul(left, right), curveP)
}

func modInverse(value *big.Int) *big.Int {
	return new(big.Int).ModInverse(value, curveP)
}

func base58CheckEncode(payload []byte) string {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	raw := append(append([]byte{}, payload...), second[:4]...)
	value := new(big.Int).SetBytes(raw)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	encoded := ""
	for value.Cmp(zero) > 0 {
		value.DivMod(value, base, mod)
		encoded = string(Base58Alphabet[mod.Int64()]) + encoded
	}
	for _, b := range raw {
		if b != 0 {
			break
		}
		encoded = "1" + encoded
	}
	return encoded
}

func base58CheckDecode(address string) ([]byte, error) {
	value := big.NewInt(0)
	base := big.NewInt(58)
	for _, char := range address {
		index := strings.IndexRune(Base58Alphabet, char)
		if index < 0 {
			return nil, fmt.Errorf("invalid Base58 character: %q", char)
		}
		value.Mul(value, base)
		value.Add(value, big.NewInt(int64(index)))
	}

	raw := value.Bytes()
	for _, char := range address {
		if char != '1' {
			break
		}
		raw = append([]byte{0}, raw...)
	}
	if len(raw) < 5 {
		return nil, errors.New("address is too short")
	}
	payload := raw[:len(raw)-4]
	checksum := raw[len(raw)-4:]
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	if !equalBytes(checksum, second[:4]) {
		return nil, errors.New("Base58Check checksum mismatch")
	}
	return payload, nil
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var diff byte
	for i := range left {
		diff |= left[i] ^ right[i]
	}
	return diff == 0
}
