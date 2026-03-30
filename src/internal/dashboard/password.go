package dashboard

import (
	"golang.org/x/crypto/bcrypt"
)

func verifyPasswordRecord(hash, plain string) error {
	// Supported: bcrypt-encoded strings ($2a$, $2b$, …).
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
