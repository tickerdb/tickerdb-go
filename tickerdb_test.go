package tickerdb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	tickerdb "github.com/tickerdb/tickerdb-go"
)

// testServer starts an httptest.Server backed by h and returns a Client
// pointing at it. The server is automatically closed when t finishes.
func testServer(t *testing.T, h http.HandlerFunc) *tickerdb.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return tickerdb.NewClient("tdb_test", tickerdb.WithBaseURL(srv.URL))
}

// jsonHandler returns an HTTP handler that always responds with statusCode and
// body serialised as JSON, plus any extra response headers.
func jsonHandler(statusCode int, body any, extra map[string]string) http.HandlerFunc {
	b, _ := json.Marshal(body)
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range extra {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write(b)
	}
}

// encodeJSON writes v as JSON to w, setting Content-Type automatically.
func encodeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// ─── Error handling ──────────────────────────────────────────────────────────

func TestAPIError_rateLimit(t *testing.T) {
	body := map[string]any{
		"error": map[string]any{
			"type":        "rate_limit_exceeded",
			"message":     "Monthly request limit reached.",
			"upgrade_url": "https://tickerdb.com/docs/pricing",
			"reset":       "2025-02-01T00:00:00Z",
		},
	}
	client := testServer(t, jsonHandler(429, body, nil))
	_, err := client.Account(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *tickerdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
	if apiErr.Type != "rate_limit_exceeded" {
		t.Errorf("Type = %q, want rate_limit_exceeded", apiErr.Type)
	}
	if !apiErr.IsRateLimitError() {
		t.Error("IsRateLimitError() = false, want true")
	}
	if apiErr.UpgradeURL != "https://tickerdb.com/docs/pricing" {
		t.Errorf("UpgradeURL = %q", apiErr.UpgradeURL)
	}
	// reset must decode as a string, not an int64 (regression for the bug fix)
	if apiErr.Reset == nil {
		t.Fatal("Reset is nil, want RFC3339 string")
	}
	if *apiErr.Reset != "2025-02-01T00:00:00Z" {
		t.Errorf("Reset = %q, want 2025-02-01T00:00:00Z", *apiErr.Reset)
	}
	rt, ok := apiErr.ResetTime()
	if !ok {
		t.Error("ResetTime() ok = false, want true")
	}
	if rt.IsZero() {
		t.Error("ResetTime() returned zero time")
	}
}

func TestAPIError_insufficientCredits(t *testing.T) {
	body := map[string]any{
		"error": map[string]any{
			"type":              "insufficient_credits",
			"message":           "This request costs 2 credits but you have 1 remaining.",
			"credits_required":  2,
			"credits_remaining": 1,
		},
	}
	client := testServer(t, jsonHandler(429, body, nil))
	_, err := client.OHLCV(context.Background(), "AAPL", nil)

	var apiErr *tickerdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.CreditsRequired == nil || *apiErr.CreditsRequired != 2 {
		t.Errorf("CreditsRequired = %v, want 2", apiErr.CreditsRequired)
	}
	if apiErr.CreditsRemaining == nil || *apiErr.CreditsRemaining != 1 {
		t.Errorf("CreditsRemaining = %v, want 1", apiErr.CreditsRemaining)
	}
}

func TestAPIError_forbidden(t *testing.T) {
	body := map[string]any{
		"error": map[string]any{
			"type":        "tier_restricted",
			"message":     "Upgrade required.",
			"upgrade_url": "https://tickerdb.com/docs/pricing",
		},
	}
	client := testServer(t, jsonHandler(403, body, nil))
	_, err := client.Account(context.Background())

	var apiErr *tickerdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsForbiddenError() {
		t.Error("IsForbiddenError() = false, want true")
	}
}

func TestAPIError_paymentRequired(t *testing.T) {
	body := map[string]any{
		"error": map[string]any{
			"type":    "card_declined",
			"message": "Your card was declined.",
		},
	}
	client := testServer(t, jsonHandler(402, body, nil))
	_, err := client.SetTeamSeats(context.Background(), "team-1", 5)

	var apiErr *tickerdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsPaymentRequiredError() {
		t.Error("IsPaymentRequiredError() = false, want true")
	}
}

func TestAPIError_notFound(t *testing.T) {
	body := map[string]any{
		"error": map[string]any{"type": "ticker_not_found", "message": "Not found."},
	}
	client := testServer(t, jsonHandler(404, body, nil))
	_, err := client.OHLCV(context.Background(), "FAKE", nil)

	var apiErr *tickerdb.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsNotFoundError() {
		t.Error("IsNotFoundError() = false, want true")
	}
}

// ─── Rate limit headers ───────────────────────────────────────────────────────

func TestRateLimits_headers(t *testing.T) {
	accountBody := minimalAccountBody("plus")
	headers := map[string]string{
		"X-Request-Limit":      "5000",
		"X-Requests-Used":      "42",
		"X-Requests-Remaining": "4958",
		"X-Request-Reset":      "2025-02-01T00:00:00Z",
	}
	client := testServer(t, jsonHandler(200, accountBody, headers))
	resp, err := client.Account(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rl := resp.RateLimits
	if rl.RequestLimit != 5000 {
		t.Errorf("RequestLimit = %d, want 5000", rl.RequestLimit)
	}
	if rl.RequestsUsed != 42 {
		t.Errorf("RequestsUsed = %d, want 42", rl.RequestsUsed)
	}
	if rl.RequestsRemaining != 4958 {
		t.Errorf("RequestsRemaining = %d, want 4958", rl.RequestsRemaining)
	}
	if rl.RequestReset.IsZero() {
		t.Error("RequestReset is zero, want parsed timestamp")
	}
}

func TestRateLimits_creditBalance(t *testing.T) {
	headers := map[string]string{
		"X-Request-Limit":      "5000",
		"X-Requests-Used":      "1",
		"X-Requests-Remaining": "4999",
		"X-Request-Reset":      "2025-02-01T00:00:00Z",
		"X-Credit-Balance":     "8.5",
	}
	client := testServer(t, jsonHandler(200, ohlcvFixture("AAPL", false, nil), headers))
	resp, err := client.OHLCV(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RateLimits.CreditBalance != 8.5 {
		t.Errorf("CreditBalance = %f, want 8.5", resp.RateLimits.CreditBalance)
	}
}

// ─── Auth header ─────────────────────────────────────────────────────────────

func TestClient_sendsAuthHeader(t *testing.T) {
	var gotAuth string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		encodeJSON(w, minimalAccountBody("free"))
	})
	if _, err := client.Account(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer tdb_test" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer tdb_test")
	}
}

// ─── Account ─────────────────────────────────────────────────────────────────

func TestAccount(t *testing.T) {
	body := map[string]any{
		"tier":      "plus",
		"tier_full": "plus",
		"email":     "user@example.com",
		"limits": map[string]any{
			"monthly_requests": 5000,
			"overage_enabled":  false,
			"watchlist_limit":  100,
			"search_results":   50,
			"webhook_urls":     2,
			"history_days":     365,
		},
		"usage": map[string]any{
			"monthly_requests_used":      42,
			"monthly_requests_remaining": 4958,
			"credit_balance":             10.5,
		},
		"scheduled_tier":      nil,
		"scheduled_change_at": nil,
	}
	client := testServer(t, jsonHandler(200, body, nil))
	resp, err := client.Account(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Tier != "plus" {
		t.Errorf("Tier = %q, want plus", resp.Tier)
	}
	if resp.Email != "user@example.com" {
		t.Errorf("Email = %q", resp.Email)
	}
	if resp.Limits.MonthlyRequests != 5000 {
		t.Errorf("Limits.MonthlyRequests = %d, want 5000", resp.Limits.MonthlyRequests)
	}
	if resp.Limits.HistoryDays != 365 {
		t.Errorf("Limits.HistoryDays = %d, want 365", resp.Limits.HistoryDays)
	}
	if resp.Limits.WebhookURLs != 2 {
		t.Errorf("Limits.WebhookURLs = %d, want 2", resp.Limits.WebhookURLs)
	}
	if resp.Usage.MonthlyRequestsUsed != 42 {
		t.Errorf("Usage.MonthlyRequestsUsed = %d, want 42", resp.Usage.MonthlyRequestsUsed)
	}
	if resp.Usage.CreditBalance != 10.5 {
		t.Errorf("Usage.CreditBalance = %f, want 10.5", resp.Usage.CreditBalance)
	}
	if resp.ScheduledTier != nil {
		t.Errorf("ScheduledTier = %v, want nil", resp.ScheduledTier)
	}
}

// ─── OHLCV ───────────────────────────────────────────────────────────────────

func TestOHLCV_response(t *testing.T) {
	client := testServer(t, jsonHandler(200, ohlcvFixture("AAPL", false, nil), nil))
	resp, err := client.OHLCV(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Ticker != "AAPL" {
		t.Errorf("Ticker = %q, want AAPL", resp.Ticker)
	}
	if resp.Adjustment != "split_and_dividend_adjusted" {
		t.Errorf("Adjustment = %q", resp.Adjustment)
	}
	if len(resp.Bars) != 1 {
		t.Fatalf("len(Bars) = %d, want 1", len(resp.Bars))
	}
	if resp.Bars[0].Close != 203.0 {
		t.Errorf("Bars[0].Close = %f, want 203.0", resp.Bars[0].Close)
	}
	if resp.HasMore {
		t.Error("HasMore = true, want false")
	}
}

func TestOHLCV_path(t *testing.T) {
	var gotPath string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		encodeJSON(w, ohlcvFixture("AAPL", false, nil))
	})
	if _, err := client.OHLCV(context.Background(), "AAPL", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/ohlcv/AAPL" {
		t.Errorf("path = %q, want /ohlcv/AAPL", gotPath)
	}
}

func TestOHLCV_params(t *testing.T) {
	var gotQuery url.Values
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		encodeJSON(w, ohlcvFixture("TSLA", false, nil))
	})
	opts := &tickerdb.OHLCVOptions{
		Start: tickerdb.Ptr("2025-01-01"),
		End:   tickerdb.Ptr("2025-03-31"),
		Order: tickerdb.Ptr("asc"),
		Limit: tickerdb.Ptr(50),
	}
	if _, err := client.OHLCV(context.Background(), "TSLA", opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checks := map[string]string{
		"start": "2025-01-01",
		"end":   "2025-03-31",
		"order": "asc",
		"limit": "50",
	}
	for param, want := range checks {
		if got := gotQuery.Get(param); got != want {
			t.Errorf("param %q = %q, want %q", param, got, want)
		}
	}
}

func TestOHLCVAll_singlePage(t *testing.T) {
	client := testServer(t, jsonHandler(200, ohlcvFixture("AAPL", false, nil), nil))
	bars, err := client.OHLCVAll(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bars) != 1 {
		t.Errorf("len(bars) = %d, want 1", len(bars))
	}
}

func TestOHLCVAll_multiPage(t *testing.T) {
	cursor := "2025-01-14"
	page := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if page == 0 {
			body = ohlcvFixture("AAPL", true, &cursor)
		} else {
			body = ohlcvFixture("AAPL", false, nil)
		}
		page++
		encodeJSON(w, body)
	})
	bars, err := client.OHLCVAll(context.Background(), "AAPL", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 bar per page × 2 pages
	if len(bars) != 2 {
		t.Errorf("len(bars) = %d, want 2", len(bars))
	}
}

func TestOHLCVAll_cursorForwarded(t *testing.T) {
	cursor := "2025-01-14"
	page := 0
	var gotCursor string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if page == 1 {
			gotCursor = r.URL.Query().Get("cursor")
		}
		var body map[string]any
		if page == 0 {
			body = ohlcvFixture("AAPL", true, &cursor)
		} else {
			body = ohlcvFixture("AAPL", false, nil)
		}
		page++
		encodeJSON(w, body)
	})
	if _, err := client.OHLCVAll(context.Background(), "AAPL", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCursor != cursor {
		t.Errorf("cursor on page 2 = %q, want %q", gotCursor, cursor)
	}
}

// ─── Search date param ────────────────────────────────────────────────────────

func TestSearch_dateParam(t *testing.T) {
	var gotQuery url.Values
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		encodeJSON(w, map[string]any{"timeframe": "daily", "results": []any{}, "result_count": 0})
	})
	opts := &tickerdb.SearchOptions{
		Filters: `[{"field":"momentum_rsi_zone","op":"eq","value":"oversold"}]`,
		Date:    tickerdb.Ptr("2025-01-15"),
	}
	if _, err := client.Search(context.Background(), opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("date") != "2025-01-15" {
		t.Errorf("date param = %q, want 2025-01-15", gotQuery.Get("date"))
	}
}

func TestSearchBuilder_onDate(t *testing.T) {
	var gotQuery url.Values
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		encodeJSON(w, map[string]any{"timeframe": "daily", "results": []any{}, "result_count": 0})
	})
	_, err := client.Query().
		Eq("momentum_rsi_zone", "oversold").
		OnDate("2025-06-01").
		Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("date") != "2025-06-01" {
		t.Errorf("date param = %q, want 2025-06-01", gotQuery.Get("date"))
	}
}

// ─── Webhook deliveries ───────────────────────────────────────────────────────

func TestWebhookDeliveries(t *testing.T) {
	body := map[string]any{
		"deliveries": []map[string]any{
			{
				"id":         "del-1",
				"webhook_id": "wh-1",
				"event_type": "watchlist.changes",
				"timeframe":  "daily",
				"run_date":   "2025-01-15",
				"status":     "success",
			},
		},
		"count": 1,
		"limit": 50,
	}
	client := testServer(t, jsonHandler(200, body, nil))
	resp, err := client.WebhookDeliveries(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
	if len(resp.Deliveries) != 1 {
		t.Fatalf("len(Deliveries) = %d, want 1", len(resp.Deliveries))
	}
	if resp.Deliveries[0].EventType != tickerdb.WebhookEventWatchlistChanges {
		t.Errorf("EventType = %q, want %q", resp.Deliveries[0].EventType, tickerdb.WebhookEventWatchlistChanges)
	}
	if resp.Deliveries[0].Status != "success" {
		t.Errorf("Status = %q, want success", resp.Deliveries[0].Status)
	}
}

func TestWebhookDeliveries_filterByWebhookID(t *testing.T) {
	var gotQuery url.Values
	body := map[string]any{"deliveries": []any{}, "count": 0, "limit": 50}
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		encodeJSON(w, body)
	})
	opts := &tickerdb.WebhookDeliveriesOptions{WebhookID: tickerdb.Ptr("wh-abc")}
	if _, err := client.WebhookDeliveries(context.Background(), opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("webhook_id") != "wh-abc" {
		t.Errorf("webhook_id param = %q, want wh-abc", gotQuery.Get("webhook_id"))
	}
}

// ─── Webhook event constants ──────────────────────────────────────────────────

func TestWebhookEventConstants(t *testing.T) {
	if tickerdb.WebhookEventWatchlistChanges != "watchlist.changes" {
		t.Errorf("WebhookEventWatchlistChanges = %q, want watchlist.changes", tickerdb.WebhookEventWatchlistChanges)
	}
	if tickerdb.WebhookEventDataReady != "data.ready" {
		t.Errorf("WebhookEventDataReady = %q, want data.ready", tickerdb.WebhookEventDataReady)
	}
}

// ─── Fixtures ────────────────────────────────────────────────────────────────

func minimalAccountBody(tier string) map[string]any {
	return map[string]any{
		"tier": tier, "tier_full": tier, "email": "x@x.com",
		"limits": map[string]any{}, "usage": map[string]any{},
	}
}

func ohlcvFixture(ticker string, hasMore bool, nextCursor *string) map[string]any {
	return map[string]any{
		"ticker":            ticker,
		"asset_class":       "stock",
		"currency":          "USD",
		"timeframe":         "daily",
		"data_status":       "eod",
		"adjustment":        "split_and_dividend_adjusted",
		"order":             "desc",
		"start":             "2025-01-01",
		"end":               nil,
		"row_count":         1,
		"has_more":          hasMore,
		"next_cursor":       nextCursor,
		"bars":              []map[string]any{{"date": "2025-01-15", "open": 200.0, "high": 205.0, "low": 198.0, "close": 203.0, "volume": 50_000_000.0}},
		"plan_history_days": 365,
		"plan":              "Plus",
	}
}
