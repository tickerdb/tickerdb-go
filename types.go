package tickerdb

import "encoding/json"

// Timeframe represents the analysis timeframe.
type Timeframe string

const (
	TimeframeDaily  Timeframe = "daily"
	TimeframeWeekly Timeframe = "weekly"
)

// Stability represents how stable/volatile a band field is.
type Stability string

const (
	StabilityFresh       Stability = "fresh"
	StabilityHolding     Stability = "holding"
	StabilityEstablished Stability = "established"
	StabilityVolatile    Stability = "volatile"
)

// BandMeta contains stability metadata for a band field (Plus/Pro tiers).
type BandMeta struct {
	Timeframe             string    `json:"timeframe"`
	PeriodsInCurrentState int       `json:"periods_in_current_state"`
	FlipsRecent           int       `json:"flips_recent"`
	FlipsLookback         string    `json:"flips_lookback"`
	Stability             Stability `json:"stability"`
}

// SummaryOptions configures the Summary request.
//
// Supports 4 modes depending on which fields are set:
//   - Snapshot (default): Current categorical state.
//   - Historical snapshot: Set Date for a point-in-time snapshot.
//   - Historical series: Set Start/End for a date range of snapshots.
//   - Events: Set Field (and optionally Band) for band transition history.
type SummaryOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Date      *string    `json:"date,omitempty"`
	// Historical series range
	Start *string `json:"start,omitempty"`
	End   *string `json:"end,omitempty"`
	// Event query parameters
	Field  *string `json:"field,omitempty"`
	Band   *string `json:"band,omitempty"`
	Limit  *int    `json:"limit,omitempty"`
	Before *string `json:"before,omitempty"`
	After  *string `json:"after,omitempty"`
	// Cross-asset correlation
	ContextTicker *string `json:"context_ticker,omitempty"`
	ContextField  *string `json:"context_field,omitempty"`
	ContextBand   *string `json:"context_band,omitempty"`
}

// SearchOptions configures the Search request.
type SearchOptions struct {
	// Filters is a JSON-encoded array of {field, op, value} objects.
	// Canonical field names come from /v1/schema/fields and use flat snake_case.
	Filters   string     `json:"filters,omitempty"`
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
	Offset    *int       `json:"offset,omitempty"`
	// Fields selects which columns to return. JSON array or comma-separated.
	// Default if omitted: ticker, asset_class, sector, performance, trend_direction,
	// momentum_rsi_zone, extremes_condition, extremes_condition_rarity, volatility_regime,
	// volume_ratio_band, fundamentals_valuation_zone, range_position.
	// Use `["*"]` for all 120+ fields. ticker is always included.
	Fields string `json:"fields,omitempty"`
	// SortBy is the column name to sort results by. Must be a valid field from the schema.
	SortBy *string `json:"sort_by,omitempty"`
	// SortDirection is the sort direction: "asc" or "desc". Defaults to "desc".
	SortDirection *string `json:"sort_direction,omitempty"`
}

// SearchResponse is the response from the Search endpoint.
type SearchResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// SchemaResponse is the response from the Schema endpoint.
type SchemaResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// WatchlistOptions configures the Watchlist request.
type WatchlistOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
}

// WatchlistRequest is the request body for the Watchlist endpoint.
type WatchlistRequest struct {
	Tickers   []string   `json:"tickers"`
	Timeframe *Timeframe `json:"timeframe,omitempty"`
}

// SummaryResponse is the response from the Summary endpoint.
type SummaryResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// WatchlistResponse is the response from the Watchlist endpoint.
type WatchlistResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// WatchlistChangesOptions configures the WatchlistChanges request.
type WatchlistChangesOptions struct {
	Timeframe *Timeframe
}

// TickerContext holds per-ticker metadata in the watchlist changes response.
type TickerContext struct {
	LastChangedDate *string `json:"last_changed_date"`
}

// WatchlistChangesResponse is the response from the WatchlistChanges endpoint.
type WatchlistChangesResponse struct {
	Timeframe      string                          `json:"timeframe"`
	RunDate        *string                         `json:"run_date"`
	Changes        map[string][]WatchlistChangeDiff `json:"changes"`
	TickerContext  map[string]TickerContext          `json:"ticker_context"`
	TickersChecked int                              `json:"tickers_checked"`
	TickersChanged int                              `json:"tickers_changed"`
	RateLimits     RateLimits                       `json:"-"`
}

// WatchlistChangeDiff represents a single field-level change.
type WatchlistChangeDiff struct {
	Field                 string      `json:"field"`
	From                  interface{} `json:"from"`
	To                    interface{} `json:"to"`
	Stability             *Stability  `json:"stability,omitempty"`
	PeriodsInCurrentState *int        `json:"periods_in_current_state,omitempty"`
	FlipsRecent           *int        `json:"flips_recent,omitempty"`
	FlipsLookback         *string     `json:"flips_lookback,omitempty"`
}

// WebhookEvents is a map of event names to enabled state.
type WebhookEvents map[string]bool

// CreateWebhookRequest is the request body for creating a webhook.
type CreateWebhookRequest struct {
	URL    string        `json:"url"`
	Events WebhookEvents `json:"events"`
}

// UpdateWebhookRequest is the request body for updating a webhook.
type UpdateWebhookRequest struct {
	ID     string        `json:"id"`
	URL    string        `json:"url,omitempty"`
	Events WebhookEvents `json:"events,omitempty"`
	Active *bool         `json:"active,omitempty"`
}

// DeleteWebhookRequest is the request body for deleting a webhook.
type DeleteWebhookRequest struct {
	ID string `json:"id"`
}

// Webhook represents a webhook configuration.
type Webhook struct {
	ID        string        `json:"id"`
	URL       string        `json:"url"`
	Events    WebhookEvents `json:"events"`
	Active    bool          `json:"active"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// WebhookCreated is the response from creating a webhook.
type WebhookCreated struct {
	ID        string        `json:"id"`
	URL       string        `json:"url"`
	Secret    string        `json:"secret"`
	Events    WebhookEvents `json:"events"`
	Active    bool          `json:"active"`
	CreatedAt string        `json:"created_at"`
}

// WebhookListResponse is the response from listing webhooks.
type WebhookListResponse struct {
	Webhooks     []Webhook `json:"webhooks"`
	WebhookCount int       `json:"webhook_count"`
	WebhookLimit int       `json:"webhook_limit"`
}

// WebhookUpdateResponse is the response from updating a webhook.
type WebhookUpdateResponse struct {
	Updated bool   `json:"updated"`
	ID      string `json:"id"`
}

// WebhookDeleteResponse is the response from deleting a webhook.
type WebhookDeleteResponse struct {
	Deleted      string `json:"deleted"`
	WebhookCount int    `json:"webhook_count"`
}

// Ptr returns a pointer to the given value. This is a convenience helper
// for constructing option structs with optional fields.
func Ptr[T any](v T) *T {
	return &v
}
