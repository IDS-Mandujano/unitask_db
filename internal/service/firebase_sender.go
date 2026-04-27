package service

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"unitask-api/internal/domain"
)

type FirebaseSender struct {
	client *messaging.Client
}

func NewFirebaseSender(credentialsPath string) (domain.NotificationSender, error) {
	ctx := context.Background()

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app init failed: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging init failed: %w", err)
	}

	return &FirebaseSender{client: client}, nil
}

func (s *FirebaseSender) SendToTokens(tokens []string, title string, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	ctx := context.Background()
	response, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return fmt.Errorf("send multicast failed: %w", err)
	}

	if response.SuccessCount == 0 {
		errors := make([]string, 0, len(response.Responses))
		for i, resp := range response.Responses {
			if !resp.Success && resp.Error != nil {
				errors = append(errors, fmt.Sprintf("token[%d]: %v", i, resp.Error))
			}
		}
		if len(errors) == 0 {
			return fmt.Errorf("send multicast failed: 0/%d entregadas", len(tokens))
		}
		return fmt.Errorf("send multicast failed: 0/%d entregadas (%s)", len(tokens), strings.Join(errors, "; "))
	}

	return nil
}
