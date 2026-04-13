# TickerDB Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/tickerdb/tickerdb-go.svg)](https://pkg.go.dev/github.com/tickerdb/tickerdb-go)

Official Go SDK for the [TickerDB](https://tickerdb.com) financial data API.

- **API Docs:** <https://tickerdb.com/docs>
- **Website:** <https://tickerdb.com>

## Installation

```bash
go get github.com/tickerdb/tickerdb-go
```

Requires Go 1.21+ and uses only the standard library (zero external dependencies).

## Quick Start

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tickerdb "github.com/tickerdb/tickerdb-go"
)

func main() {
	client := tickerdb.NewClient("YOUR_API_KEY")

	resp, err := client.Summary(context.Background(), "AAPL", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(resp.Data))
	fmt.Println("snapshot date is inside resp.Data as as_of_date")
	fmt.Printf("Requests remaining: %d\n", resp.RateLimits.RequestsRemaining)
}
```

## Client Configuration

```go
// Custom base URL
client := tickerdb.NewClient("YOUR_API_KEY",
	tickerdb.WithBaseURL("https://custom-api.example.com/v1"),
)

// Custom HTTP client (e.g., with timeout)
httpClient := &http.Client{Timeout: 30 * time.Second}
client := tickerdb.NewClient("YOUR_API_KEY",
	tickerdb.WithHTTPClient(httpClient),
)
```

## Endpoints

### Summary

Get a technical analysis summary for a single ticker.

```go
// Basic usage
resp, err := client.Summary(ctx, "AAPL", nil)

// With options
resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeWeekly),
	Date:      tickerdb.Ptr("2025-01-15"),
})
```

### Summary with Date Range

Get a summary series for one ticker across a date range by passing `Start` and `End`.

```go
resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Start: tickerdb.Ptr("2025-01-01"),
	End:   tickerdb.Ptr("2025-03-31"),
})
```

### Summary with Events Filter

Query event occurrences for a specific band field.

```go
resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Field: tickerdb.Ptr("momentum_rsi_zone"),
	Band:  tickerdb.Ptr("deep_oversold"),
})

resp, err = client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Field: tickerdb.Ptr("extremes_condition"),
	Band:  tickerdb.Ptr("deep_oversold"),
})
```

### Watchlist

Get the saved watchlist snapshot for the authenticated account.

```go
resp, err := client.Watchlist(ctx, nil)
fmt.Println(resp.AsOfDate)

resp, err := client.Watchlist(ctx, &tickerdb.WatchlistOptions{
	Date: tickerdb.Ptr("2025-01-15"),
})
```

Add tickers to the saved watchlist:

```go
resp, err := client.AddToWatchlist(ctx, []string{"AAPL", "MSFT", "TSLA"})
```

Remove tickers from the saved watchlist:

```go
resp, err := client.RemoveFromWatchlist(ctx, []string{"TSLA"})
```

### Watchlist Changes

Get field-level state changes for your saved watchlist tickers since the last pipeline run.

```go
resp, err := client.WatchlistChanges(ctx, nil)

resp, err := client.WatchlistChanges(ctx, &tickerdb.WatchlistChangesOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeWeekly),
})
```

## Band Stability Metadata

Every band field (trend direction, momentum zone, etc.) now includes a sibling `_meta` object with stability context. This tells you how long a state has been held, how often it has flipped recently, and an overall stability label.

Watchlist responses also expose a top-level `AsOfDate` field so clients can see which session date the compact snapshot represents.

```go
// New types for stability metadata
// tickerdb.Stability     — "fresh" | "holding" | "established" | "volatile"
// tickerdb.BandMeta      — full metadata (stability, periods_in_current_state, flips_recent, flips_lookback)
// Stability metadata is available on Plus and Pro tiers only.

// Example: unmarshal summary data and inspect stability
resp, err := client.Summary(ctx, "AAPL", nil)
// resp.Data contains _meta objects next to each band field, e.g.:
// "direction": "uptrend",
// "direction_meta": { "stability": "established", "periods_in_current_state": 18, ... }
```

Stability context also appears in **Watchlist Changes**, which include stability fields for each changed band.

### Query Builder

The SDK includes a fluent query builder for searching assets by categorical state. Chain methods in order: Select, filters, Sort, Limit.

```go
results, err := client.Query().
    Select("ticker", "sector", "momentum_rsi_zone").
    Eq("momentum_rsi_zone", "oversold").
    Eq("sector", "Technology").
    Sort("extremes_condition_percentile", "asc").
    Limit(10).
    Execute(ctx)
```

## Working with Responses

All response structs contain a `Data` field of type `json.RawMessage` and a `RateLimits` field. You can unmarshal `Data` into your own structs:

```go
resp, err := client.Summary(ctx, "AAPL", nil)
if err != nil {
	log.Fatal(err)
}

var result map[string]interface{}
json.Unmarshal(resp.Data, &result)
```

## Error Handling

All API errors are returned as `*tickerdb.APIError`, which implements the `error` interface.

```go
import "errors"

resp, err := client.Summary(ctx, "INVALID", nil)
if err != nil {
	var apiErr *tickerdb.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("Status: %d\n", apiErr.StatusCode)
		fmt.Printf("Type: %s\n", apiErr.Type)
		fmt.Printf("Message: %s\n", apiErr.Message)

		if apiErr.IsRateLimitError() {
			fmt.Println("Rate limited! Try again later.")
		}
		if apiErr.IsForbiddenError() {
			fmt.Printf("Upgrade at: %s\n", apiErr.UpgradeURL)
		}
	} else {
		// Network or other non-API error
		fmt.Printf("Error: %v\n", err)
	}
}
```

## Rate Limits

Every response includes rate limit information parsed from response headers:

```go
resp, err := client.Summary(ctx, "AAPL", nil)
if err != nil {
	log.Fatal(err)
}

rl := resp.RateLimits
fmt.Printf("Requests: %d/%d (resets %s)\n",
	rl.RequestsUsed, rl.RequestLimit, rl.RequestReset)
fmt.Printf("Hourly: %d/%d (resets %s)\n",
	rl.HourlyRequestsUsed, rl.HourlyRequestLimit, rl.HourlyRequestReset)
```

## License

MIT
