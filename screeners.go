package tickerdb

import "context"

// ScreenerFilter is a single filter condition within a screener.
//
// Value filters (Type "value" or omitted):
//
//	ScreenerFilter{Field: "momentum_rsi_zone", Op: "in", Value: []string{"oversold", "deep_oversold"}}
//
// Change filters (Type "change") fire when a field transitions between bands:
//
//	ScreenerFilter{Type: "change", Field: "trend_direction", Op: "changed", From: "downtrend", To: "uptrend"}
type ScreenerFilter struct {
	// Type is "value" (default) or "change".
	Type string `json:"type,omitempty"`

	// Field is the canonical flat snake_case field name (e.g. "momentum_rsi_zone").
	Field string `json:"field"`

	// Op is the filter operator: eq, neq, in, gt, gte, lt, lte, exists, changed.
	Op string `json:"op"`

	// Value is the comparison value for value filters.
	Value interface{} `json:"value,omitempty"`

	// From is the source band for change filters.
	From interface{} `json:"from,omitempty"`

	// To is the target band for change filters.
	To interface{} `json:"to,omitempty"`

	// Periods is the look-back window for change filters (default 1).
	Periods *int `json:"periods,omitempty"`
}

// ScreenerSort configures result ordering for a screener.
type ScreenerSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// Screener represents a saved or built-in screener configuration.
type Screener struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"`      // "default" or "custom"
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Timeframe   string          `json:"timeframe"` // "daily" or "weekly"
	Filters     []ScreenerFilter `json:"filters"`

	// ReturnFields is the server-derived set of fields included in screener
	// results. It is computed automatically from filters and sort on create/update
	// and cannot be set directly.
	ReturnFields []string      `json:"return_fields"`
	Sort         *ScreenerSort `json:"sort"`

	// Readonly is true for built-in default screeners.
	Readonly bool `json:"readonly,omitempty"`
}

// ListScreenersResponse is the response from listing screeners.
type ListScreenersResponse struct {
	// Defaults contains built-in screeners (readonly).
	Defaults []Screener `json:"defaults"`

	// Saved contains the account's custom screeners.
	Saved []Screener `json:"saved"`

	// Screeners is the merged list of Defaults followed by Saved.
	Screeners []Screener `json:"screeners"`

	// Fields is the full field catalog, using the same shape as /v1/schema/fields.
	Fields []Field `json:"fields"`

	RateLimits RateLimits `json:"-"`
}

// CreateScreenerRequest is the request body for creating a screener.
type CreateScreenerRequest struct {
	// Name is the display name (up to 120 characters). If omitted the server
	// derives a name from the first filter.
	Name string `json:"name,omitempty"`

	// Timeframe is "daily" (default) or "weekly".
	Timeframe string `json:"timeframe,omitempty"`

	// Filters must contain at least one valid filter condition.
	Filters []ScreenerFilter `json:"filters"`

	// Sort controls result ordering. If nil the server uses a sensible default.
	Sort *ScreenerSort `json:"sort,omitempty"`

	// LimitCount caps the number of results the screener returns (1–50, default 30).
	LimitCount *int `json:"limit_count,omitempty"`
}

// UpdateScreenerRequest is the request body for updating an existing screener.
type UpdateScreenerRequest struct {
	// ID is the screener to update (required).
	ID string `json:"id"`

	Name      *string          `json:"name,omitempty"`
	Timeframe *string          `json:"timeframe,omitempty"`
	Filters   []ScreenerFilter `json:"filters,omitempty"`
	Sort      *ScreenerSort    `json:"sort,omitempty"`
	LimitCount *int            `json:"limit_count,omitempty"`
}

// DeleteScreenerResponse is the response from deleting or hiding a screener.
type DeleteScreenerResponse struct {
	OK   bool   `json:"ok"`
	ID   string `json:"id"`
	Kind string `json:"kind"`

	// Deleted is true when a custom screener was permanently removed.
	Deleted bool `json:"deleted,omitempty"`

	// Hidden is true when a default screener was hidden (not permanently deleted).
	Hidden bool `json:"hidden,omitempty"`
}

// screenerMutationResponse is the shared envelope for create/update responses.
type screenerMutationResponse struct {
	Screener Screener `json:"screener"`
}

// ListScreeners returns all screeners for the account: built-in defaults
// (readonly) followed by custom saved screeners, plus the full field catalog.
func (c *Client) ListScreeners(ctx context.Context) (*ListScreenersResponse, error) {
	resp := &ListScreenersResponse{}
	rateLimits, err := c.doGet(ctx, "/screeners", nil, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// CreateScreener saves a new custom screener. The server derives ReturnFields
// automatically from the filters and sort; the returned Screener reflects the
// final saved state.
func (c *Client) CreateScreener(ctx context.Context, req CreateScreenerRequest) (*Screener, error) {
	var envelope screenerMutationResponse
	_, err := c.doPost(ctx, "/screeners", req, &envelope)
	if err != nil {
		return nil, err
	}
	return &envelope.Screener, nil
}

// UpdateScreener updates an existing custom screener by ID. Only fields that
// are set (non-nil / non-zero) are applied; the rest retain their current values.
func (c *Client) UpdateScreener(ctx context.Context, req UpdateScreenerRequest) (*Screener, error) {
	var envelope screenerMutationResponse
	_, err := c.doPut(ctx, "/screeners", req, &envelope)
	if err != nil {
		return nil, err
	}
	return &envelope.Screener, nil
}

// DeleteScreener removes or hides a screener by ID.
//
// kind must be "custom" or "default". Deleting a "default" screener hides it
// from the account's screener list (Hidden=true in the response); it does not
// permanently remove it. Deleting a "custom" screener permanently removes it
// (Deleted=true in the response).
func (c *Client) DeleteScreener(ctx context.Context, id, kind string) (*DeleteScreenerResponse, error) {
	body := map[string]string{"id": id, "kind": kind}
	resp := &DeleteScreenerResponse{}
	_, err := c.doDeleteJSON(ctx, "/screeners", body, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
