package execution

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/state"
)

// FillIngestor persists user_trades and updates exposure (T049).
type FillIngestor struct {
	Fills    state.FillRepository
	Exposure *ExposureBook
}

// IngestUserTrade maps a Deribit user_trade into fill_record (idempotent on trade_id).
func (f *FillIngestor) IngestUserTrade(ctx context.Context, internalOrderID string, t deribit.UserTrade) error {
	if f == nil || f.Fills == nil {
		return fmt.Errorf("execution: nil FillIngestor")
	}
	if internalOrderID == "" || t.TradeID == "" || t.InstrumentName == "" {
		return fmt.Errorf("execution: missing fill identifiers")
	}
	amt := 0.0
	if t.Amount != nil {
		amt = *t.Amount
	}
	price := 0.0
	if t.Price != nil {
		price = *t.Price
	}
	fee := 0.0
	if t.Fee != nil {
		fee = *t.Fee
	}
	dir := ""
	if t.Direction != nil {
		dir = *t.Direction
	}
	rec := &state.FillRecord{
		ID:             uuid.NewString(),
		OrderID:        internalOrderID,
		TradeID:        t.TradeID,
		InstrumentName: t.InstrumentName,
		Qty:            formatFloat(amt),
		Price:          formatFloat(price),
		Fee:            formatFloat(fee),
		FilledAt:       t.Timestamp,
	}
	inserted, err := f.Fills.InsertFill(ctx, rec)
	if err != nil {
		return err
	}
	if inserted && f.Exposure != nil && amt != 0 {
		f.Exposure.ApplyUserTrade(t.InstrumentName, dir, amt)
	}
	return nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
