// Package tickerdb provides a Go client for the TickerDB financial data API.
//
// For full API documentation, see https://tickerdb.com/docs
package tickerdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultBaseURL = "https://api.tickerdb.com/v1"
	userAgent      = "tickerdb-go"
)

// Client is the TickerDB client. Create one with NewClient.
type Client struct {
	// APIKey is the bearer token used for authentication.
	APIKey string

	// BaseURL is the base URL of the API. Defaults to https://api.tickerdb.com/v1.
	BaseURL string

	// HTTPClient is the HTTP client used for requests. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// NewClient creates a new TickerDB client with the given API key and options.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Summary retrieves a technical analysis summary for a single ticker.
//
// Supports 4 modes depending on which options are provided:
//   - Snapshot (default): Current categorical state.
//   - Historical snapshot: Set Date for a point-in-time snapshot.
//   - Historical series: Set Start/End for a date range of snapshots.
//   - Events: Set Field (and optionally Band) for band transition history.
func (c *Client) Summary(ctx context.Context, ticker string, opts *SummaryOptions) (*SummaryResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Timeframe != nil {
			params.Set("timeframe", string(*opts.Timeframe))
		}
		if opts.Date != nil {
			params.Set("date", *opts.Date)
		}
		if opts.Start != nil {
			params.Set("start", *opts.Start)
		}
		if opts.End != nil {
			params.Set("end", *opts.End)
		}
		if opts.Field != nil {
			params.Set("field", *opts.Field)
		}
		if opts.Band != nil {
			params.Set("band", *opts.Band)
		}
		if opts.Limit != nil {
			params.Set("limit", strconv.Itoa(*opts.Limit))
		}
		if opts.Before != nil {
			params.Set("before", *opts.Before)
		}
		if opts.After != nil {
			params.Set("after", *opts.After)
		}
		if opts.ContextTicker != nil {
			params.Set("context_ticker", *opts.ContextTicker)
		}
		if opts.ContextField != nil {
			params.Set("context_field", *opts.ContextField)
		}
		if opts.ContextBand != nil {
			params.Set("context_band", *opts.ContextBand)
		}
	}

	resp := &SummaryResponse{}
	rateLimits, err := c.doGet(ctx, "/summary/"+url.PathEscape(ticker), params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// Search finds assets matching filter criteria.
func (c *Client) Search(ctx context.Context, opts *SearchOptions) (*SearchResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Filters != "" {
			params.Set("filters", opts.Filters)
		}
		if opts.Timeframe != nil {
			params.Set("timeframe", string(*opts.Timeframe))
		}
		if opts.Limit != nil {
			params.Set("limit", strconv.Itoa(*opts.Limit))
		}
		if opts.Offset != nil {
			params.Set("offset", strconv.Itoa(*opts.Offset))
		}
		if opts.Fields != "" {
			params.Set("fields", opts.Fields)
		}
		if opts.SortBy != nil {
			params.Set("sort_by", *opts.SortBy)
		}
		if opts.SortDirection != nil {
			params.Set("sort_direction", *opts.SortDirection)
		}
	}

	resp := &SearchResponse{}
	rateLimits, err := c.doGet(ctx, "/search", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// Schema retrieves the schema of available fields and their valid band values.
func (c *Client) Schema(ctx context.Context) (*SchemaResponse, error) {
	resp := &SchemaResponse{}
	rateLimits, err := c.doGet(ctx, "/schema/fields", nil, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// Watchlist retrieves the saved watchlist snapshot for the authenticated account.
func (c *Client) Watchlist(ctx context.Context, opts *WatchlistOptions) (*WatchlistResponse, error) {
	params := url.Values{}
	if opts != nil && opts.Date != nil {
		params.Set("date", *opts.Date)
	}
	resp := &WatchlistResponse{}
	rateLimits, err := c.doGet(ctx, "/watchlist", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// AddToWatchlist adds tickers to the saved watchlist.
func (c *Client) AddToWatchlist(ctx context.Context, tickers []string) (*AddToWatchlistResponse, error) {
	body := WatchlistMutationRequest{
		Tickers: normalizeTickers(tickers),
	}

	resp := &AddToWatchlistResponse{}
	rateLimits, err := c.doPost(ctx, "/watchlist", body, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// RemoveFromWatchlist removes tickers from the saved watchlist.
func (c *Client) RemoveFromWatchlist(ctx context.Context, tickers []string) (*RemoveFromWatchlistResponse, error) {
	body := WatchlistMutationRequest{
		Tickers: normalizeTickers(tickers),
	}

	resp := &RemoveFromWatchlistResponse{}
	rateLimits, err := c.doDeleteJSON(ctx, "/watchlist", body, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// WatchlistChanges returns field-level state changes for the user's saved
// watchlist tickers since the last pipeline run. Available on all tiers.
func (c *Client) WatchlistChanges(ctx context.Context, opts *WatchlistChangesOptions) (*WatchlistChangesResponse, error) {
	params := url.Values{}
	if opts != nil && opts.Timeframe != nil {
		params.Set("timeframe", string(*opts.Timeframe))
	}

	resp := &WatchlistChangesResponse{}
	rateLimits, err := c.doGet(ctx, "/watchlist/changes", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ListWebhooks retrieves all webhooks for the authenticated account.
func (c *Client) ListWebhooks(ctx context.Context) (*WebhookListResponse, error) {
	resp := &WebhookListResponse{}
	_, err := c.doGet(ctx, "/webhooks", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateWebhook creates a new webhook.
func (c *Client) CreateWebhook(ctx context.Context, req CreateWebhookRequest) (*WebhookCreated, error) {
	resp := &WebhookCreated{}
	_, err := c.doPost(ctx, "/webhooks", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateWebhook updates an existing webhook.
func (c *Client) UpdateWebhook(ctx context.Context, req UpdateWebhookRequest) (*WebhookUpdateResponse, error) {
	resp := &WebhookUpdateResponse{}
	_, err := c.doPut(ctx, "/webhooks", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteWebhook deletes a webhook.
func (c *Client) DeleteWebhook(ctx context.Context, req DeleteWebhookRequest) (*WebhookDeleteResponse, error) {
	resp := &WebhookDeleteResponse{}
	_, err := c.doDeleteJSON(ctx, "/webhooks", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Fluent search query builder
// ──────────────────────────────────────────────────────────────────────────────

// SearchBuilder provides a fluent interface for constructing search queries.
//
// Usage:
//
//	results, err := client.Query().
//	    Eq("momentum_rsi_zone", "oversold").
//	    Eq("sector", "Technology").
//	    Select("ticker", "sector", "momentum_rsi_zone").
//	    Sort("extremes_condition_percentile", "asc").
//	    Limit(10).
//	    Execute(ctx)
type SearchBuilder struct {
	client         *Client
	filters        []searchFilter
	fields         []string
	sortBy         *string
	sortDirection  *string
	timeframe      *Timeframe
	limit          *int
	offset         *int
}

type searchFilter struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

// Query creates a new fluent SearchBuilder for constructing search queries.
func (c *Client) Query() *SearchBuilder {
	return &SearchBuilder{client: c}
}

// Eq adds an equality filter.
func (b *SearchBuilder) Eq(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "eq", Value: value})
	return b
}

// Neq adds a not-equal filter.
func (b *SearchBuilder) Neq(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "neq", Value: value})
	return b
}

// In adds an "in" filter for matching any of the given values.
func (b *SearchBuilder) In(field string, values ...interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "in", Value: values})
	return b
}

// Gt adds a greater-than filter.
func (b *SearchBuilder) Gt(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "gt", Value: value})
	return b
}

// Gte adds a greater-than-or-equal filter.
func (b *SearchBuilder) Gte(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "gte", Value: value})
	return b
}

// Lt adds a less-than filter.
func (b *SearchBuilder) Lt(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "lt", Value: value})
	return b
}

// Lte adds a less-than-or-equal filter.
func (b *SearchBuilder) Lte(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, searchFilter{Field: field, Op: "lte", Value: value})
	return b
}

// Select specifies which fields to return in the results.
func (b *SearchBuilder) Select(fields ...string) *SearchBuilder {
	b.fields = fields
	return b
}

// Sort sets the sort column and direction ("asc" or "desc").
func (b *SearchBuilder) Sort(field string, direction string) *SearchBuilder {
	b.sortBy = &field
	b.sortDirection = &direction
	return b
}

// Limit sets the maximum number of results.
func (b *SearchBuilder) Limit(n int) *SearchBuilder {
	b.limit = &n
	return b
}

// Offset sets the pagination offset.
func (b *SearchBuilder) Offset(n int) *SearchBuilder {
	b.offset = &n
	return b
}

// WithTimeframe sets the analysis timeframe.
func (b *SearchBuilder) WithTimeframe(tf Timeframe) *SearchBuilder {
	b.timeframe = &tf
	return b
}

// Execute runs the built query against the search endpoint.
func (b *SearchBuilder) Execute(ctx context.Context) (*SearchResponse, error) {
	opts := &SearchOptions{}

	if len(b.filters) > 0 {
		filtersJSON, err := json.Marshal(b.filters)
		if err != nil {
			return nil, fmt.Errorf("tickerdb: encoding filters: %w", err)
		}
		opts.Filters = string(filtersJSON)
	}

	if len(b.fields) > 0 {
		fieldsJSON, err := json.Marshal(b.fields)
		if err != nil {
			return nil, fmt.Errorf("tickerdb: encoding fields: %w", err)
		}
		opts.Fields = string(fieldsJSON)
	}

	opts.SortBy = b.sortBy
	opts.SortDirection = b.sortDirection
	opts.Timeframe = b.timeframe
	opts.Limit = b.limit
	opts.Offset = b.offset

	return b.client.Search(ctx, opts)
}

// doGet performs an authenticated GET request and decodes the response.
func (c *Client) doGet(ctx context.Context, path string, params url.Values, dst interface{}) (RateLimits, error) {
	u := c.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: creating request: %w", err)
	}

	return c.do(req, dst)
}

// doPost performs an authenticated POST request with a JSON body and decodes the response.
func (c *Client) doPost(ctx context.Context, path string, body interface{}, dst interface{}) (RateLimits, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: encoding request body: %w", err)
	}

	u := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonBody))
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, dst)
}

// doPut performs an authenticated PUT request with a JSON body and decodes the response.
func (c *Client) doPut(ctx context.Context, path string, body interface{}, dst interface{}) (RateLimits, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: encoding request body: %w", err)
	}

	u := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(jsonBody))
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, dst)
}

// doDeleteJSON performs an authenticated DELETE request with a JSON body and decodes the response.
func (c *Client) doDeleteJSON(ctx context.Context, path string, body interface{}, dst interface{}) (RateLimits, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: encoding request body: %w", err)
	}

	u := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, bytes.NewReader(jsonBody))
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, dst)
}

// do executes an HTTP request, checks for errors, parses rate limits, and decodes the response.
func (c *Client) do(req *http.Request, dst interface{}) (RateLimits, error) {
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return RateLimits{}, fmt.Errorf("tickerdb: sending request: %w", err)
	}
	defer resp.Body.Close()

	rateLimits := parseRateLimits(resp.Header)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return rateLimits, fmt.Errorf("tickerdb: reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var envelope apiErrorEnvelope
		if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
			return rateLimits, &APIError{
				StatusCode: resp.StatusCode,
				Type:       "unknown",
				Message:    fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes)),
			}
		}
		apiErr := &envelope.Error
		apiErr.StatusCode = resp.StatusCode
		return rateLimits, apiErr
	}

	if dst != nil {
		if err := json.Unmarshal(bodyBytes, dst); err != nil {
			return rateLimits, fmt.Errorf("tickerdb: decoding response: %w", err)
		}
	}

	return rateLimits, nil
}

func normalizeTickers(tickers []string) []string {
	normalized := make([]string, 0, len(tickers))
	for _, ticker := range tickers {
		trimmed := strings.TrimSpace(ticker)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, strings.ToUpper(trimmed))
	}
	return normalized
}
