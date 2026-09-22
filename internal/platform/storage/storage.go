package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"go-api-starter/internal/config"
)

// ObjectStorage is the provider-neutral object storage contract used by the application.
type ObjectStorage interface {
	PresignPutObject(ctx context.Context, key, contentType string, expires time.Duration) (*PresignedUpload, error)
	PutObject(ctx context.Context, key string, body io.Reader, contentType string) (*ObjectInfo, error)
	CreateMultipartUpload(ctx context.Context, key, contentType string) (*MultipartUpload, error)
	PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int, expires time.Duration) (*PresignedPart, error)
	CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []CompletedPart) error
	AbortMultipartUpload(ctx context.Context, key, uploadID string) error
	ListParts(ctx context.Context, key, uploadID string) ([]CompletedPart, error)
	HeadObject(ctx context.Context, key string) (*ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
	PublicURL(key string) string
}

type PresignedUpload struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Key       string            `json:"key"`
	ExpiresAt int64             `json:"expires_at"`
}

type PresignedPart struct {
	PartNumber int               `json:"part_number"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers,omitempty"`
	ExpiresAt  int64             `json:"expires_at"`
}

type MultipartUpload struct {
	UploadID string `json:"upload_id"`
	Key      string `json:"key"`
}

type CompletedPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

type ObjectInfo struct {
	Key         string
	URL         string
	Size        int64
	ContentType string
	ETag        string
}

type S3Provider struct {
	client        *s3.Client
	presigner     *s3.PresignClient
	bucket        string
	endpoint      string
	publicBaseURL string
	pathStyle     bool
}

func NewS3Provider(ctx context.Context, cfg *config.StorageConfig) (*S3Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage configuration is required")
	}
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("storage endpoint and bucket are required")
	}
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("storage credentials are required")
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		loadOptions = append(loadOptions, awsconfig.WithRegion(cfg.Region))
	}
	// Some S3-compatible providers, including Alibaba Cloud OSS, do not
	// support the AWS SDK's streaming checksum trailer.
	loadOptions = append(loadOptions, awsconfig.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired))
	loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
	))
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load storage client configuration: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
		options.UsePathStyle = cfg.ForcePathStyle
	})
	return &S3Provider{
		client:        client,
		presigner:     s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		endpoint:      strings.TrimRight(cfg.Endpoint, "/"),
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		pathStyle:     cfg.ForcePathStyle,
	}, nil
}

func (p *S3Provider) PresignPutObject(ctx context.Context, key, contentType string, expires time.Duration) (*PresignedUpload, error) {
	input := &s3.PutObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	request, err := p.presigner.PresignPutObject(ctx, input, func(options *s3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return nil, fmt.Errorf("presign object upload: %w", err)
	}
	headers := map[string]string{}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	return &PresignedUpload{Method: request.Method, URL: request.URL, Headers: headers, Key: key, ExpiresAt: time.Now().Add(expires).Unix()}, nil
}

func (p *S3Provider) PutObject(ctx context.Context, key string, body io.Reader, contentType string) (*ObjectInfo, error) {
	input := &s3.PutObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key), Body: body}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if seeker, ok := body.(io.Seeker); ok {
		position, err := seeker.Seek(0, io.SeekCurrent)
		if err == nil {
			end, endErr := seeker.Seek(0, io.SeekEnd)
			_, restoreErr := seeker.Seek(position, io.SeekStart)
			if endErr == nil && restoreErr == nil {
				input.ContentLength = aws.Int64(end)
			}
		}
	}
	result, err := p.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}
	return &ObjectInfo{Key: key, URL: p.PublicURL(key), ContentType: contentType, ETag: aws.ToString(result.ETag)}, nil
}

func (p *S3Provider) CreateMultipartUpload(ctx context.Context, key, contentType string) (*MultipartUpload, error) {
	input := &s3.CreateMultipartUploadInput{Bucket: aws.String(p.bucket), Key: aws.String(key)}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	result, err := p.client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("create multipart upload: %w", err)
	}
	return &MultipartUpload{UploadID: aws.ToString(result.UploadId), Key: key}, nil
}

func (p *S3Provider) PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int, expires time.Duration) (*PresignedPart, error) {
	request, err := p.presigner.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(p.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(partNumber)),
	}, func(options *s3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return nil, fmt.Errorf("presign upload part: %w", err)
	}
	return &PresignedPart{PartNumber: partNumber, URL: request.URL, ExpiresAt: time.Now().Add(expires).Unix()}, nil
}

func (p *S3Provider) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []CompletedPart) error {
	completed := make([]types.CompletedPart, 0, len(parts))
	for _, part := range parts {
		completed = append(completed, types.CompletedPart{ETag: aws.String(part.ETag), PartNumber: aws.Int32(int32(part.PartNumber))})
	}
	_, err := p.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket: aws.String(p.bucket), Key: aws.String(key), UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{Parts: completed},
	})
	if err != nil {
		return fmt.Errorf("complete multipart upload: %w", err)
	}
	return nil
}

func (p *S3Provider) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	_, err := p.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{Bucket: aws.String(p.bucket), Key: aws.String(key), UploadId: aws.String(uploadID)})
	if err != nil {
		return fmt.Errorf("abort multipart upload: %w", err)
	}
	return nil
}

func (p *S3Provider) ListParts(ctx context.Context, key, uploadID string) ([]CompletedPart, error) {
	result, err := p.client.ListParts(ctx, &s3.ListPartsInput{Bucket: aws.String(p.bucket), Key: aws.String(key), UploadId: aws.String(uploadID)})
	if err != nil {
		return nil, fmt.Errorf("list multipart parts: %w", err)
	}
	parts := make([]CompletedPart, 0, len(result.Parts))
	for _, part := range result.Parts {
		parts = append(parts, CompletedPart{PartNumber: int(aws.ToInt32(part.PartNumber)), ETag: aws.ToString(part.ETag)})
	}
	return parts, nil
}

func (p *S3Provider) HeadObject(ctx context.Context, key string) (*ObjectInfo, error) {
	result, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("head object: %w", err)
	}
	return &ObjectInfo{Key: key, Size: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType), ETag: aws.ToString(result.ETag)}, nil
}

func (p *S3Provider) DeleteObject(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (p *S3Provider) PublicURL(key string) string {
	base := p.publicBaseURL
	if base == "" {
		base = p.endpoint
		if parsed, err := url.Parse(base); err == nil && parsed.Host != "" {
			if p.pathStyle {
				parsed.Path = strings.TrimRight(parsed.Path, "/") + "/" + p.bucket
			} else if !strings.HasPrefix(parsed.Host, p.bucket+".") {
				parsed.Host = p.bucket + "." + parsed.Host
			}
			base = strings.TrimRight(parsed.String(), "/")
		} else {
			base = strings.TrimRight(base, "/")
		}
	}
	return strings.TrimRight(base, "/") + "/" + path.Clean("/" + key)[1:]
}
