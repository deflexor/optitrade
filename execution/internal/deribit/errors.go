package deribit

import "errors"

var (
	errAuthEmpty     = errors.New("deribit: auth response missing access_token")
	errShortTuple    = errors.New("deribit: expected [price, size] tuple")
	ErrNoCredentials = errors.New("deribit: private call requires credentials")
)
