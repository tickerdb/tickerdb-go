package tickerdb

import "encoding/json"

// Timeframe represents the analysis timeframe.
type Timeframe string

const (
	TimeframeDaily  Timeframe = "daily"
	TimeframeWeekly Timeframe = "weekly"
)

// AssetClass represents the type of financial asset.
type AssetClass string

const (
	AssetClassStock  AssetClass = "stock"
	AssetClassCrypto AssetClass = "crypto"
	AssetClassETF    AssetClass = "etf"
	AssetClassAll    AssetClass = "all"
)

// Direction represents bullish/bearish or buying/selling direction.
type Direction string

const (
	DirectionBullish Direction = "bullish"
	DirectionBearish Direction = "bearish"
	DirectionAll     Direction = "all"

	DirectionBuying  Direction = "buying"
	DirectionSelling Direction = "selling"

	DirectionUndervalued Direction = "undervalued"
	DirectionOvervalued  Direction = "overvalued"
)

// OversoldSeverity represents the minimum severity for oversold scans.
type OversoldSeverity string

const (
	OversoldSeverityOversold     OversoldSeverity = "oversold"
	OversoldSeverityDeepOversold OversoldSeverity = "deep_oversold"
)

// ValuationSeverity represents the minimum severity for valuation scans.
type ValuationSeverity string

const (
	ValuationSeverityDeepValue      ValuationSeverity = "deep_value"
	ValuationSeverityDeeplyOvervalued ValuationSeverity = "deeply_overvalued"
)

// VolumeRatioBand represents volume ratio classification.
type VolumeRatioBand string

const (
	VolumeRatioBandExtremelyLow  VolumeRatioBand = "extremely_low"
	VolumeRatioBandLow           VolumeRatioBand = "low"
	VolumeRatioBandNormal        VolumeRatioBand = "normal"
	VolumeRatioBandAboveAverage  VolumeRatioBand = "above_average"
	VolumeRatioBandHigh          VolumeRatioBand = "high"
	VolumeRatioBandExtremelyHigh VolumeRatioBand = "extremely_high"
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
type SummaryOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Date      *string    `json:"date,omitempty"`
}

// HistoryOptions configures the History request.
type HistoryOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Start     string     `json:"start"`
	End       string     `json:"end"`
}

// CompareOptions configures the Compare request.
type CompareOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Date      *string    `json:"date,omitempty"`
}

// WatchlistOptions configures the Watchlist request.
type WatchlistOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
}

// ListEventsOptions configures optional parameters for ListEvents.
// The required ticker and field are passed as separate arguments.
type ListEventsOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Band      *string    `json:"band,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
	Before    *string    `json:"before,omitempty"`
	After     *string    `json:"after,omitempty"`
	// Cross-asset correlation: a second ticker to filter against (e.g. "SPY").
	// Requires ContextField and ContextBand. Plus/Pro only. Costs 2 credits.
	ContextTicker *string `json:"context_ticker,omitempty"`
	// Band field to check on the context ticker (e.g. "trend_direction").
	ContextField *string `json:"context_field,omitempty"`
	// Only return events where the context ticker was in this band on the event date.
	ContextBand *string `json:"context_band,omitempty"`
}

// EventEntry represents a single band transition event.
type EventEntry struct {
	Date               string                           `json:"date"`
	Band               string                           `json:"band"`
	PrevBand           string                           `json:"prev_band"`
	StabilityAtEntry   *Stability                       `json:"stability_at_entry"`
	FlipsRecentAtEntry *int                             `json:"flips_recent_at_entry,omitempty"`
	FlipsLookback      *string                          `json:"flips_lookback,omitempty"`
	DurationDays       *int                             `json:"duration_days,omitempty"`
	DurationWeeks      *int                             `json:"duration_weeks,omitempty"`
	Aftermath          map[string]*AftermathPerformance `json:"aftermath"`
}

// AftermathPerformance holds the performance band for an aftermath window.
type AftermathPerformance struct {
	Performance string `json:"performance"`
}

// EventsContext describes the cross-asset correlation filter applied.
type EventsContext struct {
	Ticker string `json:"ticker"`
	Field  string `json:"field"`
	Band   string `json:"band"`
}

// ListEventsResponse is returned by ListEvents.
type ListEventsResponse struct {
	Ticker           string         `json:"ticker"`
	Field            string         `json:"field"`
	Timeframe        string         `json:"timeframe"`
	Events           []EventEntry   `json:"events"`
	TotalOccurrences int            `json:"total_occurrences"`
	QueryRange       string         `json:"query_range"`
	Context          *EventsContext `json:"context,omitempty"`
	RateLimits       RateLimits     `json:"-"`
}

// ScanOversoldOptions configures the ScanOversold request.
type ScanOversoldOptions struct {
	Timeframe   *Timeframe        `json:"timeframe,omitempty"`
	AssetClass  *AssetClass       `json:"asset_class,omitempty"`
	Sector      *string           `json:"sector,omitempty"`
	MinSeverity *OversoldSeverity `json:"min_severity,omitempty"`
	SortBy      *string           `json:"sort_by,omitempty"` // "severity", "days_oversold", "condition_percentile"
	Limit       *int              `json:"limit,omitempty"`   // 1-50
	Date        *string           `json:"date,omitempty"`
}

// ScanBreakoutsOptions configures the ScanBreakouts request.
type ScanBreakoutsOptions struct {
	Timeframe  *Timeframe  `json:"timeframe,omitempty"`
	AssetClass *AssetClass `json:"asset_class,omitempty"`
	Sector     *string     `json:"sector,omitempty"`
	Direction  *Direction  `json:"direction,omitempty"` // "bullish", "bearish", "all"
	SortBy     *string     `json:"sort_by,omitempty"`   // "volume_ratio", "level_strength", "condition_percentile"
	Limit      *int        `json:"limit,omitempty"`     // 1-50
	Date       *string     `json:"date,omitempty"`
}

// ScanUnusualVolumeOptions configures the ScanUnusualVolume request.
type ScanUnusualVolumeOptions struct {
	Timeframe    *Timeframe       `json:"timeframe,omitempty"`
	AssetClass   *AssetClass      `json:"asset_class,omitempty"`
	Sector       *string          `json:"sector,omitempty"`
	MinRatioBand *VolumeRatioBand `json:"min_ratio_band,omitempty"`
	SortBy       *string          `json:"sort_by,omitempty"` // "volume_percentile"
	Limit        *int             `json:"limit,omitempty"`   // 1-50
	Date         *string          `json:"date,omitempty"`
}

// ScanValuationOptions configures the ScanValuation request.
type ScanValuationOptions struct {
	Timeframe   *Timeframe         `json:"timeframe,omitempty"`
	Sector      *string            `json:"sector,omitempty"`
	Direction   *Direction         `json:"direction,omitempty"`   // "undervalued", "overvalued", "all"
	MinSeverity *ValuationSeverity `json:"min_severity,omitempty"` // "deep_value", "deeply_overvalued"
	SortBy      *string            `json:"sort_by,omitempty"`     // "valuation_percentile", "pe_vs_history"
	Limit       *int               `json:"limit,omitempty"`       // 1-50
	Date        *string            `json:"date,omitempty"`
}

// ScanInsiderActivityOptions configures the ScanInsiderActivity request.
type ScanInsiderActivityOptions struct {
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	Sector    *string    `json:"sector,omitempty"`
	Direction *Direction `json:"direction,omitempty"` // "buying", "selling", "all"
	SortBy    *string    `json:"sort_by,omitempty"`   // "zone_severity", "shares_volume", "net_ratio"
	Limit     *int       `json:"limit,omitempty"`     // 1-50
	Date      *string    `json:"date,omitempty"`
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

// HistoryRow is one row returned by the History endpoint.
type HistoryRow struct {
	Date          string          `json:"date"`
	SchemaVersion string          `json:"schema_version"`
	Summary       json.RawMessage `json:"summary"`
	Levels        json.RawMessage `json:"levels"`
}

// HistoryResponse is the response from the History endpoint.
type HistoryResponse struct {
	Ticker     string       `json:"ticker"`
	Timeframe  Timeframe    `json:"timeframe"`
	Start      string       `json:"start"`
	End        string       `json:"end"`
	RowCount   int          `json:"row_count"`
	Rows       []HistoryRow `json:"rows"`
	RateLimits RateLimits   `json:"-"`
}

// CompareResponse is the response from the Compare endpoint.
type CompareResponse struct {
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

// AssetsResponse is the response from the Assets endpoint.
type AssetsResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// SectorEntry represents a single sector with its asset count.
type SectorEntry struct {
	Name       string `json:"name"`
	AssetCount int    `json:"asset_count"`
}

// SectorsResponse is the response from the ListSectors endpoint.
type SectorsResponse struct {
	Sectors      []SectorEntry `json:"sectors"`
	TotalSectors int           `json:"total_sectors"`
	RateLimits   RateLimits    `json:"-"`
}

// ScanOversoldResponse is the response from the ScanOversold endpoint.
type ScanOversoldResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// ScanBreakoutsResponse is the response from the ScanBreakouts endpoint.
type ScanBreakoutsResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// ScanUnusualVolumeResponse is the response from the ScanUnusualVolume endpoint.
type ScanUnusualVolumeResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// ScanValuationResponse is the response from the ScanValuation endpoint.
type ScanValuationResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// ScanInsiderActivityResponse is the response from the ScanInsiderActivity endpoint.
type ScanInsiderActivityResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
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
