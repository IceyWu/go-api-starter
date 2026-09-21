package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateWebhookURLRejectsUnsafeTargets(t *testing.T) {
	for _, rawURL := range []string{
		"file:///etc/passwd",
		"http://localhost/callback",
		"http://127.0.0.1/callback",
		"http://169.254.169.254/latest/meta-data",
	} {
		t.Run(rawURL, func(t *testing.T) {
			_, err := validateWebhookURL(rawURL)
			require.ErrorIs(t, err, errUnsafeWebhookURL)
		})
	}
}

func TestWebhookNotifierHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewWebhookNotifier().NotifyWebhook(ctx, "https://example.com/callback", WebhookPayload{TaskID: "task-1"})
	require.True(t, errors.Is(err, context.Canceled), "expected context cancellation, got %v", err)
}
