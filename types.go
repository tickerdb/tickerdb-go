package tickerdb

import "encoding/json"

// Timeframe represents the analysis timeframe.
type Timeframe string

// Supported analysis timeframes.
const (
	TimeframeDaily  Timeframe = "daily"
	TimeframeWeekly Timeframe = "weekly"
)

// Stability represents how stable/volatile a band field is.
type Stability string

// Stability values returned by metadata-enabled responses.
const (
	StabilityFresh       Stability = "fresh"
	StabilityHolding     Stability = "holding"
	StabilityEstablished Stability = "established"
	StabilityVolatile    Stability = "volatile"
)

// BandMeta contains stability metadata for a band field (Plus/Pro tiers).
// It is available in summary responses when requested.
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
	// Fields selects which summary fields to return. JSON array or comma-separated.
	// Supports sections like trend and dotted paths like trend.direction.
	Fields string `json:"fields,omitempty"`
	// Meta includes sibling _meta / status_meta stability objects in snapshot and
	// history responses. Explicit *_meta field paths still work without this flag.
	Meta *bool `json:"meta,omitempty"`
	// Date range sampling mode ("even")
	Sample *string `json:"sample,omitempty"`
	// Event query parameters
	// Field accepts canonical band fields such as trend_ma_crossover_event,
	// trend_distance_ma50, trend_direction, and momentum_rsi_zone.
	Field  *string `json:"field,omitempty"`
	Band   *string `json:"band,omitempty"`
	Limit  *int    `json:"limit,omitempty"`
	Before *string `json:"before,omitempty"`
	After  *string `json:"after,omitempty"`
	// Cross-asset correlation
	ContextTicker *string `json:"context_ticker,omitempty"`
	// ContextField uses the same canonical band field names as Field.
	ContextField *string `json:"context_field,omitempty"`
	ContextBand  *string `json:"context_band,omitempty"`
}

// SearchOptions configures the Search request.
type SearchOptions struct {
	// Filters is a JSON-encoded array of {field, op, value} objects.
	// Canonical field names come from /v1/schema/fields and use flat snake_case.
	Filters   string     `json:"filters,omitempty"`
	Timeframe *Timeframe `json:"timeframe,omitempty"`
	// Date requests a historical snapshot for the given date (YYYY-MM-DD).
	// Omit for the latest available snapshot.
	Date      *string    `json:"date,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
	Offset    *int       `json:"offset,omitempty"`
	// Fields selects which columns to return. JSON array or comma-separated.
	// Default if omitted: ticker, asset_class, sector, performance, trend_direction,
	// trend_ma20_slope, trend_ma_compression_band, trend_ma_crossover_event,
	// momentum_rsi_zone, extremes_condition, extremes_condition_rarity, volatility_regime,
	// volume_ratio_band, fundamentals_valuation_zone, range_position.
	// Request ma8 through ma200 for raw MA values.
	// Request trend_ma8_slope through trend_ma200_slope for the full MA slope set.
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

// Field describes a single queryable snapshot field as returned by
// GET /v1/schema/fields.
type Field struct {
	// Name is the canonical flat snake_case column name (e.g. "momentum_rsi_zone").
	Name string `json:"name"`
	// Type is the SQL data type: "text", "integer", "numeric", "boolean", or "bigint".
	Type string `json:"type"`
	// Category groups related fields (e.g. "trend", "momentum", "fundamentals").
	Category string `json:"category"`
	// Values lists the allowed enum values for text fields. Nil for numeric/boolean fields.
	Values []string `json:"values,omitempty"`
	// Description is a human-readable explanation of the field.
	Description string `json:"description"`
}

// SchemaFields is the typed response shape embedded in GET /v1/schema/fields.
type SchemaFields struct {
	TotalFields int     `json:"total_fields"`
	Categories  []string `json:"categories"`
	Operators   []string `json:"operators"`
	Fields      []Field  `json:"fields"`
}

// SchemaResponse is the response from the Schema endpoint.
type SchemaResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// Fields unmarshals the raw Data payload into a typed SchemaFields struct.
func (r *SchemaResponse) Fields() (*SchemaFields, error) {
	var s SchemaFields
	if err := json.Unmarshal(r.Data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SummaryResponse is the response from the Summary endpoint.
type SummaryResponse struct {
	Data       json.RawMessage `json:"data"`
	RateLimits RateLimits      `json:"-"`
}

// Ptr returns a pointer to the given value. This is a convenience helper
// for constructing option structs with optional fields.
func Ptr[T any](v T) *T {
	return &v
}
