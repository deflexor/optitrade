package dashboard

import _ "embed"

//go:embed default_operator_auth.json
var defaultOperatorAuthJSON []byte

// DefaultEmbeddedAuth is the built-in allowlist used when no auth file path is set.
// Username: opti, password: opti (development default only — override with OPTITRADE_DASHBOARD_AUTH_PATH in production).
func DefaultEmbeddedAuth() (*DashboardAuthFile, error) {
	return ParseDashboardAuthJSON(defaultOperatorAuthJSON)
}
