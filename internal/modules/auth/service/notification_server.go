// internal/modules/auth/service/notification_service.go
package service

import (
	"context"
	"encoding/json"
	"time"

	"tsu-self/internal/pkg/log"

	"github.com/nats-io/nats.go"
)

type NotificationService struct {
	nc     *nats.Conn
	logger log.Logger
}

type PermissionChangedEvent struct {
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type"` // "role_assigned", "role_revoked", "permission_changed"
	Roles     []string               `json:"roles,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type UserRegisteredEvent struct {
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Username  string                 `json:"username"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func NewNotificationService(nc *nats.Conn, logger log.Logger) *NotificationService {
	return &NotificationService{
		nc:     nc,
		logger: logger,
	}
}

func (s *NotificationService) PublishPermissionChanged(ctx context.Context, userID string, eventType string, roles []string, metadata map[string]interface{}) error {
	event := &PermissionChangedEvent{
		UserID:    userID,
		EventType: eventType,
		Roles:     roles,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := "auth.permission.changed"
	if err := s.nc.Publish(subject, data); err != nil {
		s.logger.ErrorContext(ctx, "发布权限变更事件失败", log.String("user_id", userID), log.String("event_type", eventType), log.Any("error", err))
		return err
	}

	s.logger.InfoContext(ctx, "权限变更事件已发布", log.String("user_id", userID), log.String("event_type", eventType))
	return nil
}

func (s *NotificationService) PublishUserRegistered(ctx context.Context, userID string, metadata map[string]interface{}) error {
	// 从metadata中提取email和username
	email, _ := metadata["email"].(string)
	username, _ := metadata["username"].(string)

	event := &UserRegisteredEvent{
		UserID:    userID,
		Email:     email,
		Username:  username,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := "auth.user.registered"
	if err := s.nc.Publish(subject, data); err != nil {
		s.logger.ErrorContext(ctx, "发布用户注册事件失败",
			log.String("user_id", userID),
			log.String("email", email),
			log.String("username", username),
			log.Any("error", err))
		return err
	}

	s.logger.InfoContext(ctx, "用户注册事件已发布",
		log.String("user_id", userID),
		log.String("email", email),
		log.String("username", username))
	return nil
}
