package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session represents an authentication session (a refresh-token family). Only
// the SHA-256 hash of the current refresh token is stored, never the plaintext.
type Session struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// SessionID identifies the token family for reuse detection.
	SessionID string `bson:"sessionId" json:"sessionId"`
	UserID    string `bson:"userId" json:"userId"`

	// RefreshTokenHash is the hash of the currently active refresh token.
	RefreshTokenHash string `bson:"refreshTokenHash" json:"-"`
	// PrevRefreshTokenHash keeps the immediately previous token hash so that
	// legitimate concurrent refreshes (multi-tab) can be distinguished from
	// reuse within a short grace window.
	PrevRefreshTokenHash string    `bson:"prevRefreshTokenHash,omitempty" json:"-"`
	PrevHashGraceUntil   time.Time `bson:"prevHashGraceUntil,omitempty" json:"-"`

	Revoked   bool       `bson:"revoked" json:"revoked"`
	RevokedAt *time.Time `bson:"revokedAt,omitempty" json:"revokedAt,omitempty"`

	ExpiresAt time.Time `bson:"expiresAt" json:"expiresAt"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`

	Metadata map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
}
