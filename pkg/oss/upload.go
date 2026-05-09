package oss

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// UploadResult represents the result of a file upload
type UploadResult struct {
	Key      string `json:"key"`
	URL      string `json:"url"`
	MD5      string `json:"md5"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// CalculateMD5 calculates MD5 hash of a file
func CalculateMD5(file multipart.File) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateMD5FromReader calculates MD5 hash from an io.Reader
func CalculateMD5FromReader(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GenerateObjectKey generates a unique object key for OSS
func GenerateObjectKey(filename, userID string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("uploads/%s/%s%s", userID, timestamp, ext)
}

// GenerateObjectKeyWithPath generates a unique object key with custom path
func GenerateObjectKeyWithPath(filename, path string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Format("20060102150405")
	baseName := strings.TrimSuffix(filepath.Base(filename), ext)
	return fmt.Sprintf("%s/%s_%s%s", path, baseName, timestamp, ext)
}

// UploadFile uploads a file to OSS
func UploadFile(file multipart.File, header *multipart.FileHeader, objectKey string) (*UploadResult, error) {
	bkt := GetBucket()
	if bkt == nil {
		return nil, fmt.Errorf("OSS bucket not initialized")
	}

	// Calculate MD5
	md5Hash, err := CalculateMD5(file)
	if err != nil {
		return nil, err
	}

	// Get file size
	size := header.Size

	// Upload to OSS
	options := []alioss.Option{
		alioss.ContentType(header.Header.Get("Content-Type")),
	}

	err = bkt.PutObject(objectKey, file, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to OSS: %w", err)
	}

	// Generate URL
	url := GetObjectURL(objectKey)

	return &UploadResult{
		Key:      objectKey,
		URL:      url,
		MD5:      md5Hash,
		Size:     size,
		MimeType: header.Header.Get("Content-Type"),
	}, nil
}

// UploadFromReader uploads content from an io.Reader to OSS
func UploadFromReader(reader io.Reader, objectKey string, contentType string) (*UploadResult, error) {
	bkt := GetBucket()
	if bkt == nil {
		return nil, fmt.Errorf("OSS bucket not initialized")
	}

	// Upload to OSS
	options := []alioss.Option{
		alioss.ContentType(contentType),
	}

	err := bkt.PutObject(objectKey, reader, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to OSS: %w", err)
	}

	// Generate URL
	url := GetObjectURL(objectKey)

	return &UploadResult{
		Key:      objectKey,
		URL:      url,
		MimeType: contentType,
	}, nil
}

// DeleteFile deletes a file from OSS
func DeleteFile(objectKey string) error {
	bkt := GetBucket()
	if bkt == nil {
		return fmt.Errorf("OSS bucket not initialized")
	}

	err := bkt.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("failed to delete file from OSS: %w", err)
	}

	return nil
}

// GetObjectURL generates the public URL for an object
func GetObjectURL(objectKey string) string {
	bkt := GetBucket()
	if bkt == nil {
		return ""
	}

	endpoint := bkt.Client.Config.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	if strings.HasPrefix(endpoint, bkt.BucketName+".") {
		return fmt.Sprintf("https://%s/%s", endpoint, objectKey)
	}
	return fmt.Sprintf("https://%s.%s/%s", bkt.BucketName, endpoint, objectKey)
}

// GeneratePresignedURL generates a presigned URL for temporary access
func GeneratePresignedURL(objectKey string, expireSeconds int64) (string, error) {
	bkt := GetBucket()
	if bkt == nil {
		return "", fmt.Errorf("OSS bucket not initialized")
	}

	signedURL, err := bkt.SignURL(objectKey, alioss.HTTPGet, expireSeconds)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return signedURL, nil
}

// CheckFileExists checks if a file exists in OSS
func CheckFileExists(objectKey string) (bool, error) {
	bkt := GetBucket()
	if bkt == nil {
		return false, fmt.Errorf("OSS bucket not initialized")
	}

	exists, err := bkt.IsObjectExist(objectKey)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return exists, nil
}

// CopyFile copies a file within OSS
func CopyFile(sourceKey, destKey string) error {
	bkt := GetBucket()
	if bkt == nil {
		return fmt.Errorf("OSS bucket not initialized")
	}

	_, err := bkt.CopyObject(sourceKey, destKey)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
