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

	// HourlyRequestLimit is the maximum number of requests allowed per hour.
	HourlyRequestLimit int

	// HourlyRequestsUsed is the number of requests consumed in the current hour.
	HourlyRequestsUsed int

	// HourlyRequestsRemaining is the number of requests remaining in the current hour.
	HourlyRequestsRemaining int

	// HourlyRequestReset is the time when the hourly limit resets.
	HourlyRequestReset time.Time
}

// parseRateLimits extracts rate limit information from HTTP response headers.
func parseRateLimits(h http.Header) RateLimits {
	return RateLimits{
		RequestLimit:            headerInt(h, "X-Request-Limit"),
		RequestsUsed:            headerInt(h, "X-Requests-Used"),
		RequestsRemaining:       headerInt(h, "X-Requests-Remaining"),
		RequestReset:            headerTime(h, "X-Request-Reset"),
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
