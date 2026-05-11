package notificationv1

type SendNotificationRequest struct {
	UserId  string `json:"user_id,omitempty"`
	Channel string `json:"channel,omitempty"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

type SendReminderRequest struct {
	EventId     string `json:"event_id,omitempty"`
	UserId      string `json:"user_id,omitempty"`
	Message     string `json:"message,omitempty"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

type NotificationResponse struct {
	NotificationId string `json:"notification_id,omitempty"`
	Status         string `json:"status,omitempty"`
}

type GetNotificationsRequest struct {
	UserId string `json:"user_id,omitempty"`
}

type Notification struct {
	Id      string `json:"id,omitempty"`
	UserId  string `json:"user_id,omitempty"`
	Channel string `json:"channel,omitempty"`
	Subject string `json:"subject,omitempty"`
	Status  string `json:"status,omitempty"`
}

type GetNotificationsResponse struct {
	Notifications []*Notification `json:"notifications,omitempty"`
}
