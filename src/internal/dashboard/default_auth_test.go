package dashboard

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestDefaultEmbeddedAuth_optiPassword(t *testing.T) {
	t.Parallel()
	f, err := DefaultEmbeddedAuth()
	if err != nil {
		t.Fatal(err)
	}
	u := f.Lookup("opti")
	if u == nil {
		t.Fatal("missing opti user")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("opti")); err != nil {
		t.Fatalf("password opti: %v", err)
	}
}
