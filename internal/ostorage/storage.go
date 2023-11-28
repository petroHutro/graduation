package ostorage

import (
	"bytes"
	"context"
	"fmt"
	"graduation/internal/config"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectSt struct {
	ObjectName  string
	FilePath    string
	ContentType string
}

type storageData struct {
	bucketName string
}

type Storage struct {
	client *minio.Client
	storageData
}

func Connect(conf *config.ObjectStorage) (*Storage, error) {
	client, err := minio.New(conf.StorageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.StorageAccessKey, conf.StorageSecretKey, ""),
		Secure: conf.StorageUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}

	storage := Storage{client: client, storageData: storageData{bucketName: conf.StorageBucketName}}
	if err := storage.creatBucket(); err != nil {
		return nil, fmt.Errorf("cannot creat bucket: %w", err)
	}

	return &storage, nil
}

func (s *Storage) creatBucket() error {
	exists, err := s.client.BucketExists(context.Background(), s.bucketName)
	if err != nil {
		return fmt.Errorf("cannot exists: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(context.Background(), s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("cannot creat bucket: %w", err)
		}
	}

	return nil
}

func (s *Storage) Set(objectName string, fileContent []byte) error {
	_, err := s.client.PutObject(
		context.Background(),
		s.bucketName,
		objectName,
		bytes.NewReader(fileContent),
		-1,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)

	if err != nil {
		return fmt.Errorf("cannot set file: %w", err)
	}

	return nil
}

func (s *Storage) Delete(objectName string) error {
	err := s.client.RemoveObject(
		context.Background(),
		s.bucketName,
		objectName,
		minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("cannot set file: %w", err)
	}

	return nil
}

func (s *Storage) Get(objectName string) (string, error) {
	expiration := 1 * time.Hour
	presignedURL, err := s.client.PresignedGetObject(context.Background(), s.bucketName, objectName, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("cannot set file: %w", err)
	}

	return presignedURL.String(), nil
}
