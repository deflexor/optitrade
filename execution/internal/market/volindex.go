package market

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit"
)

// VolIndexResolution is a supported public/get_volatility_index_data resolution string.
const VolIndexResolution1m = "60"

// FetchLatestVolIndex polls volatility index candles for currency and returns the latest complete candle close.
// window defines the lookback from endTimestamp (typically a few minutes).
// VolatilityIndexSource fetches vol index candles (implemented by *deribit.REST).
type VolatilityIndexSource interface {
	GetVolatilityIndexData(ctx context.Context, params deribit.GetVolatilityIndexDataParams) (*deribit.VolatilityIndexData, error)
}

func FetchLatestVolIndex(ctx context.Context, api VolatilityIndexSource, currency string, endMs int64, window time.Duration, resolution string) (close float64, candleStartMs int64, err error) {
	if api == nil {
		return 0, 0, fmt.Errorf("market: nil REST client")
	}
	if resolution == "" {
		resolution = VolIndexResolution1m
	}
	start := endMs - window.Milliseconds()
	if start < 0 {
		start = 0
	}
	res, err := api.GetVolatilityIndexData(ctx, deribit.GetVolatilityIndexDataParams{
		Currency:       currency,
		StartTimestamp: start,
		EndTimestamp:   endMs,
		Resolution:     resolution,
	})
	if err != nil {
		return 0, 0, err
	}
	if len(res.Data) == 0 {
		return 0, 0, fmt.Errorf("market: empty volatility index data")
	}
	last := res.Data[len(res.Data)-1]
	if len(last) < 5 {
		return 0, 0, fmt.Errorf("market: short volatility candle")
	}
	// JSON numbers decode as float64; timestamp is ms since epoch (fits float64 for our range).
	candleStartMs = int64(last[0])
	close = last[4]
	return close, candleStartMs, nil
}

// ParseVolIndexNotificationData decodes a Deribit volatility index subscription payload (ticker style) if needed by WS layer.
// The exact shape varies by channel; this helper accepts {"volatility" : 0.42, "timestamp": 123} or {"index_price": "0.42"}.
func ParseVolIndexNotificationData(data map[string]any) (value float64, tsMs int64, err error) {
	if v, ok := data["volatility"]; ok {
		value, err = floatFromAny(v)
		if err != nil {
			return 0, 0, err
		}
	}
	if value == 0 {
		if v, ok := data["index_price"]; ok {
			value, err = floatFromAny(v)
			if err != nil {
				return 0, 0, err
			}
		}
	}
	if v, ok := data["timestamp"]; ok {
		tsMs, err = int64FromAny(v)
		if err != nil {
			return 0, 0, err
		}
	}
	if v, ok := data["t"]; ok && tsMs == 0 {
		tsMs, _ = int64FromAny(v)
	}
	return value, tsMs, nil
}

func floatFromAny(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case string:
		return strconv.ParseFloat(x, 64)
	default:
		return 0, fmt.Errorf("market: cannot parse float from %T", v)
	}
}

func int64FromAny(v any) (int64, error) {
	switch x := v.(type) {
	case float64:
		return int64(x), nil
	case int64:
		return x, nil
	case string:
		return strconv.ParseInt(x, 10, 64)
	default:
		return 0, fmt.Errorf("market: cannot parse int from %T", v)
	}
}
