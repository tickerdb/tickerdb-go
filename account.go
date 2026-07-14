package tickerdb

import "context"

// AccountLimits describes the plan-level limits for the authenticated account.
type AccountLimits struct {
	// MonthlyRequests is the total requests allowed per billing period.
	MonthlyRequests int `json:"monthly_requests"`

	// OverageEnabled reports whether pay-as-you-go overage is active.
	OverageEnabled bool `json:"overage_enabled"`

	// SearchResults is the maximum results returned per search request.
	SearchResults int `json:"search_results"`

	// HistoryDays is the number of calendar days of history accessible on
	// the current plan.
	HistoryDays int `json:"history_days"`
}

// AccountUsage describes current-period consumption for the authenticated account.
type AccountUsage struct {
	// MonthlyRequestsUsed is the number of requests consumed this period.
	MonthlyRequestsUsed int `json:"monthly_requests_used"`

	// MonthlyRequestsRemaining is the number of requests still available.
	MonthlyRequestsRemaining int `json:"monthly_requests_remaining"`

	// CreditBalance is the prepaid credit balance available for metered
	// endpoints such as OHLCV.
	CreditBalance float64 `json:"credit_balance"`
}

// AccountResponse is the response from the Account endpoint.
type AccountResponse struct {
	// Tier is the base plan name: "free", "plus", "pro", or "business".
	Tier string `json:"tier"`

	// TierFull is the full internal tier identifier (may include suffixes
	// such as team membership variants).
	TierFull string `json:"tier_full"`

	// Email is the email address of the authenticated account.
	Email string `json:"email"`

	// Limits contains the plan-level caps for this account.
	Limits AccountLimits `json:"limits"`

	// Usage contains current-period consumption counters.
	Usage AccountUsage `json:"usage"`

	// ScheduledTier is the tier this account will move to at the next billing
	// boundary (set when a downgrade is pending). Nil if no change is scheduled.
	ScheduledTier *string `json:"scheduled_tier"`

	// ScheduledChangeAt is the RFC 3339 timestamp of the pending tier change.
	// Nil if no change is scheduled.
	ScheduledChangeAt *string `json:"scheduled_change_at"`

	RateLimits RateLimits `json:"-"`
}

// Account retrieves plan details, limits, and current-period usage for the
// authenticated account. This call does not consume a request credit.
func (c *Client) Account(ctx context.Context) (*AccountResponse, error) {
	resp := &AccountResponse{}
	rateLimits, err := c.doGet(ctx, "/account", nil, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}
