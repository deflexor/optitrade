package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// PublicClient performs unsigned OKX v5 public REST requests.
type PublicClient struct {
	BaseURL string
	HTTP    *http.Client
}

// InstrumentSummary is a subset of OKX /api/v5/public/instruments row for options.
type InstrumentSummary struct {
	InstId  string `json:"instId"`
	ExpTime string `json:"expTime"`
}

func (c *PublicClient) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *PublicClient) base() string {
	base := strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		return defaultBaseURL
	}
	return base
}

type apiEnvelope struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg,omitempty"`
	Data json.RawMessage `json:"data"`
}

type marketBookRow struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

// GetOrderBookDeribit fetches OKX market books and maps the first depth row into deribit.OrderBook.
func (c *PublicClient) GetOrderBookDeribit(ctx context.Context, instId string, sz int) (deribit.OrderBook, error) {
	if strings.TrimSpace(instId) == "" {
		return deribit.OrderBook{}, fmt.Errorf("okx public: empty instId")
	}
	if sz <= 0 {
		return deribit.OrderBook{}, fmt.Errorf("okx public: sz must be positive")
	}
	u, err := url.Parse(c.base() + "/api/v5/market/books")
	if err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books url: %w", err)
	}
	q := u.Query()
	q.Set("instId", instId)
	q.Set("sz", strconv.Itoa(sz))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books http %d: %s", resp.StatusCode, truncateBody(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books json envelope: %w", err)
	}
	if env.Code != "0" {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books api code %q msg %q", env.Code, env.Msg)
	}

	var rows []marketBookRow
	if err := json.Unmarshal(env.Data, &rows); err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books data: %w", err)
	}
	if len(rows) == 0 {
		return deribit.OrderBook{}, fmt.Errorf("okx public: books empty data")
	}

	bids, err := parseOKXBookSide(rows[0].Bids)
	if err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: bids: %w", err)
	}
	asks, err := parseOKXBookSide(rows[0].Asks)
	if err != nil {
		return deribit.OrderBook{}, fmt.Errorf("okx public: asks: %w", err)
	}

	return deribit.OrderBook{
		InstrumentName: instId,
		Bids:           bids,
		Asks:           asks,
	}, nil
}

func parseOKXBookSide(levels [][]string) ([]deribit.PriceLevel, error) {
	out := make([]deribit.PriceLevel, 0, len(levels))
	for i, tup := range levels {
		if len(tup) < 2 {
			return nil, fmt.Errorf("level %d: want at least [price,size]", i)
		}
		price, err := strconv.ParseFloat(tup[0], 64)
		if err != nil {
			return nil, fmt.Errorf("level %d price %q: %w", i, tup[0], err)
		}
		amt, err := strconv.ParseFloat(tup[1], 64)
		if err != nil {
			return nil, fmt.Errorf("level %d size %q: %w", i, tup[1], err)
		}
		out = append(out, deribit.PriceLevel{Price: price, Amount: amt})
	}
	return out, nil
}

// GetInstruments calls GET /api/v5/public/instruments with instType and uly (e.g. OPTION, BTC-USD).
func (c *PublicClient) GetInstruments(ctx context.Context, instType, uly string) ([]InstrumentSummary, error) {
	instType = strings.TrimSpace(instType)
	if instType == "" {
		return nil, fmt.Errorf("okx public: empty instType")
	}
	uly = strings.TrimSpace(uly)
	if uly == "" {
		return nil, fmt.Errorf("okx public: empty uly")
	}
	u, err := url.Parse(c.base() + "/api/v5/public/instruments")
	if err != nil {
		return nil, fmt.Errorf("okx public: instruments url: %w", err)
	}
	q := u.Query()
	q.Set("instType", instType)
	q.Set("uly", uly)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("okx public: instruments request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("okx public: instruments http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("okx public: instruments read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("okx public: instruments http %d: %s", resp.StatusCode, truncateBody(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("okx public: instruments json envelope: %w", err)
	}
	if env.Code != "0" {
		return nil, fmt.Errorf("okx public: instruments api code %q msg %q", env.Code, env.Msg)
	}

	var rows []InstrumentSummary
	if err := json.Unmarshal(env.Data, &rows); err != nil {
		return nil, fmt.Errorf("okx public: instruments data: %w", err)
	}
	return rows, nil
}

// GetIndexPrice calls GET /api/v5/market/index-tickers?instId=... and returns idxPx.
func (c *PublicClient) GetIndexPrice(ctx context.Context, instId string) (float64, error) {
	instId = strings.TrimSpace(instId)
	if instId == "" {
		return 0, fmt.Errorf("okx public: empty index instId")
	}
	u, err := url.Parse(c.base() + "/api/v5/market/index-tickers")
	if err != nil {
		return 0, fmt.Errorf("okx public: index url: %w", err)
	}
	q := u.Query()
	q.Set("instId", instId)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("okx public: index request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return 0, fmt.Errorf("okx public: index http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return 0, fmt.Errorf("okx public: index read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("okx public: index http %d: %s", resp.StatusCode, truncateBody(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return 0, fmt.Errorf("okx public: index json envelope: %w", err)
	}
	if env.Code != "0" {
		return 0, fmt.Errorf("okx public: index api code %q msg %q", env.Code, env.Msg)
	}

	var rows []struct {
		IdxPx string `json:"idxPx"`
	}
	if err := json.Unmarshal(env.Data, &rows); err != nil {
		return 0, fmt.Errorf("okx public: index data: %w", err)
	}
	if len(rows) == 0 || strings.TrimSpace(rows[0].IdxPx) == "" {
		return 0, fmt.Errorf("okx public: index empty data")
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(rows[0].IdxPx), 64)
	if err != nil || v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("okx public: index bad idxPx %q", rows[0].IdxPx)
	}
	return v, nil
}

func truncateBody(b []byte) string {
	const max = 512
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
