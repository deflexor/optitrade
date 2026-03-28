package state

// Order lifecycle strings stored in order_record.state (aligned with Deribit order_state where useful).
const (
	OrderStateNew             = "new"
	OrderStateOpen            = "open"
	OrderStateFilled          = "filled"
	OrderStateCanceled        = "canceled"
	OrderStateRejected        = "rejected"
	OrderStatePartiallyFilled = "partially_filled"
)

// NonTerminalOrderStates returns states we still reconcile against the exchange book.
func NonTerminalOrderStates() []string {
	return []string{
		OrderStateNew,
		OrderStateOpen,
		OrderStatePartiallyFilled,
	}
}
