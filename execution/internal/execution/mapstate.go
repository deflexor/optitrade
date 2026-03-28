package execution

import (
	"strings"

	"github.com/dfr/optitrade/execution/internal/state"
)

// MapExchangeOrderState maps Deribit order_state to order_record.state.
func MapExchangeOrderState(ex string) string {
	switch strings.ToLower(strings.TrimSpace(ex)) {
	case "filled":
		return state.OrderStateFilled
	case "cancelled", "canceled":
		return state.OrderStateCanceled
	case "rejected":
		return state.OrderStateRejected
	case "open":
		return state.OrderStateOpen
	case "untriggered", "triggered":
		return state.OrderStateOpen
	default:
		return state.OrderStateOpen
	}
}
