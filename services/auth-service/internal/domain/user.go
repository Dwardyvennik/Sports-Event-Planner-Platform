package domain

import "time"

type User struct {
	ID        string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}
