# TickerDB Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/tickerdb/tickerdb-sdk-go.svg)](https://pkg.go.dev/github.com/tickerdb/tickerdb-sdk-go)

Official Go SDK for the [TickerDB](https://tickerdb.com) financial data API.

- **API Docs:** <https://tickerdb.com/docs>
- **Website:** <https://tickerdb.com>

## Installation

```bash
go get github.com/tickerdb/tickerdb-sdk-go
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

	tickerdb "github.com/tickerdb/tickerdb-sdk-go"
)

func main() {
	client := tickerdb.NewClient("YOUR_API_KEY")

	resp, err := client.Summary(context.Background(), "AAPL", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(resp.Data))
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

### History

Get a historical series for one ticker across a date range.

```go
resp, err := client.History(ctx, "AAPL", &tickerdb.HistoryOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeDaily),
	Start:     "2025-01-01",
	End:       "2025-03-31",
})
```

### Compare

Compare multiple tickers side by side.

```go
resp, err := client.Compare(ctx, []string{"AAPL", "MSFT", "GOOGL"}, nil)

resp, err := client.Compare(ctx, []string{"AAPL", "MSFT"}, &tickerdb.CompareOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeDaily),
})
```

### Watchlist

Analyze a list of tickers (POST request).

```go
resp, err := client.Watchlist(ctx, []string{"AAPL", "MSFT", "TSLA"}, nil)

resp, err := client.Watchlist(ctx, []string{"AAPL", "MSFT"}, &tickerdb.WatchlistOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeWeekly),
})
```

### Watchlist Changes

Get field-level state changes for your saved watchlist tickers since the last pipeline run.

```go
resp, err := client.WatchlistChanges(ctx, nil)

resp, err := client.WatchlistChanges(ctx, &tickerdb.WatchlistChangesOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeWeekly),
})
```

### Assets

List all available assets.

```go
resp, err := client.Assets(ctx)
```

### Scan: Oversold

Find oversold assets.

```go
resp, err := client.ScanOversold(ctx, &tickerdb.ScanOversoldOptions{
	AssetClass:  tickerdb.Ptr(tickerdb.AssetClassStock),
	MinSeverity: tickerdb.Ptr(tickerdb.OversoldSeverityDeepOversold),
	SortBy:      tickerdb.Ptr("severity"),
	Limit:       tickerdb.Ptr(10),
})
```

### Scan: Breakouts

Find assets with breakout patterns.

```go
resp, err := client.ScanBreakouts(ctx, &tickerdb.ScanBreakoutsOptions{
	Direction:  tickerdb.Ptr(tickerdb.DirectionBullish),
	AssetClass: tickerdb.Ptr(tickerdb.AssetClassStock),
	SortBy:     tickerdb.Ptr("volume_ratio"),
	Limit:      tickerdb.Ptr(20),
})
```

### Scan: Unusual Volume

Find assets with unusual trading volume.

```go
resp, err := client.ScanUnusualVolume(ctx, &tickerdb.ScanUnusualVolumeOptions{
	MinRatioBand: tickerdb.Ptr(tickerdb.VolumeRatioBandHigh),
	AssetClass:   tickerdb.Ptr(tickerdb.AssetClassStock),
	Limit:        tickerdb.Ptr(15),
})
```

### Scan: Valuation

Find undervalued or overvalued assets.

```go
resp, err := client.ScanValuation(ctx, &tickerdb.ScanValuationOptions{
	Direction:   tickerdb.Ptr(tickerdb.DirectionUndervalued),
	MinSeverity: tickerdb.Ptr(tickerdb.ValuationSeverityDeepValue),
	Sector:      tickerdb.Ptr("Technology"),
	Limit:       tickerdb.Ptr(10),
})
```

### Scan: Insider Activity

Find assets with notable insider trading.

```go
resp, err := client.ScanInsiderActivity(ctx, &tickerdb.ScanInsiderActivityOptions{
	Direction: tickerdb.Ptr(tickerdb.DirectionBuying),
	SortBy:    tickerdb.Ptr("zone_severity"),
	Limit:     tickerdb.Ptr(10),
})
```

## Band Stability Metadata

Every band field (trend direction, momentum zone, etc.) now includes a sibling `_meta` object with stability context. This tells you how long a state has been held, how often it has flipped recently, and an overall stability label.

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

Stability context also appears in related endpoints:

- **Watchlist Changes** include stability fields for each changed band.
- **Scanners** return `*_stability` and `*_flips_recent` columns for relevant bands.
- **Events** include `StabilityAtEntry`, `FlipsRecentAtEntry`, and `FlipsLookback` on each entry.

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
