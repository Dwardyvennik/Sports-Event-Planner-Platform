package domain

import (
	"errors"
	"time"
)

const (
	ChannelEmail = "email"
	ChannelMock  = "mock"
)

const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidNotification  = errors.New("invalid notification")
)

type Notification struct {
	ID        string
	UserID    string
	Channel   string
	Subject   string
	Body      string
	Status    string
	CreatedAt time.Time
	SentAt    *time.Time
}

type Reminder struct {
	ID          string
	EventID     string
	UserID      string
	Message     string
	ScheduledAt time.Time
	Status      string
	CreatedAt   time.Time
}
