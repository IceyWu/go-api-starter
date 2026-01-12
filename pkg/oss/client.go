package oss

import (
	"fmt"
	"go-api-starter/internal/config"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	ossClient *oss.Client
	bucket    *oss.Bucket
)

// InitOSS initializes the OSS client
func InitOSS(cfg *config.OSSConfig) error {
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return fmt.Errorf("OSS credentials not configured")
	}

	// Clean endpoint - remove bucket name if present
	endpoint := cfg.Endpoint
	// If endpoint contains bucket name (e.g., lpalette.oss-accelerate.aliyuncs.com)
	// extract the actual endpoint (e.g., oss-accelerate.aliyuncs.com)
	if strings.HasPrefix(endpoint, cfg.Bucket+".") {
		endpoint = strings.TrimPrefix(endpoint, cfg.Bucket+".")
	}
	
	// Ensure endpoint has protocol
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	client, err := oss.New(endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return fmt.Errorf("failed to create OSS client: %w", err)
	}

	bkt, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket: %w", err)
	}

	ossClient = client
	bucket = bkt
	return nil
}

// GetClient returns the OSS client
func GetClient() *oss.Client {
	return ossClient
}

// GetBucket returns the OSS bucket
func GetBucket() *oss.Bucket {
	return bucket
}
