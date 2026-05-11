package domain

import "time"

type Notification struct {
	ID        string
	UserID    string
	Channel   string
	Subject   string
	Status    string
	CreatedAt time.Time
	SentAt    *time.Time
}
