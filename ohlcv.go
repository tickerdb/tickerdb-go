package tickerdb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// OHLCVOptions configures an OHLCV request.
type OHLCVOptions struct {
	// Start restricts bars to this date or later (YYYY-MM-DD). Defaults to
	// the earliest date allowed by the account's plan history window.
	Start *string

	// End restricts bars to this date or earlier (YYYY-MM-DD).
	End *string

	// Cursor fetches the next page of results. Use the NextCursor value from
	// the previous OHLCVResponse. Cannot be combined with Start.
	Cursor *string

	// Order controls the sort direction of returned bars: "asc" (oldest first)
	// or "desc" (newest first, default).
	Order *string

	// Limit caps the number of bars returned per request (1–1000, default 100).
	// The server costs credits per 100 bars (minimum 1 credit per request).
	Limit *int
}

// OHLCVBar is a single daily OHLCV candlestick.
type OHLCVBar struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// OHLCVResponse is the response from the OHLCV endpoint.
type OHLCVResponse struct {
	Ticker     string  `json:"ticker"`
	AssetClass string  `json:"asset_class"`
	Currency   *string `json:"currency"`

	// Timeframe is always "daily".
	Timeframe string `json:"timeframe"`

	// DataStatus is always "eod" (end-of-day).
	DataStatus string `json:"data_status"`

	// Adjustment is "split_and_dividend_adjusted" for stocks/ETFs, "none" for crypto.
	Adjustment string `json:"adjustment"`

	// Order is the sort direction of Bars: "asc" or "desc".
	Order string `json:"order"`

	// Start is the effective start date used for the query (YYYY-MM-DD).
	Start string `json:"start"`

	// End is the end date filter if one was requested (YYYY-MM-DD).
	End *string `json:"end"`

	// RowCount is the number of bars in this response.
	RowCount int `json:"row_count"`

	// HasMore is true when additional bars are available. Fetch them by passing
	// NextCursor as OHLCVOptions.Cursor in the next request.
	HasMore bool `json:"has_more"`

	// NextCursor is the date value to use as Cursor on the next request.
	// Nil when HasMore is false.
	NextCursor *string `json:"next_cursor"`

	// Bars contains the OHLCV candlesticks for this page.
	Bars []OHLCVBar `json:"bars"`

	// PlanHistoryDays is the number of history days the account plan allows.
	PlanHistoryDays int `json:"plan_history_days"`

	// Plan is the display name of the account's current plan.
	Plan string `json:"plan"`

	RateLimits RateLimits `json:"-"`
}

// OHLCV retrieves daily OHLCV bars for a single ticker.
//
// Bars are split-and-dividend adjusted for stocks and ETFs; crypto bars are
// unadjusted. Results are paginated: check HasMore and pass NextCursor as
// OHLCVOptions.Cursor to fetch subsequent pages.
//
// Credit cost: 1 credit per 100 bars (minimum 1 credit per request).
func (c *Client) OHLCV(ctx context.Context, ticker string, opts *OHLCVOptions) (*OHLCVResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Start != nil {
			params.Set("start", *opts.Start)
		}
		if opts.End != nil {
			params.Set("end", *opts.End)
		}
		if opts.Cursor != nil {
			params.Set("cursor", *opts.Cursor)
		}
		if opts.Order != nil {
			params.Set("order", *opts.Order)
		}
		if opts.Limit != nil {
			params.Set("limit", strconv.Itoa(*opts.Limit))
		}
	}

	resp := &OHLCVResponse{}
	rateLimits, err := c.doGet(ctx, "/ohlcv/"+url.PathEscape(ticker), params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// OHLCVAll fetches all available OHLCV bars for a ticker by following the
// cursor until HasMore is false, and returns the concatenated slice.
//
// Start, End, Order, and Limit from opts are forwarded to each page request.
// Do not set Cursor in opts — OHLCVAll manages pagination internally.
//
// Each page consumes credits (1 per 100 bars, minimum 1 per request).
// The call respects ctx cancellation between pages. A hard cap of 500 pages
// (~500 000 bars at default limit) prevents runaway loops.
func (c *Client) OHLCVAll(ctx context.Context, ticker string, opts *OHLCVOptions) ([]OHLCVBar, error) {
	const maxPages = 500

	// Shallow-copy opts so we can set Cursor without mutating the caller's struct.
	var base OHLCVOptions
	if opts != nil {
		base = *opts
	}
	base.Cursor = nil // managed internally

	var all []OHLCVBar
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return all, err
		}

		resp, err := c.OHLCV(ctx, ticker, &base)
		if err != nil {
			return all, err
		}
		all = append(all, resp.Bars...)

		if !resp.HasMore || resp.NextCursor == nil {
			return all, nil
		}
		base.Cursor = resp.NextCursor
		base.Start = nil // cursor and start are mutually exclusive
	}

	return all, fmt.Errorf("tickerdb: OHLCVAll exceeded %d page limit for %s", maxPages, ticker)
}
