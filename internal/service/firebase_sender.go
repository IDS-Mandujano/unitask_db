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

	// Detallar exactamente qué pasó con cada token
	fmt.Printf("\n📊 Firebase Send Report:\n")
	fmt.Printf("  Total tokens: %d\n", len(tokens))
	fmt.Printf("  ✅ Success: %d\n", response.SuccessCount)
	fmt.Printf("  ❌ Failed: %d\n", response.FailureCount)

	successTokens := make([]string, 0)
	failedTokens := make([]string, 0)

	for i, resp := range response.Responses {
		if resp.Success {
			successTokens = append(successTokens, tokens[i][:20]+"...")
		} else if resp.Error != nil {
			failedTokens = append(failedTokens, fmt.Sprintf("token[%d]: %v", i, resp.Error))
		}
	}

	if len(successTokens) > 0 {
		fmt.Printf("  ✅ Tokens entregados: %d\n", len(successTokens))
	}

	if len(failedTokens) > 0 {
		fmt.Printf("  ❌ Tokens fallidos:\n")
		for _, fail := range failedTokens {
			fmt.Printf("     - %s\n", fail)
		}
	}
	fmt.Printf("\n")

	if response.SuccessCount == 0 {
		errors := make([]string, 0, len(response.Responses))
		for i, resp := range response.Responses {
			if !resp.Success && resp.Error != nil {
				errors = append(errors, fmt.Sprintf("token[%d]: %v", i, resp.Error))
			}
		}
		if len(errors) == 0 {
			return fmt.Errorf("send multicast failed: 0/%d entregadas (sin detalles de error)", len(tokens))
		}
		return fmt.Errorf("send multicast failed: 0/%d entregadas (%s)", len(tokens), strings.Join(errors, "; "))
	}

	return nil
}
