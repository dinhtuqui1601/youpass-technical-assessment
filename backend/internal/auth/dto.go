package auth

import "time"

// RegisterRequest is the payload for creating a demo account.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
}

// LoginRequest is the payload for signing in.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserDTO is the public representation of a user.
type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// LoginResponse is returned after login and after a successful refresh.
type LoginResponse struct {
	AccessToken string  `json:"accessToken"`
	ExpiresIn   int64   `json:"expiresIn"`
	User        UserDTO `json:"user"`
}
