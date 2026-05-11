package usecase

import "context"

type NotificationRepository interface {
	Ping(context.Context) error
}

type NotificationUseCase struct {
	notifications NotificationRepository
}

func NewNotificationUseCase(notifications NotificationRepository) *NotificationUseCase {
	return &NotificationUseCase{notifications: notifications}
}

func (u *NotificationUseCase) Health(ctx context.Context) error {
	if u.notifications == nil {
		return nil
	}
	return u.notifications.Ping(ctx)
}
