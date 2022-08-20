package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type StorageClient struct {
	config *Config
}

func NewStorageClient(config *Config) *StorageClient {
	return &StorageClient{
		config: config,
	}
}

var ErrNoSuchKey = errors.New("no such key")

func (s *StorageClient) newClient() (*s3.Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", s.config.R2AccountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.config.R2AccessKeyId, s.config.R2AccessKeySecret, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return client, nil
}

func (s *StorageClient) Put(bucket string, key string, data []byte) error {
	client, err := s.newClient()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("error putting object in bucket %v with key %v: %w", bucket, key, err)
	}
	return nil
}

func (s *StorageClient) Get(bucket string, key string) ([]byte, error) {
	client, err := s.newClient()
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}
	object, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKeyErr *types.NoSuchKey
		if errors.As(err, &noSuchKeyErr) {
			return nil, ErrNoSuchKey
		} else {
			return nil, fmt.Errorf("error getting object in bucket %v with key %v: %w", bucket, key, err)
		}
	}
	defer object.Body.Close()
	body, err := io.ReadAll(object.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading object body: %w", err)
	}
	return body, nil
}
