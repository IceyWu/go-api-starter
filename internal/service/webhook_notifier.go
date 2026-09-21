package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/logger"
)

// WebhookNotifier handles webhook notifications
type WebhookNotifier struct {
	httpClient *http.Client
	maxRetries int
}

// NewWebhookNotifier creates a new WebhookNotifier
func NewWebhookNotifier() *WebhookNotifier {
	return &WebhookNotifier{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		maxRetries: 3,
	}
}

// WebhookPayload represents the webhook callback payload
type WebhookPayload struct {
	TaskID  string                    `json:"task_id"`
	Status  string                    `json:"status"`
	Results []model.TranscodingResult `json:"results,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// NotifyWebhook sends a webhook notification with retry logic
func (w *WebhookNotifier) NotifyWebhook(webhookURL string, payload WebhookPayload) error {
	if webhookURL == "" {
		return nil // No webhook configured
	}

	logger.Log.Infof("Sending webhook notification to: %s", webhookURL)

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Retry logic
	var lastErr error
	for i := 0; i < w.maxRetries; i++ {
		if i > 0 {
			// Exponential backoff: 1s, 2s, 4s
			waitTime := time.Duration(1<<uint(i-1)) * time.Second
			logger.Log.Infof("Retrying webhook notification after %v (attempt %d/%d)", waitTime, i+1, w.maxRetries)
			time.Sleep(waitTime)
		}

		// Send HTTP POST request
		resp, err := w.httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to send webhook request: %w", err)
			logger.Log.Warnf("Webhook request failed (attempt %d/%d): %v", i+1, w.maxRetries, err)
			continue
		}

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logger.Log.Infof("Webhook notification sent successfully: %s (status: %d)", webhookURL, resp.StatusCode)
			resp.Body.Close()
			return nil
		}

		lastErr = fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
		logger.Log.Warnf("Webhook returned status %d (attempt %d/%d)", resp.StatusCode, i+1, w.maxRetries)
		resp.Body.Close()
	}

	logger.Log.Errorf("Failed to send webhook notification after %d attempts: %v", w.maxRetries, lastErr)
	return lastErr
}
