package tickerdb

import "fmt"

// APIError represents an error response from the TickerDB.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int `json:"-"`

	// Type is the error type identifier (e.g., "invalid_token", "rate_limit_exceeded").
	Type string `json:"type"`

	// Message is a human-readable description of the error.
	Message string `json:"message"`

	// UpgradeURL is present on 403 and 429 responses, pointing to a plan upgrade page.
	UpgradeURL string `json:"upgrade_url,omitempty"`

	// Reset is a Unix timestamp indicating when the rate limit resets (429 responses).
	Reset *int64 `json:"reset,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.UpgradeURL != "" {
		return fmt.Sprintf("tickerdb: %d %s: %s (upgrade: %s)", e.StatusCode, e.Type, e.Message, e.UpgradeURL)
	}
	return fmt.Sprintf("tickerdb: %d %s: %s", e.StatusCode, e.Type, e.Message)
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

// apiErrorEnvelope is used to unmarshal the error response JSON.
type apiErrorEnvelope struct {
	Error APIError `json:"error"`
}
