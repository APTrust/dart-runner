package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/APTrust/dart-runner/constants"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	messageChannel     chan *EventMessage
	minioClient        *minio.Client
	storageService     *StorageService
	totalBytesToUpload int64
	bytesUploaded      int64
	filesUploaded      int64
	etags              map[string]string
}

// NewS3Client creates a new S3 client. If useSSL is true (and it
// should be in all environments outside local dev) this will connect
// via HTTPS. On local dev machines, when talking to Minio, useSSL should
// be false.
//
// Param messageChannel is a channel through which this client can
// send progress info back the UI. When running DART, the messageChannel
// should not be nil. When running CLI jobs from Dart Runner, this channel
// should be nil because there's no UI to report to.
func NewS3Client(ss *StorageService, useSSL bool, messageChannel chan *EventMessage) (*S3Client, error) {
	accessKeyId := ss.GetLogin()
	secretKey := ss.GetPassword()
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretKey, ""),
		Secure: useSSL,
	}
	client, err := minio.New(ss.HostAndPort(), options)
	if err != nil {
		return nil, err
	}
	return &S3Client{
		messageChannel:     messageChannel,
		minioClient:        client,
		storageService:     ss,
		totalBytesToUpload: 0,
		bytesUploaded:      0,
		filesUploaded:      0,
		etags:              make(map[string]string),
	}, nil
}

// Upload uploads a file or directory from source to destination
// on the remote server.
func (c *S3Client) Upload(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	c.totalBytesToUpload = int64(0)
	c.bytesUploaded = int64(0)
	c.filesUploaded = int64(0)
	c.etags = make(map[string]string)

	if info.IsDir() {
		c.totalBytesToUpload, err = GetUploadPayloadSize(source)
		if err != nil {
			return err
		}
		return c.uploadDirectory(source)
	}
	s3Key := filepath.Base(source)
	return c.uploadFile(source, s3Key)
}

// uploadDirectory recursively uploads a directory to the remote server
func (c *S3Client) uploadDirectory(sourceFile string) error {
	s3KeyPrefix := filepath.Base(sourceFile)
	return filepath.Walk(sourceFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourceFile, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for local path %s: %w", sourceFile, err)
		}
		// Create the S3 key for this file and then upload it.
		s3Key := filepath.ToSlash(filepath.Join(s3KeyPrefix, relPath))
		if info.Mode().IsRegular() {
			err = c.uploadFile(path, s3Key)
			if err != nil {
				return err
			}
		} else {
			Dart.Log.Warningf("S3 client is skipping upload of %s because it's not a regular file", path)
		}
		return nil
	})
}

// uploadFile uploads a single file to the remote S3 service
func (c *S3Client) uploadFile(sourceFile, s3Key string) error {
	remoteURL := c.storageService.URL(s3Key)
	Dart.Log.Infof("Starting S3 upload %s to %s", sourceFile, remoteURL)
	putOptions := minio.PutObjectOptions{}
	if c.messageChannel != nil {
		progress := NewStreamProgress(c.totalBytesToUpload, c.messageChannel)
		putOptions = minio.PutObjectOptions{
			Progress: progress,
		}
		c.messageChannel <- StartEvent(constants.StageUpload, fmt.Sprintf("Uploading to %s", c.storageService.Name))
	}
	uploadInfo, err := c.minioClient.FPutObject(
		context.Background(),
		c.storageService.Bucket,
		s3Key,
		sourceFile,
		putOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to upload %s to %s: %w", sourceFile, remoteURL, err)
	} else {
		c.filesUploaded += 1
		c.bytesUploaded += uploadInfo.Size
		c.etags[remoteURL] = uploadInfo.ETag
		Dart.Log.Infof("finished s3 upload of file %s; got e-tag %s", sourceFile, uploadInfo.ETag)
	}
	return nil
}

func (c *S3Client) FilesUploaded() int64 {
	return c.filesUploaded
}

func (c *S3Client) BytesUploaded() int64 {
	return c.bytesUploaded
}

func (c *S3Client) PayloadSize() int64 {
	return c.totalBytesToUpload
}

func (c *S3Client) EtagMap() map[string]string {
	etagMap := make(map[string]string)
	for key, value := range c.etags {
		etagMap[key] = value
	}
	return etagMap
}

// ListBuckets returns a list of existing S3 buckets.
func (c *S3Client) ListBuckets() ([]minio.BucketInfo, error) {
	return c.minioClient.ListBuckets(context.Background())
}

// ListObjects lists the objects in the specified bucket.
func (c *S3Client) ListObjects(bucketName, prefix string, opts minio.ListObjectsOptions) []minio.ObjectInfo {
	objects := make([]minio.ObjectInfo, 0)
	for object := range c.minioClient.ListObjects(context.Background(), bucketName, opts) {
		objects = append(objects, object)
	}
	return objects
}

// GetObject returns the object with the specified bucket name and key.
func (c *S3Client) GetObject(bucket, key string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return c.minioClient.GetObject(context.Background(), bucket, key, opts)
}
