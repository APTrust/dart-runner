package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UploadOperation struct {
	Errors         map[string]string  `json:"errors"`
	PayloadSize    int64              `json:"payloadSize"`
	Result         *OperationResult   `json:"result"`
	SourceFiles    []string           `json:"sourceFiles"`
	StorageService *StorageService    `json:"storageService"`
	MessageChannel chan *EventMessage `json:"-"`
}

func NewUploadOperation(ss *StorageService, files []string) *UploadOperation {
	opResult := NewOperationResult("upload", "Uploader - "+constants.AppVersion)
	if ss != nil {
		opResult.RemoteTargetName = ss.Name
	}
	return &UploadOperation{
		Errors:         make(map[string]string),
		Result:         opResult,
		SourceFiles:    files,
		StorageService: ss,
	}
}

func (u *UploadOperation) Validate() bool {
	u.Errors = make(map[string]string)
	if u.StorageService == nil {
		u.Errors["UploadOperation.StorageService"] = "UploadOperation requires a StorageService"
	} else if !u.StorageService.Validate() {
		for key, errMsg := range u.StorageService.Errors {
			ssKeyName := "StorageService." + key
			u.Errors[ssKeyName] = errMsg
		}
	}
	if len(u.SourceFiles) == 0 {
		u.Errors["UploadOperation.SourceFiles"] = "UploadOperation requires one or more files to upload"
	}
	missingFiles := make([]string, 0)
	for _, file := range u.SourceFiles {
		if !util.FileExists(file) {
			missingFiles = append(missingFiles, file)
		}
	}
	if len(missingFiles) > 0 {
		u.Errors["UploadOperation.SourceFiles"] = fmt.Sprintf("UploadOperation source files are missing: %s", strings.Join(missingFiles, ";"))
	}
	for key, value := range u.Errors {
		Dart.Log.Errorf("%s: %s", key, value)
	}
	return len(u.Errors) == 0
}

func (u *UploadOperation) CalculatePayloadSize() error {
	u.PayloadSize = 0
	for _, fileOrDir := range u.SourceFiles {
		stat, err := os.Stat(fileOrDir)
		if err != nil {
			Dart.Log.Errorf("UploadOperation.CalculatePayloadSize - can't stat %s: %v", fileOrDir, err)
			return err
		}
		if stat.IsDir() {
			children, err := util.RecursiveFileList(fileOrDir, false)
			if err != nil {
				Dart.Log.Errorf("UploadOperation.CalculatePayloadSize - can't recusively list %s: %v", fileOrDir, err)
				return err
			}
			for _, child := range children {
				if !child.FileInfo.IsDir() {
					u.PayloadSize += child.FileInfo.Size()
				}
			}
		} else {
			u.PayloadSize += stat.Size()
		}
	}
	return nil
}

func (u *UploadOperation) DoUpload(messageChannel chan *EventMessage) bool {
	ok := false
	switch u.StorageService.Protocol {
	case constants.ProtocolS3:
		ok = u.sendToS3(messageChannel)
	case constants.ProtocolSFTP:
		ok = u.sendToSFTP(messageChannel)
	default:
		u.Errors["Protocol"] = fmt.Sprintf("Unsupported upload protocol: %s", u.StorageService.Protocol)
	}
	if len(u.Errors) > 0 {
		Dart.Log.Error("One or more errors occurred while uploading to %s service %s at %s", u.StorageService.Protocol, u.StorageService.Name, u.StorageService.HostAndPort())
	}
	for key, value := range u.Errors {
		Dart.Log.Errorf("%s: %s", key, value)
	}
	return ok
}

// upload an item to s3 bucket. If messageChannel is not nil,
// the uploader will send progress updates through it. Otherwise,
// no progress updates.
func (u *UploadOperation) sendToS3(messageChannel chan *EventMessage) bool {
	accessKeyId := u.StorageService.GetLogin()
	secretKey := u.StorageService.GetPassword()
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretKey, ""),
		Secure: u.useSSL(),
	}
	client, err := minio.New(u.StorageService.HostAndPort(), options)
	if err != nil {
		u.Errors[u.StorageService.Name] = fmt.Sprintf("Error connecting to S3: %s", err.Error())
		return false
	}
	allSucceeded := true
	for _, sourceFile := range u.SourceFiles {
		info, err := os.Stat(sourceFile)
		if err != nil {
			u.Errors[sourceFile] = fmt.Sprintf("Error accessing file %s: %s", sourceFile, err.Error())
			allSucceeded = false
			continue
		}
		if info.IsDir() {
			err = u.sendDirectoryToS3(client, sourceFile, messageChannel)
			if err != nil {
				allSucceeded = false
			}
		} else {
			s3Key := filepath.Base(sourceFile)
			succeeded := u.sendFileToS3(client, sourceFile, s3Key, messageChannel)
			if !succeeded {
				allSucceeded = false
			}
		}
	}
	return allSucceeded
}

// uploadDirectory recursively uploads a directory to the remote server
func (u *UploadOperation) sendDirectoryToS3(client *minio.Client, sourceFile string, messageChannel chan *EventMessage) error {
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
			succeeded := u.sendFileToS3(client, path, s3Key, messageChannel)
			if !succeeded {
				key := fmt.Sprintf("%s - %s", u.StorageService.Name, s3Key)
				return fmt.Errorf(u.Errors[key])
			}
		} else {
			Dart.Log.Warningf("S3 client is skipping upload of %s because it's not a regular file", path)
		}

		return nil
	})
}

func (u *UploadOperation) sendFileToS3(client *minio.Client, sourceFile, s3Key string, messageChannel chan *EventMessage) bool {
	succeeded := true
	u.Result.RemoteURL = u.StorageService.URL(s3Key)
	Dart.Log.Infof("Starting S3 upload %s to %s", sourceFile, u.Result.RemoteURL)
	putOptions := minio.PutObjectOptions{}
	if messageChannel != nil {
		progress := NewStreamProgress(u.PayloadSize, messageChannel)
		putOptions = minio.PutObjectOptions{
			Progress: progress,
		}
		messageChannel <- StartEvent(constants.StageUpload, fmt.Sprintf("Uploading to %s", u.StorageService.Name))
	}
	uploadInfo, err := client.FPutObject(
		context.Background(),
		u.StorageService.Bucket,
		s3Key,
		sourceFile,
		putOptions,
	)
	if err != nil {
		key := fmt.Sprintf("%s - %s", u.StorageService.Name, s3Key)
		u.Errors[key] = fmt.Sprintf("Error copying %s to S3: %s", sourceFile, err.Error())
		Dart.Log.Error(u.Errors[key])
		succeeded = false
	} else {
		u.Result.RemoteChecksum = uploadInfo.ETag
		Dart.Log.Infof("Finished S3 upload of file %s; got e-tag %s", sourceFile, uploadInfo.ETag)
	}
	return succeeded
}

// upload an item to SFTP server. If messageChannel is not nil,
// the uploader will send progress updates through it. Otherwise,
// no progress updates.
func (u *UploadOperation) sendToSFTP(messageChannel chan *EventMessage) bool {

	// Set up StreamProgress so the uploader can send status
	// info back to the progress bar. We only do this when
	// messageChannel is not nil, which means we're running
	// the DART 3 GUI. For command-line/unattended Dart Runner
	// jobs, Dart Runner will pass nil as messageChannel param.
	var progress *StreamProgress
	if messageChannel != nil {
		progress = NewStreamProgress(u.PayloadSize, messageChannel)
		messageChannel <- StartEvent(constants.StageUpload, fmt.Sprintf("Uploading to %s", u.StorageService.Name))
	}

	// Set up an SFTP client. This may fail under certain conditions.
	sftpClient, err := NewSFTPClient(u.StorageService, progress)
	if err != nil {
		if sftpClient != nil {
			sftpClient.Close()
		}
		u.Errors[u.StorageService.Name] = fmt.Sprintf("Error initializing SFTP client for %s : %s", u.StorageService.Name, err.Error())
		return false
	}

	// If we got this far, we have a connection.
	// Make sure to clean it up.
	defer sftpClient.Close()

	// Okay, now we're set up to start uploading.
	// The allSucceeded flag will let us know later if there
	// were any errors.
	allSucceeded := true

	for _, fileOrDirectoryPath := range u.SourceFiles {
		// Now, do the upload. Note that we may be uploading
		// a single file or an entire directory tree.
		dest := filepath.Join(u.StorageService.Bucket, filepath.Base(fileOrDirectoryPath))
		err = sftpClient.Upload(fileOrDirectoryPath, filepath.ToSlash(dest))
		if err != nil {
			key := fmt.Sprintf("%s - %s", u.StorageService.Name, fileOrDirectoryPath)
			u.Errors[key] = fmt.Sprintf("Error copying %s to S3: %s", fileOrDirectoryPath, err.Error())
			allSucceeded = false
		} else {
			Dart.Log.Infof("Finished SFTP upload of file/directory %s to %s", fileOrDirectoryPath, u.StorageService.Name)
		}
	}
	return allSucceeded
}

// useSSL returns a boolean describing whether we should use secure
// connections for S3 uploads. This returns true unless we're talking
// to localhost (which we do in unit tests).
func (u *UploadOperation) useSSL() bool {
	useSSL := !strings.HasPrefix(u.StorageService.Host, "localhost") && !strings.HasPrefix(u.StorageService.Host, "127.0.0.1")
	Dart.Log.Infof("Use SSL for upload = %t", useSSL)
	return useSSL
}
