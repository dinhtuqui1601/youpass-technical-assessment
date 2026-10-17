package share

import "time"

// ShareLinkDTO is the public representation of a share link.
type ShareLinkDTO struct {
	ID           string     `json:"id"`
	Code         string     `json:"code"`
	SubmissionID string     `json:"submissionId"`
	Status       string     `json:"status"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
	URL          string     `json:"url"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// UpdateShareLinkRequest is the payload for toggling a link status.
type UpdateShareLinkRequest struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
}

// PublicShareDTO is returned by the public share endpoint.
type PublicShareDTO struct {
	Code    string `json:"code"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
}
