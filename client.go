// Package tickerdb provides a Go client for the TickerDB financial data API.
//
// For full API documentation, see https://tickerdb.ai/docs
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
	defaultBaseURL = "https://api.tickerdb.ai/v1"
	userAgent      = "tickerdb-sdk-go"
)

// Client is the TickerDB client. Create one with NewClient.
type Client struct {
	// APIKey is the bearer token used for authentication.
	APIKey string

	// BaseURL is the base URL of the API. Defaults to https://api.tickerdb.ai/v1.
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
func (c *Client) Summary(ctx context.Context, ticker string, opts *SummaryOptions) (*SummaryResponse, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Timeframe != nil {
			params.Set("timeframe", string(*opts.Timeframe))
		}
		if opts.Date != nil {
			params.Set("date", *opts.Date)
		}
	}

	resp := &SummaryResponse{}
	rateLimits, err := c.doGet(ctx, fmt.Sprintf("/summary/%s", url.PathEscape(ticker)), params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// History retrieves a historical series for a single ticker across a date range.
func (c *Client) History(ctx context.Context, ticker string, opts *HistoryOptions) (*HistoryResponse, error) {
	if opts == nil || opts.Start == "" || opts.End == "" {
		return nil, fmt.Errorf("tickerdb: history requires start and end dates")
	}

	params := url.Values{}
	if opts.Timeframe != nil {
		params.Set("timeframe", string(*opts.Timeframe))
	}
	params.Set("start", opts.Start)
	params.Set("end", opts.End)

	resp := &HistoryResponse{}
	rateLimits, err := c.doGet(ctx, fmt.Sprintf("/history/%s", url.PathEscape(ticker)), params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// Compare retrieves a comparison of multiple tickers.
func (c *Client) Compare(ctx context.Context, tickers []string, opts *CompareOptions) (*CompareResponse, error) {
	params := url.Values{}
	params.Set("tickers", strings.Join(tickers, ","))
	if opts != nil {
		if opts.Timeframe != nil {
			params.Set("timeframe", string(*opts.Timeframe))
		}
		if opts.Date != nil {
			params.Set("date", *opts.Date)
		}
	}

	resp := &CompareResponse{}
	rateLimits, err := c.doGet(ctx, "/compare", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// Watchlist retrieves analysis for a list of tickers.
func (c *Client) Watchlist(ctx context.Context, tickers []string, opts *WatchlistOptions) (*WatchlistResponse, error) {
	body := WatchlistRequest{
		Tickers: tickers,
	}
	if opts != nil {
		body.Timeframe = opts.Timeframe
	}

	resp := &WatchlistResponse{}
	rateLimits, err := c.doPost(ctx, "/watchlist", body, resp)
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

// Assets retrieves all available assets.
func (c *Client) Assets(ctx context.Context) (*AssetsResponse, error) {
	resp := &AssetsResponse{}
	rateLimits, err := c.doGet(ctx, "/assets", nil, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ListSectors retrieves all valid sector values with asset counts.
func (c *Client) ListSectors(ctx context.Context) (*SectorsResponse, error) {
	resp := &SectorsResponse{}
	rateLimits, err := c.doGet(ctx, "/list/sectors", nil, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ListEvents searches for historical band transition events for a ticker.
// ticker and field are required; opts may be nil for defaults.
func (c *Client) ListEvents(ctx context.Context, ticker, field string, opts *ListEventsOptions) (*ListEventsResponse, error) {
	params := url.Values{}
	params.Set("ticker", ticker)
	params.Set("field", field)
	if opts != nil {
		if opts.Timeframe != nil {
			params.Set("timeframe", string(*opts.Timeframe))
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

	resp := &ListEventsResponse{}
	rateLimits, err := c.doGet(ctx, "/events", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ScanOversold retrieves oversold assets matching the given criteria.
func (c *Client) ScanOversold(ctx context.Context, opts *ScanOversoldOptions) (*ScanOversoldResponse, error) {
	params := url.Values{}
	if opts != nil {
		addScanParams(params, opts.Timeframe, opts.Sector, opts.Limit, opts.Date)
		if opts.AssetClass != nil {
			params.Set("asset_class", string(*opts.AssetClass))
		}
		if opts.MinSeverity != nil {
			params.Set("min_severity", string(*opts.MinSeverity))
		}
		if opts.SortBy != nil {
			params.Set("sort_by", *opts.SortBy)
		}
	}

	resp := &ScanOversoldResponse{}
	rateLimits, err := c.doGet(ctx, "/scan/oversold", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ScanBreakouts retrieves assets with breakout patterns.
func (c *Client) ScanBreakouts(ctx context.Context, opts *ScanBreakoutsOptions) (*ScanBreakoutsResponse, error) {
	params := url.Values{}
	if opts != nil {
		addScanParams(params, opts.Timeframe, opts.Sector, opts.Limit, opts.Date)
		if opts.AssetClass != nil {
			params.Set("asset_class", string(*opts.AssetClass))
		}
		if opts.Direction != nil {
			params.Set("direction", string(*opts.Direction))
		}
		if opts.SortBy != nil {
			params.Set("sort_by", *opts.SortBy)
		}
	}

	resp := &ScanBreakoutsResponse{}
	rateLimits, err := c.doGet(ctx, "/scan/breakouts", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ScanUnusualVolume retrieves assets with unusual volume.
func (c *Client) ScanUnusualVolume(ctx context.Context, opts *ScanUnusualVolumeOptions) (*ScanUnusualVolumeResponse, error) {
	params := url.Values{}
	if opts != nil {
		addScanParams(params, opts.Timeframe, opts.Sector, opts.Limit, opts.Date)
		if opts.AssetClass != nil {
			params.Set("asset_class", string(*opts.AssetClass))
		}
		if opts.MinRatioBand != nil {
			params.Set("min_ratio_band", string(*opts.MinRatioBand))
		}
		if opts.SortBy != nil {
			params.Set("sort_by", *opts.SortBy)
		}
	}

	resp := &ScanUnusualVolumeResponse{}
	rateLimits, err := c.doGet(ctx, "/scan/unusual-volume", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ScanValuation retrieves assets based on valuation metrics.
func (c *Client) ScanValuation(ctx context.Context, opts *ScanValuationOptions) (*ScanValuationResponse, error) {
	params := url.Values{}
	if opts != nil {
		addScanParams(params, opts.Timeframe, opts.Sector, opts.Limit, opts.Date)
		if opts.Direction != nil {
			params.Set("direction", string(*opts.Direction))
		}
		if opts.MinSeverity != nil {
			params.Set("min_severity", string(*opts.MinSeverity))
		}
		if opts.SortBy != nil {
			params.Set("sort_by", *opts.SortBy)
		}
	}

	resp := &ScanValuationResponse{}
	rateLimits, err := c.doGet(ctx, "/scan/valuation", params, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// ScanInsiderActivity retrieves assets with notable insider trading activity.
func (c *Client) ScanInsiderActivity(ctx context.Context, opts *ScanInsiderActivityOptions) (*ScanInsiderActivityResponse, error) {
	params := url.Values{}
	if opts != nil {
		addScanParams(params, opts.Timeframe, opts.Sector, opts.Limit, opts.Date)
		if opts.Direction != nil {
			params.Set("direction", string(*opts.Direction))
		}
		if opts.SortBy != nil {
			params.Set("sort_by", *opts.SortBy)
		}
	}

	resp := &ScanInsiderActivityResponse{}
	rateLimits, err := c.doGet(ctx, "/scan/insider-activity", params, resp)
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

// addScanParams adds common scan query parameters.
func addScanParams(params url.Values, timeframe *Timeframe, sector *string, limit *int, date *string) {
	if timeframe != nil {
		params.Set("timeframe", string(*timeframe))
	}
	if sector != nil {
		params.Set("sector", *sector)
	}
	if limit != nil {
		params.Set("limit", strconv.Itoa(*limit))
	}
	if date != nil {
		params.Set("date", *date)
	}
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
