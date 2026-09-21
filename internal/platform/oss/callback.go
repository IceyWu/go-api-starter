package oss

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CallbackRequest represents the OSS callback request
type CallbackRequest struct {
	Bucket   string `json:"bucket"`
	Object   string `json:"object"`
	ETag     string `json:"etag"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// VerifyOSSCallback verifies the OSS callback signature
func VerifyOSSCallback(r *http.Request, accessKeySecret string) (bool, error) {
	// Get authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, fmt.Errorf("missing authorization header")
	}

	// Get public key URL
	pubKeyURL := r.Header.Get("x-oss-pub-key-url")
	if pubKeyURL == "" {
		return false, fmt.Errorf("missing x-oss-pub-key-url header")
	}

	// Decode public key URL
	pubKeyURLBytes, err := base64.StdEncoding.DecodeString(pubKeyURL)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key URL: %w", err)
	}

	// Verify public key URL is from OSS
	pubKeyURLStr := string(pubKeyURLBytes)
	if !strings.HasPrefix(pubKeyURLStr, "https://gosspublic.alicdn.com/") &&
		!strings.HasPrefix(pubKeyURLStr, "https://gosspublic.alicdn.com/") {
		return false, fmt.Errorf("invalid public key URL")
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read request body: %w", err)
	}

	// Build string to sign
	authPath := r.URL.Path
	if r.URL.RawQuery != "" {
		authPath = authPath + "?" + r.URL.RawQuery
	}

	strToSign := fmt.Sprintf("%s\n%s", authPath, string(body))

	// Calculate MD5 of body
	h := md5.New()
	h.Write(body)
	bodyMD5 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Verify MD5
	contentMD5 := r.Header.Get("Content-MD5")
	if contentMD5 != bodyMD5 {
		return false, fmt.Errorf("content MD5 mismatch")
	}

	// For simple verification, we can use HMAC-SHA1 with access key secret
	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	mac.Write([]byte(strToSign))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Extract signature from authorization header
	signature := strings.TrimPrefix(authHeader, "OSS ")

	return signature == expectedSignature, nil
}

// ParseCallbackBody parses the callback request body
func ParseCallbackBody(body string) (*CallbackRequest, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse callback body: %w", err)
	}

	return &CallbackRequest{
		Bucket:   values.Get("bucket"),
		Object:   values.Get("object"),
		ETag:     values.Get("etag"),
		Size:     parseInt64(values.Get("size")),
		MimeType: values.Get("mimeType"),
	}, nil
}

func parseInt64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}
