package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DashboardAuthFile is the allowlisted-operator credential file (JSON).
type DashboardAuthFile struct {
	Version string           `json:"version"`
	Users   []AuthUserRecord `json:"users"`
}

// AuthUserRecord is one allowlisted operator (password is a verifier hash string).
type AuthUserRecord struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// LoadDashboardAuthFile reads and validates auth JSON from path.
func LoadDashboardAuthFile(path string) (*DashboardAuthFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dashboard auth file: %w", err)
	}
	return ParseDashboardAuthJSON(data)
}

// ParseDashboardAuthJSON unmarshals and validates allowlist JSON.
func ParseDashboardAuthJSON(data []byte) (*DashboardAuthFile, error) {
	var raw dashboardAuthFileRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("dashboard auth json: %w", err)
	}
	if strings.TrimSpace(raw.Version) == "" {
		return nil, fmt.Errorf("dashboard auth: missing version")
	}
	if len(raw.Users) == 0 {
		return nil, fmt.Errorf("dashboard auth: users must be non-empty")
	}
	seen := make(map[string]struct{}, len(raw.Users))
	out := &DashboardAuthFile{Version: strings.TrimSpace(raw.Version), Users: make([]AuthUserRecord, 0, len(raw.Users))}
	for _, u := range raw.Users {
		user := strings.TrimSpace(u.Username)
		ph := strings.TrimSpace(u.PasswordHash)
		if ph == "" {
			ph = strings.TrimSpace(u.PasswordHashAlt)
		}
		if user == "" || ph == "" {
			return nil, fmt.Errorf("dashboard auth: empty username or password_hash")
		}
		lc := strings.ToLower(user)
		if _, dup := seen[lc]; dup {
			return nil, fmt.Errorf("dashboard auth: duplicate username %q", user)
		}
		seen[lc] = struct{}{}
		out.Users = append(out.Users, AuthUserRecord{Username: user, PasswordHash: ph})
	}
	return out, nil
}

type dashboardAuthFileRaw struct {
	Version string `json:"version"`
	Users   []struct {
		Username      string `json:"username"`
		PasswordHash  string `json:"password_hash"`
		PasswordHashAlt string `json:"passwordHash"`
	} `json:"users"`
}

// Lookup returns the record for username (case-insensitive) or nil.
func (f *DashboardAuthFile) Lookup(username string) *AuthUserRecord {
	if f == nil {
		return nil
	}
	want := strings.ToLower(strings.TrimSpace(username))
	for i := range f.Users {
		if strings.ToLower(f.Users[i].Username) == want {
			return &f.Users[i]
		}
	}
	return nil
}
