package tron

import (
	"strings"
	"testing"
)

func TestSignDigestVerifies(t *testing.T) {
	privateKey := strings.Repeat("88", 32)
	digest := strings.Repeat("11", 32)

	signature, err := SignDigest(privateKey, digest)
	if err != nil {
		t.Fatal(err)
	}
	if len(signature.SignatureHex) != 130 {
		t.Fatalf("got signature hex length %d", len(signature.SignatureHex))
	}
	if signature.RecoveryID < 0 || signature.RecoveryID > 3 {
		t.Fatalf("invalid recovery id %d", signature.RecoveryID)
	}

	account, err := AccountFromPrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyDigestSignature(account.PublicKeyHex, digest, signature)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature did not verify")
	}
}

func TestSignDigestRejectsInvalidDigest(t *testing.T) {
	_, err := SignDigest(strings.Repeat("88", 32), "abcd")
	if err == nil {
		t.Fatal("expected error")
	}
}
