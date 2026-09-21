package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/logger"
)

// WebhookNotifier handles webhook notifications
type WebhookNotifier struct {
	httpClient *http.Client
	maxRetries int
}

const maxWebhookResponseBytes = 1 << 20

var errUnsafeWebhookURL = errors.New("unsafe webhook URL")

// NewWebhookNotifier creates a new WebhookNotifier
func NewWebhookNotifier() *WebhookNotifier {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = safeWebhookDialContext
	return &WebhookNotifier{
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
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
func (w *WebhookNotifier) NotifyWebhook(ctx context.Context, webhookURL string, payload WebhookPayload) error {
	if webhookURL == "" {
		return nil // No webhook configured
	}
	u, err := validateWebhookURL(webhookURL)
	if err != nil {
		return err
	}

	logger.Log.Info("sending webhook notification", "host", u.Host)

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
			timer := time.NewTimer(waitTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		// Send HTTP POST request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := w.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send webhook request: %w", err)
			logger.Log.Warnf("Webhook request failed (attempt %d/%d): %v", i+1, w.maxRetries, err)
			continue
		}

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			_, _ = io.CopyN(io.Discard, resp.Body, maxWebhookResponseBytes)
			logger.Log.Info("webhook notification sent successfully", "host", u.Host, "status", resp.StatusCode)
			resp.Body.Close()
			return nil
		}

		lastErr = fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
		logger.Log.Warn("webhook returned non-2xx status", "status", resp.StatusCode, "attempt", i+1, "max_attempts", w.maxRetries)
		resp.Body.Close()
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return lastErr
		}
	}

	logger.Log.Errorf("Failed to send webhook notification after %d attempts: %v", w.maxRetries, lastErr)
	return lastErr
}

func validateWebhookURL(rawURL string) (*url.URL, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return nil, fmt.Errorf("%w: URL must use http/https and include a hostname", errUnsafeWebhookURL)
	}
	if strings.EqualFold(u.Hostname(), "localhost") || isUnsafeWebhookIP(net.ParseIP(u.Hostname())) {
		return nil, fmt.Errorf("%w: private or loopback address", errUnsafeWebhookURL)
	}
	return u, nil
}

func safeWebhookDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if isUnsafeWebhookIP(ip) {
			return nil, fmt.Errorf("%w: resolved private or loopback address", errUnsafeWebhookURL)
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("%w: hostname has no addresses", errUnsafeWebhookURL)
	}
	return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
}

func isUnsafeWebhookIP(ip net.IP) bool {
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast())
}
