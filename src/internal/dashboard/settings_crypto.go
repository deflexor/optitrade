package dashboard

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// SettingsCrypto holds AES-GCM key material for operator settings at rest.
type SettingsCrypto struct {
	gcm cipher.AEAD
}

const settingsBlobVersion byte = 1

// LoadSettingsCrypto builds AES-GCM from OPTITRADE_SETTINGS_KEY_FILE (raw key bytes)
// or OPTITRADE_SETTINGS_SECRET (base64, hex, or exactly 32 raw bytes).
func LoadSettingsCrypto() (*SettingsCrypto, error) {
	var key []byte
	if p := strings.TrimSpace(os.Getenv("OPTITRADE_SETTINGS_KEY_FILE")); p != "" {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read OPTITRADE_SETTINGS_KEY_FILE: %w", err)
		}
		key = normalizeKeyFileBytes(raw)
	} else if s := strings.TrimSpace(os.Getenv("OPTITRADE_SETTINGS_SECRET")); s != "" {
		var err error
		key, err = decodeKeyString(s)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("set OPTITRADE_SETTINGS_SECRET or OPTITRADE_SETTINGS_KEY_FILE for dashboard settings encryption")
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("settings encryption key must decode to 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SettingsCrypto{gcm: gcm}, nil
}

func normalizeKeyFileBytes(raw []byte) []byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 32 {
		return raw
	}
	s := string(raw)
	if len(s) == 64 && hexLooks(s) {
		if b, err := hex.DecodeString(s); err == nil && len(b) == 32 {
			return b
		}
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b
	}
	return raw
}

func hexLooks(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func decodeKeyString(s string) ([]byte, error) {
	// Prefer exact 32-byte UTF-8 secret (common for dev); avoids accidental base64
	// interpretation of ASCII passphrases that happen to be valid base64.
	if len(s) == 32 {
		return []byte(s), nil
	}
	// hex (64 chars = 32 bytes)
	if len(s) == 64 && hexLooks(s) {
		return hex.DecodeString(s)
	}
	// base64 standard → 32 bytes
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	return nil, fmt.Errorf("OPTITRADE_SETTINGS_SECRET: use raw 32-byte string, 64-char hex, or base64 of 32 bytes")
}

// Encrypt appends version, nonce, ciphertext for secrets JSON.
func (c *SettingsCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	if c == nil {
		return nil, errors.New("nil SettingsCrypto")
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := c.gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 1+len(nonce)+len(ct))
	out[0] = settingsBlobVersion
	copy(out[1:], nonce)
	copy(out[1+len(nonce):], ct)
	return out, nil
}

// Decrypt reverses Encrypt.
func (c *SettingsCrypto) Decrypt(blob []byte) ([]byte, error) {
	if c == nil {
		return nil, errors.New("nil SettingsCrypto")
	}
	if len(blob) < 2 {
		return nil, errors.New("empty secrets blob")
	}
	if blob[0] != settingsBlobVersion {
		return nil, fmt.Errorf("unsupported secrets blob version %d", blob[0])
	}
	ns := c.gcm.NonceSize()
	if len(blob) < 1+ns {
		return nil, errors.New("truncated secrets blob")
	}
	nonce := blob[1 : 1+ns]
	ct := blob[1+ns:]
	return c.gcm.Open(nil, nonce, ct, nil)
}
