// Package zoom returns the Zoom meeting link to use for a TA<->student session.
//
// Configuration is two env vars:
//   - ZOOM_JOIN_URL  (required) the URL students click to join (typically the
//     TA's Personal Meeting Room invite link, e.g.
//     https://us05web.zoom.us/j/1234567890?pwd=AbCdEf123).
//   - ZOOM_START_URL (optional) host launcher link; falls back to ZOOM_JOIN_URL
//     when unset.
//
// CreateMeeting parses the meeting ID and passcode out of the URL so the rest
// of the backend (DB row, SSE payload, modals) can display them.
package zoom

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
)

// Meeting is the subset of Zoom meeting fields the rest of the backend cares about.
type Meeting struct {
	MeetingID string `json:"meeting_id"`
	JoinURL   string `json:"join_url"`
	StartURL  string `json:"start_url"`
	Passcode  string `json:"passcode"`
}

// Client provisions Zoom meetings.
type Client struct{}

// NewClient returns a Zoom client.
func NewClient() *Client { return &Client{} }

// DefaultClient is shared across handlers.
var DefaultClient = NewClient()

// ErrNotConfigured is returned when ZOOM_JOIN_URL is missing.
var ErrNotConfigured = errors.New("zoom: ZOOM_JOIN_URL is not set")

// IsConfigured reports whether ZOOM_JOIN_URL is set.
func IsConfigured() bool {
	return strings.TrimSpace(os.Getenv("ZOOM_JOIN_URL")) != ""
}

// CreateMeeting returns the configured Zoom meeting. The topic argument is
// accepted for API symmetry with future per-meeting integrations but is not
// used; every session points at the same Personal Meeting Room URL.
func (c *Client) CreateMeeting(_ context.Context, _ string) (Meeting, error) {
	join := strings.TrimSpace(os.Getenv("ZOOM_JOIN_URL"))
	if join == "" {
		return Meeting{}, ErrNotConfigured
	}
	start := strings.TrimSpace(os.Getenv("ZOOM_START_URL"))
	if start == "" {
		start = join
	}
	return Meeting{
		MeetingID: extractMeetingIDFromURL(join),
		JoinURL:   join,
		StartURL:  start,
		Passcode:  extractPasscodeFromURL(join),
	}, nil
}

// extractMeetingIDFromURL pulls the numeric meeting ID out of the path of a
// standard Zoom URL like https://us05web.zoom.us/j/1234567890?pwd=abc.
// Returns "" if the URL doesn't follow that shape — the rest of the system
// tolerates an empty MeetingID and just won't display it.
func extractMeetingIDFromURL(s string) string {
	parsed, err := url.Parse(s)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i, part := range parts {
		if (part == "j" || part == "s" || part == "my") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if len(parts) > 0 {
		if last := parts[len(parts)-1]; last != "" {
			return last
		}
	}
	return ""
}

// extractPasscodeFromURL pulls the ?pwd= parameter out of a Zoom URL.
func extractPasscodeFromURL(s string) string {
	parsed, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return parsed.Query().Get("pwd")
}
