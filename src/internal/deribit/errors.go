package deribit

import "errors"

var (
	errAuthEmpty        = errors.New("deribit: auth response missing access_token")
	errShortTuple       = errors.New("deribit: expected [price, size] tuple")
	ErrNoCredentials    = errors.New("deribit: private call requires credentials")
	errMissingInstrument = errors.New("deribit: instrument_name required")
	errMissingLabel      = errors.New("deribit: label required")
	errMissingOrderID    = errors.New("deribit: order_id required")
)
