package core

import (
	"context"
	"fmt"
	"os"
	"path"
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
	return &UploadOperation{
		Errors:         make(map[string]string),
		Result:         NewOperationResult("upload", "Uploader - "+constants.AppVersion),
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
	if u.SourceFiles == nil || len(u.SourceFiles) == 0 {
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
	return len(u.Errors) == 0
}

func (u *UploadOperation) CalculatePayloadSize() error {
	u.PayloadSize = 0
	for _, fileOrDir := range u.SourceFiles {
		stat, err := os.Stat(fileOrDir)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			children, err := util.RecursiveFileList(fileOrDir)
			if err != nil {
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

func (u *UploadOperation) DoUploadWithProgress(progress *S3UploadProgress) bool {
	ok := false
	switch u.StorageService.Protocol {
	case constants.ProtocolS3:
		ok = u.sendToS3(progress)
	case constants.ProtocolSFTP:
		ok = u.sendToSFTP() // progress not yet supported for sftp
	default:
		u.Errors["Protocol"] = fmt.Sprintf("Unsupported upload protocol: %s", u.StorageService.Protocol)
	}
	return ok
}

// TODO: Deprecate??
func (u *UploadOperation) DoUpload() bool {
	ok := false
	switch u.StorageService.Protocol {
	case constants.ProtocolS3:
		ok = u.sendToS3(nil)
	case constants.ProtocolSFTP:
		ok = u.sendToSFTP()
	default:
		u.Errors["Protocol"] = fmt.Sprintf("Unsupported upload protocol: %s", u.StorageService.Protocol)
	}
	return ok
}

func (u *UploadOperation) sendToS3(progress *S3UploadProgress) bool {
	accessKeyId := u.StorageService.GetLogin()
	secretKey := u.StorageService.GetPassword()
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretKey, ""),
		Secure: u.useSSL(),
	}
	client, err := minio.New(u.StorageService.HostAndPort(), options)
	if err != nil {
		u.Errors["S3Upload"] = fmt.Sprintf("Error connecting to S3: %s", err.Error())
		return false
	}
	allSucceeded := true
	for _, sourceFile := range u.SourceFiles {
		s3Key := path.Base(sourceFile)
		u.Result.RemoteURL = u.StorageService.URL(s3Key)
		putOptions := minio.PutObjectOptions{}
		if progress != nil {
			putOptions = minio.PutObjectOptions{
				Progress: progress,
			}
		}
		uploadInfo, err := client.FPutObject(
			context.Background(),
			u.StorageService.Bucket,
			s3Key,
			sourceFile,
			putOptions,
		)
		if err != nil {
			u.Errors[s3Key] = fmt.Sprintf("Error copying %s to S3: %s", sourceFile, err.Error())
			allSucceeded = false
		} else {
			u.Result.RemoteChecksum = uploadInfo.ETag
		}
	}
	return allSucceeded
}

// useSSL returns a boolean describing whether we should use secure
// connections for S3 uploads. This returns true unless we're talking
// to localhost (which we do in unit tests).
func (u *UploadOperation) useSSL() bool {
	return !strings.HasPrefix(u.StorageService.Host, "localhost") && !strings.HasPrefix(u.StorageService.Host, "127.0.0.1")
}

func (u *UploadOperation) sendToSFTP() bool {
	u.Errors["SFTPUpload"] = "SFTP upload is not yet supported."
	return false
}
