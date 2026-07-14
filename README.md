# TickerDB - Pre-computed market data for agents.

[![Go Reference](https://pkg.go.dev/badge/github.com/tickerdb/tickerdb-go.svg)](https://pkg.go.dev/github.com/tickerdb/tickerdb-go)

Connect your agent to hundreds of indicators like trend_direction, support_level, and analyst_consensus to improve reasoning and reduce token usage.

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
	"fmt"
	"log"

	tickerdb "github.com/tickerdb/tickerdb-go"
)

func main() {
	client := tickerdb.NewClient("tdb_your_api_key")

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
client := tickerdb.NewClient("tdb_your_api_key",
	tickerdb.WithBaseURL("https://custom-api.example.com/v1"),
)

// Custom HTTP client (e.g., with timeout)
httpClient := &http.Client{Timeout: 30 * time.Second}
client := tickerdb.NewClient("tdb_your_api_key",
	tickerdb.WithHTTPClient(httpClient),
)
```

## Endpoints

### Account

Get the authenticated account's plan, limits, and usage.

```go
resp, err := client.Account(ctx)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Plan:      %s\n", resp.TierFull)
fmt.Printf("Email:     %s\n", resp.Email)
fmt.Printf("Credits:   %.2f\n", resp.Usage.CreditBalance)
fmt.Printf("Requests:  %d / %d this month\n",
	resp.Usage.MonthlyRequestsUsed, resp.Limits.MonthlyRequests)
fmt.Printf("History:   %d days\n", resp.Limits.HistoryDays)
```

### OHLCV

Fetch daily or weekly OHLCV bars for a ticker. Each request consumes **1 credit per 100 bars** (rounded up).

```go
resp, err := client.OHLCV(ctx, "AAPL", nil)
if err != nil {
	log.Fatal(err)
}

for _, bar := range resp.Bars {
	fmt.Printf("%s  O:%.2f H:%.2f L:%.2f C:%.2f V:%.0f\n",
		bar.Date, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
}
fmt.Printf("Credits remaining: %.2f\n", resp.RateLimits.CreditBalance)
```

With options:

```go
resp, err := client.OHLCV(ctx, "AAPL", &tickerdb.OHLCVOptions{
	Start: tickerdb.Ptr("2024-01-01"),
	End:   tickerdb.Ptr("2024-12-31"),
	Order: tickerdb.Ptr("asc"),
	Limit: tickerdb.Ptr(100),
})
```

#### Cursor Pagination

`OHLCV` returns up to 500 bars per page. Use `OHLCVAll` to transparently collect all pages:

```go
// OHLCVAll follows next_cursor automatically (caps at 500 pages).
bars, err := client.OHLCVAll(ctx, "AAPL", &tickerdb.OHLCVOptions{
	Start: tickerdb.Ptr("2020-01-01"),
})
fmt.Printf("Fetched %d total bars\n", len(bars))
```

Or paginate manually:

```go
var allBars []tickerdb.OHLCVBar
opts := &tickerdb.OHLCVOptions{Start: tickerdb.Ptr("2020-01-01")}

for {
	page, err := client.OHLCV(ctx, "AAPL", opts)
	if err != nil {
		log.Fatal(err)
	}
	allBars = append(allBars, page.Bars...)
	if !page.HasMore {
		break
	}
	opts.Cursor = page.NextCursor
}
```

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

Summary payloads are forward-compatible JSON. Current snapshots include top-level freshness like `as_of_date`, same-candle `ohlcv.open/high/low/close/volume`, categorical sections such as `trend`, `momentum`, and `volume`, semantic MA fields such as `trend.ma_slopes` (`ma_8` through `ma_200`), `trend.ma_compression_band`, and `trend.ma_crossover_event`, support/resistance prices, and tier-gated fundamentals/sector context when available.

Summary stays band-first by default, so sibling `_meta` / `status_meta` stability objects are omitted unless you opt in:

```go
resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Meta: tickerdb.Ptr(true),
})

resp, err = client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Fields: `["trend.direction","trend.direction_meta"]`,
})
```

#### Summary with Date Range

```go
resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Start: tickerdb.Ptr("2025-01-01"),
	End:   tickerdb.Ptr("2025-03-31"),
})
```

#### Summary with Events Filter

```go
resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Field: tickerdb.Ptr("momentum_rsi_zone"),
	Band:  tickerdb.Ptr("deep_oversold"),
})
```

### Search

Search across all tickers by categorical state.

```go
resp, err := client.Search(ctx, &tickerdb.SearchOptions{
	Filters: `[{"field":"momentum_rsi_zone","op":"eq","value":"oversold"}]`,
})
```

Request a historical snapshot by passing `Date`:

```go
resp, err := client.Search(ctx, &tickerdb.SearchOptions{
	Filters: `[{"field":"trend_direction","op":"eq","value":"uptrend"}]`,
	Date:    tickerdb.Ptr("2025-01-15"),
})
```

#### Query Builder

The fluent query builder generates the filter JSON for you. Chain methods in order: `Select`, filters, `OnDate`, `Sort`, `Limit`.

```go
results, err := client.Query().
	Select("ticker", "sector", "momentum_rsi_zone").
	Eq("momentum_rsi_zone", "oversold").
	Eq("sector", "Technology").
	OnDate("2025-01-15").
	Sort("extremes_condition_percentile", "asc").
	Limit(10).
	Execute(ctx)
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

Add or remove tickers:

```go
resp, err := client.AddToWatchlist(ctx, []string{"AAPL", "MSFT", "TSLA"})
resp, err := client.RemoveFromWatchlist(ctx, []string{"TSLA"})
```

### Watchlist Changes

Get field-level state changes for your saved watchlist tickers.

```go
resp, err := client.WatchlistChanges(ctx, nil)

resp, err := client.WatchlistChanges(ctx, &tickerdb.WatchlistChangesOptions{
	Timeframe: tickerdb.Ptr(tickerdb.TimeframeWeekly),
})
```

### Webhooks

Manage webhooks that fire after each pipeline run.

```go
// Create
wh, err := client.CreateWebhook(ctx, tickerdb.CreateWebhookRequest{
	URL: "https://example.com/hooks/tickerdb",
	Events: tickerdb.WebhookEvents{
		tickerdb.WebhookEventWatchlistChanges: true,
		tickerdb.WebhookEventDataReady:        false,
	},
})
fmt.Println(wh.Secret) // store securely — shown once

// List / update / delete
list, err := client.ListWebhooks(ctx)
_, err = client.UpdateWebhook(ctx, tickerdb.UpdateWebhookRequest{ID: wh.ID, Active: tickerdb.Ptr(false)})
_, err = client.DeleteWebhook(ctx, wh.ID)
```

#### Webhook Deliveries

Inspect delivery history for all webhooks on the account, or filter to one:

```go
// All deliveries
resp, err := client.WebhookDeliveries(ctx, nil)

// Filtered by webhook
resp, err := client.WebhookDeliveries(ctx, &tickerdb.WebhookDeliveriesOptions{
	WebhookID: tickerdb.Ptr(wh.ID),
	Limit:     tickerdb.Ptr(20),
})

for _, d := range resp.Deliveries {
	fmt.Printf("%s  %s  status=%s\n", d.RunDate, d.EventType, d.Status)
}
```

Webhook event type constants:

```go
tickerdb.WebhookEventWatchlistChanges // "watchlist.changes"
tickerdb.WebhookEventDataReady        // "data.ready"
```

### Team

Manage team membership (Business plan).

```go
// List teams and incoming invites
list, err := client.ListTeams(ctx)
for _, t := range list.Teams {
	fmt.Printf("%s  seats %d/%d  your role: %s\n",
		t.Name, t.SeatsUsed, t.MaxSeats, t.YourRole)
}

// Create a team
r, err := client.CreateTeam(ctx, "Acme Trading")
fmt.Println(r.Team.ID)

// Invite a member
r, err = client.InviteTeamMember(ctx, teamID, "alice@example.com", "member")
fmt.Println(r.Invite.ExpiresAt)

// Manage members
_, err = client.PromoteTeamMember(ctx, teamID, userID, "admin")
_, err = client.RemoveTeamMember(ctx, teamID, userID)

// Manage invites
_, err = client.ResendTeamInvite(ctx, teamID, inviteID)
_, err = client.CancelTeamInvite(ctx, teamID, inviteID)

// Team settings
_, err = client.RenameTeam(ctx, teamID, "Acme Quant")
_, err = client.SetTeamSeats(ctx, teamID, 10)

// Leave a team (non-owners only)
_, err = client.LeaveTeam(ctx, teamID)
```

## Band Stability Metadata

Set `Meta: tickerdb.Ptr(true)` in `SummaryOptions` to include full stability metadata across the response, or request specific `*_meta` paths via `Fields`.

```go
// tickerdb.Stability  — "fresh" | "holding" | "established" | "volatile"
// tickerdb.BandMeta   — stability, periods_in_current_state, flips_recent, flips_lookback
// Stability metadata is available on Plus and Pro tiers only.

resp, err := client.Summary(ctx, "AAPL", &tickerdb.SummaryOptions{
	Meta: tickerdb.Ptr(true),
})
// resp.Data contains _meta objects next to each band field, e.g.:
// "direction": "uptrend",
// "direction_meta": { "stability": "established", "periods_in_current_state": 18, ... }
```

Stability context also appears in **Watchlist** (always included) and in **Watchlist Changes** (per changed band).

## Error Handling

All API errors are returned as `*tickerdb.APIError`.

```go
import "errors"

_, err := client.OHLCV(ctx, "AAPL", nil)
if err != nil {
	var apiErr *tickerdb.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("Status:  %d\n", apiErr.StatusCode)
		fmt.Printf("Type:    %s\n", apiErr.Type)
		fmt.Printf("Message: %s\n", apiErr.Message)

		switch {
		case apiErr.IsRateLimitError():
			if rt, ok := apiErr.ResetTime(); ok {
				fmt.Printf("Resets at: %s\n", rt.Format(time.RFC3339))
			}
		case apiErr.IsForbiddenError():
			fmt.Printf("Upgrade at: %s\n", apiErr.UpgradeURL)
		case apiErr.IsPaymentRequiredError():
			fmt.Println("Payment required.")
		case apiErr.IsAuthError():
			fmt.Println("Check your API key.")
		}

		// Credit shortfall (OHLCV requests that exceed credit balance)
		if apiErr.CreditsRequired != nil {
			fmt.Printf("Need %d credits, have %d\n",
				*apiErr.CreditsRequired, *apiErr.CreditsRemaining)
		}
	} else {
		fmt.Printf("Network error: %v\n", err)
	}
}
```

Helper methods on `*APIError`:

| Method | Status code |
|---|---|
| `IsAuthError()` | 401 |
| `IsForbiddenError()` | 403 |
| `IsNotFoundError()` | 404 |
| `IsRateLimitError()` | 429 |
| `IsPaymentRequiredError()` | 402 |

## Rate Limits

Every response includes rate limit information parsed from response headers:

```go
resp, err := client.OHLCV(ctx, "AAPL", nil)
if err != nil {
	log.Fatal(err)
}

rl := resp.RateLimits
fmt.Printf("Requests: %d / %d (resets %s)\n",
	rl.RequestsUsed, rl.RequestLimit, rl.RequestReset.Format(time.RFC3339))
fmt.Printf("Credits remaining: %.2f\n", rl.CreditBalance)
```

`RateLimits` fields:

| Field | Header | Notes |
|---|---|---|
| `RequestLimit` | `X-Request-Limit` | Monthly request cap |
| `RequestsUsed` | `X-Requests-Used` | Requests consumed this month |
| `RequestsRemaining` | `X-Requests-Remaining` | Requests left this month |
| `RequestReset` | `X-Request-Reset` | When the monthly counter resets |
| `CreditBalance` | `X-Credit-Balance` | Remaining OHLCV credits |

## License

MIT
