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

	endpoint := cfg.Endpoint
	// Remove protocol if present (OSS SDK will add it automatically)
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	// 确保使用 HTTPS（预签名 URL 等会继承 endpoint 的协议）
	client, err := oss.New("https://"+endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
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
