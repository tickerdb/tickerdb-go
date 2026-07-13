package tickerdb

import (
	"fmt"
	"time"
)

// APIError represents an error response from the TickerDB API.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int `json:"-"`

	// Type is the error type identifier (e.g., "invalid_token", "rate_limit_exceeded").
	Type string `json:"type"`

	// Message is a human-readable description of the error.
	Message string `json:"message"`

	// UpgradeURL is present on 403 and 429 responses, pointing to a plan upgrade page.
	UpgradeURL string `json:"upgrade_url,omitempty"`

	// Reset is an RFC 3339 timestamp indicating when the rate limit resets (429
	// responses). Use ResetTime() to parse it into a time.Time.
	Reset *string `json:"reset,omitempty"`

	// CreditsRequired is the number of credits needed for the request (429
	// insufficient_credits responses from metered endpoints such as OHLCV).
	CreditsRequired *int `json:"credits_required,omitempty"`

	// CreditsRemaining is the number of credits available at the time of the
	// error (429 insufficient_credits responses).
	CreditsRemaining *int `json:"credits_remaining,omitempty"`
}

// Error formats the API error for logs and callers.
func (e *APIError) Error() string {
	if e.UpgradeURL != "" {
		return fmt.Sprintf("tickerdb: %d %s: %s (upgrade: %s)", e.StatusCode, e.Type, e.Message, e.UpgradeURL)
	}
	return fmt.Sprintf("tickerdb: %d %s: %s", e.StatusCode, e.Type, e.Message)
}

// ResetTime parses the Reset field into a time.Time. Returns the zero value
// and false if Reset is not set or cannot be parsed.
func (e *APIError) ResetTime() (time.Time, bool) {
	if e.Reset == nil {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, *e.Reset)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// IsRateLimitError reports whether the error is a 429 rate limit error.
func (e *APIError) IsRateLimitError() bool {
	return e.StatusCode == 429
}

// IsAuthError reports whether the error is a 401 authentication error.
func (e *APIError) IsAuthError() bool {
	return e.StatusCode == 401
}

// IsForbiddenError reports whether the error is a 403 tier-restricted error.
func (e *APIError) IsForbiddenError() bool {
	return e.StatusCode == 403
}

// IsNotFoundError reports whether the error is a 404 not found error.
func (e *APIError) IsNotFoundError() bool {
	return e.StatusCode == 404
}

// IsPaymentRequiredError reports whether the error is a 402 payment required
// error (e.g., card declined on team seat changes).
func (e *APIError) IsPaymentRequiredError() bool {
	return e.StatusCode == 402
}

// apiErrorEnvelope is used to unmarshal the error response JSON.
type apiErrorEnvelope struct {
	Error APIError `json:"error"`
}
