package tron

import (
	"crypto/rand"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type Signature struct {
	AddressBase58   string `json:"address_base58"`
	AddressHex      string `json:"address_hex"`
	PublicKeyHex    string `json:"public_key_hex"`
	DigestHex       string `json:"digest_hex"`
	RHex            string `json:"r_hex"`
	SHex            string `json:"s_hex"`
	RecoveryID      int    `json:"recovery_id"`
	SignatureHex    string `json:"signature_hex"`
	DERSignatureHex string `json:"der_signature_hex"`
}

type ecdsaSignature struct {
	R *big.Int
	S *big.Int
}

func SignDigest(privateKeyHex string, digestHex string) (Signature, error) {
	account, err := AccountFromPrivateKey(privateKeyHex)
	if err != nil {
		return Signature{}, err
	}
	digest, err := decodeDigest(digestHex)
	if err != nil {
		return Signature{}, err
	}
	privateScalar := privateScalarFromHex(account.PrivateKeyHex)

	var r, s *big.Int
	var recoveryID int
	for {
		nonce, err := randomScalar()
		if err != nil {
			return Signature{}, err
		}
		rPoint, err := scalarBaseMult(nonce)
		if err != nil {
			return Signature{}, err
		}
		r = new(big.Int).Mod(rPoint.x, curveN)
		if r.Sign() == 0 {
			continue
		}

		digestScalar := new(big.Int).SetBytes(digest)
		s = new(big.Int).Mul(r, privateScalar)
		s.Add(s, digestScalar)
		s.Mul(s, new(big.Int).ModInverse(nonce, curveN))
		s.Mod(s, curveN)
		if s.Sign() == 0 {
			continue
		}

		recoveryID = 0
		if rPoint.y.Bit(0) == 1 {
			recoveryID |= 1
		}
		if rPoint.x.Cmp(curveN) >= 0 {
			recoveryID |= 2
		}
		halfOrder := new(big.Int).Rsh(new(big.Int).Set(curveN), 1)
		if s.Cmp(halfOrder) > 0 {
			s.Sub(curveN, s)
			recoveryID ^= 1
		}
		break
	}

	der, err := asn1.Marshal(ecdsaSignature{R: r, S: s})
	if err != nil {
		return Signature{}, fmt.Errorf("encode DER signature: %w", err)
	}
	raw := append(pad32(r), pad32(s)...)
	raw = append(raw, byte(recoveryID))

	return Signature{
		AddressBase58:   account.AddressBase58,
		AddressHex:      account.AddressHex,
		PublicKeyHex:    account.PublicKeyHex,
		DigestHex:       hex.EncodeToString(digest),
		RHex:            hex.EncodeToString(pad32(r)),
		SHex:            hex.EncodeToString(pad32(s)),
		RecoveryID:      recoveryID,
		SignatureHex:    hex.EncodeToString(raw),
		DERSignatureHex: hex.EncodeToString(der),
	}, nil
}

func VerifyDigestSignature(publicKeyHex string, digestHex string, signature Signature) (bool, error) {
	publicKey, err := hex.DecodeString(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(publicKeyHex)), "0x"))
	if err != nil {
		return false, fmt.Errorf("decode public key: %w", err)
	}
	if len(publicKey) != 65 || publicKey[0] != 0x04 {
		return false, errors.New("public key must be uncompressed 65-byte hex")
	}
	digest, err := decodeDigest(digestHex)
	if err != nil {
		return false, err
	}
	r, ok := new(big.Int).SetString(signature.RHex, 16)
	if !ok {
		return false, errors.New("invalid r hex")
	}
	s, ok := new(big.Int).SetString(signature.SHex, 16)
	if !ok {
		return false, errors.New("invalid s hex")
	}
	if r.Sign() <= 0 || r.Cmp(curveN) >= 0 || s.Sign() <= 0 || s.Cmp(curveN) >= 0 {
		return false, nil
	}

	w := new(big.Int).ModInverse(s, curveN)
	if w == nil {
		return false, nil
	}
	e := new(big.Int).SetBytes(digest)
	u1 := new(big.Int).Mod(new(big.Int).Mul(e, w), curveN)
	u2 := new(big.Int).Mod(new(big.Int).Mul(r, w), curveN)
	gMul, err := scalarBaseMult(u1)
	if err != nil {
		return false, err
	}
	pub := point{
		x: new(big.Int).SetBytes(publicKey[1:33]),
		y: new(big.Int).SetBytes(publicKey[33:65]),
	}
	qMul := scalarMult(pub, u2)
	sum := pointAdd(gMul, qMul)
	if sum.inf {
		return false, nil
	}
	x := new(big.Int).Mod(sum.x, curveN)
	return x.Cmp(r) == 0, nil
}

func decodeDigest(digestHex string) ([]byte, error) {
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(digestHex)), "0x")
	if len(normalized) != 64 {
		return nil, errors.New("digest must be 32 bytes encoded as 64 hex characters")
	}
	digest, err := hex.DecodeString(normalized)
	if err != nil {
		return nil, fmt.Errorf("decode digest: %w", err)
	}
	return digest, nil
}

func randomScalar() (*big.Int, error) {
	for {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		scalar := new(big.Int).SetBytes(bytes)
		if scalar.Sign() > 0 && scalar.Cmp(curveN) < 0 {
			return scalar, nil
		}
	}
}

func privateScalarFromHex(privateKeyHex string) *big.Int {
	privateBytes, _ := hex.DecodeString(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(privateKeyHex)), "0x"))
	return new(big.Int).SetBytes(privateBytes)
}

func pad32(value *big.Int) []byte {
	out := make([]byte, 32)
	bytes := value.Bytes()
	copy(out[32-len(bytes):], bytes)
	return out
}

func scalarMult(base point, scalar *big.Int) point {
	result := point{inf: true}
	addend := clonePoint(base)
	for i := scalar.BitLen() - 1; i >= 0; i-- {
		result = pointDouble(result)
		if scalar.Bit(i) == 1 {
			result = pointAdd(result, addend)
		}
	}
	return result
}
