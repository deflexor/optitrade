package dashboard

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"
)

func TestSettingsCryptoRoundTrip(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{'k'}, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	c := &SettingsCrypto{gcm: gcm}

	plain := []byte(`{"deribit_client_id":"abc","deribit_client_secret":"s"}`)
	blob, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.Decrypt(blob)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("decrypt mismatch: %q vs %q", out, plain)
	}
}
