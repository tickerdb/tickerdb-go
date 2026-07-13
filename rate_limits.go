package tickerdb

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimits contains the rate limit information returned in API response headers.
type RateLimits struct {
	// RequestLimit is the maximum number of requests allowed per period.
	RequestLimit int

	// RequestsUsed is the number of requests consumed in the current period.
	RequestsUsed int

	// RequestsRemaining is the number of requests remaining in the current period.
	RequestsRemaining int

	// RequestReset is the time when the request limit resets.
	RequestReset time.Time

	// CreditBalance is the prepaid credit balance remaining on the account.
	// Only populated on metered endpoints that consume credits (e.g. OHLCV).
	// Zero means either no credits remaining or the endpoint does not report it.
	CreditBalance float64

	// Deprecated: The API no longer emits hourly rate-limit headers. This field
	// is always zero and will be removed in a future major version.
	HourlyRequestLimit int

	// Deprecated: The API no longer emits hourly rate-limit headers. This field
	// is always zero and will be removed in a future major version.
	HourlyRequestsUsed int

	// Deprecated: The API no longer emits hourly rate-limit headers. This field
	// is always zero and will be removed in a future major version.
	HourlyRequestsRemaining int

	// Deprecated: The API no longer emits hourly rate-limit headers. This field
	// is always zero and will be removed in a future major version.
	HourlyRequestReset time.Time
}

// parseRateLimits extracts rate limit information from response headers.
func parseRateLimits(h http.Header) RateLimits {
	return RateLimits{
		RequestLimit:      headerInt(h, "X-Request-Limit"),
		RequestsUsed:      headerInt(h, "X-Requests-Used"),
		RequestsRemaining: headerInt(h, "X-Requests-Remaining"),
		RequestReset:      headerTime(h, "X-Request-Reset"),
		CreditBalance:     headerFloat(h, "X-Credit-Balance"),
		// Hourly fields retained for parsing in case of older deployments;
		// the current API does not emit these headers.
		HourlyRequestLimit:      headerInt(h, "X-Hourly-Request-Limit"),
		HourlyRequestsUsed:      headerInt(h, "X-Hourly-Requests-Used"),
		HourlyRequestsRemaining: headerInt(h, "X-Hourly-Requests-Remaining"),
		HourlyRequestReset:      headerTime(h, "X-Hourly-Request-Reset"),
	}
}

func headerInt(h http.Header, key string) int {
	v := h.Get(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func headerFloat(h http.Header, key string) float64 {
	v := h.Get(key)
	if v == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func headerTime(h http.Header, key string) time.Time {
	v := h.Get(key)
	if v == "" {
		return time.Time{}
	}

	// Try parsing as Unix timestamp first.
	if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
		return time.Unix(ts, 0)
	}

	// Try parsing as ISO 8601 / RFC 3339.
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t
	}

	return time.Time{}
}
