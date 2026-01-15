package oss

import (
	"fmt"
	"strings"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// MultipartInitResponse represents multipart upload initialization response
type MultipartInitResponse struct {
	UploadID string `json:"upload_id"`
	Key      string `json:"key"`
	Host     string `json:"host"`
}

// PartUploadURL represents presigned URL for uploading a part
type PartUploadURL struct {
	PartNumber int    `json:"part_number"`
	URL        string `json:"url"`
	Expire     int64  `json:"expire"`
}

// CompletePart represents a completed part
type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// InitMultipartUpload initializes a multipart upload
func InitMultipartUpload(key string) (*MultipartInitResponse, error) {
	bkt := GetBucket()
	if bkt == nil {
		return nil, fmt.Errorf("OSS bucket not initialized")
	}

	result, err := bkt.InitiateMultipartUpload(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	// Build host URL
	endpoint := bkt.Client.Config.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	
	var host string
	if strings.HasPrefix(endpoint, bkt.BucketName+".") {
		host = fmt.Sprintf("https://%s", endpoint)
	} else {
		host = fmt.Sprintf("https://%s.%s", bkt.BucketName, endpoint)
	}

	return &MultipartInitResponse{
		UploadID: result.UploadID,
		Key:      result.Key,
		Host:     host,
	}, nil
}

// GeneratePartUploadURL generates presigned URL for uploading a part
func GeneratePartUploadURL(key, uploadID string, partNumber int, expireSeconds int64) (*PartUploadURL, error) {
	bkt := GetBucket()
	if bkt == nil {
		return nil, fmt.Errorf("OSS bucket not initialized")
	}

	// Generate presigned URL for PUT request
	options := []alioss.Option{
		alioss.AddParam("partNumber", fmt.Sprintf("%d", partNumber)),
		alioss.AddParam("uploadId", uploadID),
	}

	signedURL, err := bkt.SignURL(key, alioss.HTTPPut, expireSeconds, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &PartUploadURL{
		PartNumber: partNumber,
		URL:        signedURL,
		Expire:     expireSeconds,
	}, nil
}

// CompleteMultipartUpload completes a multipart upload
func CompleteMultipartUpload(key, uploadID string, parts []CompletePart) error {
	bkt := GetBucket()
	if bkt == nil {
		return fmt.Errorf("OSS bucket not initialized")
	}

	// Convert to OSS SDK format
	var ossParts []alioss.UploadPart
	for _, p := range parts {
		ossParts = append(ossParts, alioss.UploadPart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}

	imur := alioss.InitiateMultipartUploadResult{
		Key:      key,
		UploadID: uploadID,
	}

	_, err := bkt.CompleteMultipartUpload(imur, ossParts)
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

// AbortMultipartUpload aborts a multipart upload
func AbortMultipartUpload(key, uploadID string) error {
	bkt := GetBucket()
	if bkt == nil {
		return fmt.Errorf("OSS bucket not initialized")
	}

	imur := alioss.InitiateMultipartUploadResult{
		Key:      key,
		UploadID: uploadID,
	}

	err := bkt.AbortMultipartUpload(imur)
	if err != nil {
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}

	return nil
}

// ListParts lists uploaded parts
func ListParts(key, uploadID string) ([]CompletePart, error) {
	bkt := GetBucket()
	if bkt == nil {
		return nil, fmt.Errorf("OSS bucket not initialized")
	}

	imur := alioss.InitiateMultipartUploadResult{
		Key:      key,
		UploadID: uploadID,
	}

	result, err := bkt.ListUploadedParts(imur)
	if err != nil {
		return nil, fmt.Errorf("failed to list parts: %w", err)
	}

	var parts []CompletePart
	for _, p := range result.UploadedParts {
		parts = append(parts, CompletePart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}

	return parts, nil
}
