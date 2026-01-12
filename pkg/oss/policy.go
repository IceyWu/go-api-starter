package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PolicyConfig represents the upload policy configuration
type PolicyConfig struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

// UploadToken represents the upload token response
type UploadToken struct {
	AccessID  string `json:"accessid"`  // 返回给前端时使用小写，前端会转换为 OSSAccessKeyId
	Policy    string `json:"policy"`
	Signature string `json:"signature"`
	Dir       string `json:"dir"`
	Host      string `json:"host"`
	Expire    int64  `json:"expire"`
	Key       string `json:"key"`
}

// GenerateUploadToken generates upload token for client-side direct upload
func GenerateUploadToken(accessKeyID, accessKeySecret, bucket, endpoint, dir, key string, maxSize int64, expireSeconds int64) (*UploadToken, error) {
	now := time.Now()
	expireTime := now.Add(time.Duration(expireSeconds) * time.Second)
	expireTimeStr := expireTime.UTC().Format("2006-01-02T15:04:05Z")

	// Build policy
	policy := PolicyConfig{
		Expiration: expireTimeStr,
		Conditions: []interface{}{
			map[string]string{"bucket": bucket},
			[]interface{}{"content-length-range", 0, maxSize},
			[]interface{}{"starts-with", "$key", dir},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy: %w", err)
	}

	policyBase64 := base64.StdEncoding.EncodeToString(policyJSON)

	// Generate signature
	h := hmac.New(sha1.New, []byte(accessKeySecret))
	h.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Build host URL
	var host string
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	
	// Remove protocol if present in endpoint
	cleanEndpoint := endpoint
	cleanEndpoint = strings.TrimPrefix(cleanEndpoint, "https://")
	cleanEndpoint = strings.TrimPrefix(cleanEndpoint, "http://")
	
	// Check if endpoint already starts with bucket name
	if strings.HasPrefix(cleanEndpoint, bucket+".") {
		// Endpoint already contains bucket (e.g., lpalette.oss-accelerate.aliyuncs.com)
		host = fmt.Sprintf("https://%s", cleanEndpoint)
	} else {
		// Standard endpoint format (e.g., oss-cn-hangzhou.aliyuncs.com)
		host = fmt.Sprintf("https://%s.%s", bucket, cleanEndpoint)
	}

	return &UploadToken{
		AccessID:  accessKeyID,
		Policy:    policyBase64,
		Signature: signature,
		Dir:       dir,
		Host:      host,
		Expire:    expireTime.Unix(),
		Key:       key,
	}, nil
}
