package tickerdb

import "context"

// TeamMember is a member of a team.
type TeamMember struct {
	UserID   string  `json:"user_id"`
	Email    string  `json:"email"`
	Name     *string `json:"name"`
	Role     string  `json:"role"` // "owner", "admin", or "member"
	JoinedAt string  `json:"joined_at"`
}

// TeamPendingInvite is an outstanding invite on a team the caller manages.
type TeamPendingInvite struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// Team is a team the authenticated user belongs to.
type Team struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	MaxSeats       int                 `json:"max_seats"`
	ExtraSeats     int                 `json:"extra_seats"`
	SeatsUsed      int                 `json:"seats_used"`
	SeatsAvailable int                 `json:"seats_available"`
	YourRole       string              `json:"your_role"`
	Members        []TeamMember        `json:"members"`
	PendingInvites []TeamPendingInvite `json:"pending_invites"`
}

// MyPendingInvite is an invite directed at the authenticated user's email.
type MyPendingInvite struct {
	ID           string `json:"id"`
	TeamID       string `json:"team_id"`
	TeamName     string `json:"team_name"`
	Role         string `json:"role"`
	InviterEmail string `json:"inviter_email"`
	ExpiresAt    string `json:"expires_at"`
}

// ListTeamsResponse is the response from ListTeams.
type ListTeamsResponse struct {
	// Teams contains every team the authenticated user belongs to.
	Teams []Team `json:"teams"`

	// MyPendingInvites lists outstanding invites addressed to the
	// authenticated user's email that have not yet been accepted.
	MyPendingInvites []MyPendingInvite `json:"my_pending_invites"`

	RateLimits RateLimits `json:"-"`
}

// TeamActionResponse is the response from team mutation actions (POST /v1/team).
// Which fields are populated depends on the action performed:
//
//   - CreateTeam:        Team
//   - InviteTeamMember:  Invite
//   - RemoveTeamMember:  Removed
//   - CancelTeamInvite:  Cancelled
//   - ResendTeamInvite:  Resent, ExpiresAt
//   - PromoteTeamMember: UserID, PreviousRole, NewRole
//   - RenameTeam:        TeamID, Name
//   - SetTeamSeats:      TeamID, MaxSeats, ExtraSeats, SeatsUsed, SeatPriceMonthly
//   - LeaveTeam:         Message only
type TeamActionResponse struct {
	Message string `json:"message"`

	// Populated by CreateTeam.
	Team *struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		MaxSeats int    `json:"max_seats"`
		YourRole string `json:"your_role"`
	} `json:"team,omitempty"`

	// Populated by InviteTeamMember and ResendTeamInvite.
	Invite *struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		ExpiresAt string `json:"expires_at"`
		TeamID    string `json:"team_id"`
	} `json:"invite,omitempty"`

	// Populated by RemoveTeamMember: the removed user's ID.
	Removed *string `json:"removed,omitempty"`

	// Populated by CancelTeamInvite: the cancelled invite ID.
	Cancelled *string `json:"cancelled,omitempty"`

	// Populated by ResendTeamInvite: the invite ID that was resent.
	Resent *string `json:"resent,omitempty"`

	// Populated by ResendTeamInvite: new expiry timestamp.
	ExpiresAt *string `json:"expires_at,omitempty"`

	// Populated by PromoteTeamMember.
	UserID       *string `json:"user_id,omitempty"`
	PreviousRole *string `json:"previous_role,omitempty"`
	NewRole      *string `json:"new_role,omitempty"`

	// Populated by RenameTeam and SetTeamSeats.
	TeamID *string `json:"team_id,omitempty"`
	Name   *string `json:"name,omitempty"`

	// Populated by SetTeamSeats.
	MaxSeats         *int    `json:"max_seats,omitempty"`
	ExtraSeats       *int    `json:"extra_seats,omitempty"`
	SeatsUsed        *int    `json:"seats_used,omitempty"`
	SeatPriceMonthly *string `json:"seat_price_monthly,omitempty"`
}

// teamPostBody is the envelope for all POST /v1/team action requests.
type teamPostBody struct {
	Action string `json:"action"`
	// Remaining fields are action-specific and only serialised when non-zero.
	TeamID      string `json:"team_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	InviteID    string `json:"invite_id,omitempty"`
	TotalSeats  int    `json:"total_seats,omitempty"`
}

func (c *Client) teamPost(ctx context.Context, body teamPostBody) (*TeamActionResponse, error) {
	resp := &TeamActionResponse{}
	_, err := c.doPost(ctx, "/team", body, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListTeams returns every team the authenticated user belongs to, plus any
// outstanding invites addressed to the user's email. Does not consume a credit.
func (c *Client) ListTeams(ctx context.Context) (*ListTeamsResponse, error) {
	resp := &ListTeamsResponse{}
	rateLimits, err := c.doGet(ctx, "/team", nil, resp)
	if err != nil {
		return nil, err
	}
	resp.RateLimits = rateLimits
	return resp, nil
}

// CreateTeam creates a new team. Requires a Business plan. Each account may
// own at most one team.
func (c *Client) CreateTeam(ctx context.Context, name string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "create", Name: name})
}

// InviteTeamMember sends a team invitation to the given email address.
// role must be "member" or "admin". Requires owner or admin privileges.
func (c *Client) InviteTeamMember(ctx context.Context, teamID, email, role string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "invite", TeamID: teamID, Email: email, Role: role})
}

// RemoveTeamMember removes a user from the team and downgrades their plan to
// Starter. Requires owner or admin privileges; owners cannot be removed.
func (c *Client) RemoveTeamMember(ctx context.Context, teamID, userID string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "remove_member", TeamID: teamID, UserID: userID})
}

// CancelTeamInvite cancels an outstanding invite by its ID.
// Requires owner or admin privileges.
func (c *Client) CancelTeamInvite(ctx context.Context, teamID, inviteID string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "cancel_invite", TeamID: teamID, InviteID: inviteID})
}

// ResendTeamInvite resends an existing invite and resets its expiry to 7 days
// from now. Requires owner or admin privileges.
func (c *Client) ResendTeamInvite(ctx context.Context, teamID, inviteID string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "resend_invite", TeamID: teamID, InviteID: inviteID})
}

// PromoteTeamMember changes a member's role. role must be "admin" or "member".
// Only the team owner can demote an admin or promote to admin.
func (c *Client) PromoteTeamMember(ctx context.Context, teamID, userID, role string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "promote", TeamID: teamID, UserID: userID, Role: role})
}

// LeaveTeam removes the authenticated user from the team and downgrades their
// plan to Starter. The team owner cannot leave.
func (c *Client) LeaveTeam(ctx context.Context, teamID string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "leave", TeamID: teamID})
}

// RenameTeam changes the team's display name. Only the team owner can rename.
func (c *Client) RenameTeam(ctx context.Context, teamID, name string) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "rename", TeamID: teamID, Name: name})
}

// SetTeamSeats sets the total seat capacity of the team. totalSeats must be
// at least the number of included seats on the Business plan. Extra seats
// beyond the included base are billed immediately at the per-seat monthly rate.
// Only the team owner can change the seat count.
func (c *Client) SetTeamSeats(ctx context.Context, teamID string, totalSeats int) (*TeamActionResponse, error) {
	return c.teamPost(ctx, teamPostBody{Action: "set_seats", TeamID: teamID, TotalSeats: totalSeats})
}
