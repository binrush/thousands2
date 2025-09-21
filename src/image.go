package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	S3_ENDPOINT = "https://s3.timeweb.cloud"
	S3_BUCKET   = "302f9aa7-62c4d4d3-ccfd-4077-86c8-cca52e0da376"
)

type ImageManager interface {
	Upload(ctx context.Context, imageData []byte, key string) error
}

type S3ImageManager struct {
	s3Client *s3.Client
}

func NewS3ImageManager(accessKey, secretKey, endpoint string, ctx context.Context) (*S3ImageManager, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
		o.Region = "ru-1-ru"
	})

	return &S3ImageManager{s3Client: s3Client}, nil
}

func (im *S3ImageManager) Upload(ctx context.Context, imageData []byte, key string) error {
	_, err := im.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(S3_BUCKET),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to S3: %v", err)
	}

	return nil
}
