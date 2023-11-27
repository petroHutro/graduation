package objectstorage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	// accessKey = "ZR8GVPArBksl2AKhSH7R"
	// secretKey = "r7EA3uoh7JSwSmSZfVB43s0Yrmkfpr3u7Jb4XnDF"
	accessKey = "ZRpj735BW7WAi0Wt2lYq"
	secretKey = "RuoeY2yEFbaB5TfBDRVOAv6wMGCCUY2iNK2OPTop"
	// accessKey = "minioadmin"
	// secretKey = "minioadmin"

	endpoint   = "127.0.0.1:9000"
	useSSL     = false
	bucketName = "event123"
)

type ObjectSt struct {
	ObjectName  string
	FilePath    string
	ContentType string
}

type Storage struct {
	Client *minio.Client
}

func Connect() (*Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}

	storage := Storage{Client: client}
	if err := storage.creatBucket(); err != nil {
		return nil, fmt.Errorf("cannot creat bucket: %w", err)
	}

	return &storage, nil
}

func (s *Storage) creatBucket() error {
	exists, err := s.Client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return fmt.Errorf("cannot exists: %w", err)
	}

	if !exists {
		err = s.Client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("cannot creat bucket: %w", err)
		}
	}

	return nil
}

func (s *Storage) Set(objectName string, fileContent []byte) error {
	// func (s *Storage) Set(objectName, filePath, contentType string) error {
	// _, err := s.Client.FPutObject(
	// 	context.Background(),
	// 	bucketName,
	// 	objectName,
	// 	filePath,
	// 	minio.PutObjectOptions{ContentType: contentType},
	// )

	_, err := s.Client.PutObject(
		context.Background(),
		bucketName,
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
	err := s.Client.RemoveObject(
		context.Background(),
		bucketName,
		objectName,
		minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("cannot set file: %w", err)
	}

	return nil
}

func (s *Storage) Get(objectName string) (string, error) {
	expiration := 1 * time.Hour
	presignedURL, err := s.Client.PresignedGetObject(context.Background(), bucketName, objectName, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("cannot set file: %w", err)
	}

	return presignedURL.String(), nil
}
