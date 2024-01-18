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
	Errors             map[string]string  `json:"errors"`
	PayloadSize        int64              `json:"payloadSize"`
	Result             *OperationResult   `json:"result"`
	SourceFiles        []string           `json:"sourceFiles"`
	StorageService     *StorageService    `json:"storageService"`
	MessageChannel     chan *EventMessage `json:"-"`
	ExpandedSourceList []string           `json:"expandedSourceList"`
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
	for key, value := range u.Errors {
		Dart.Log.Errorf("%s: %s", key, value)
	}
	return len(u.Errors) == 0
}

func (u *UploadOperation) CalculatePayloadSize() error {
	u.expandSourceFileList()
	u.PayloadSize = 0
	for _, fileOrDir := range u.ExpandedSourceList {
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
	u.expandSourceFileList()
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
	for _, sourceFile := range u.ExpandedSourceList {
		s3Key := filepath.Base(sourceFile)
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
			allSucceeded = false
		} else {
			u.Result.RemoteChecksum = uploadInfo.ETag
			Dart.Log.Infof("Finished S3 upload of file %s; got e-tag %s", sourceFile, uploadInfo.ETag)
		}
	}
	return allSucceeded
}

// upload an item to SFTP server. If messageChannel is not nil,
// the uploader will send progress updates through it. Otherwise,
// no progress updates.
func (u *UploadOperation) sendToSFTP(messageChannel chan *EventMessage) bool {
	allSucceeded := true
	for _, file := range u.ExpandedSourceList {
		var progress *StreamProgress
		if messageChannel != nil {
			progress = NewStreamProgress(u.PayloadSize, messageChannel)
			messageChannel <- StartEvent(constants.StageUpload, fmt.Sprintf("Uploading to %s", u.StorageService.Name))
		}
		SFTPUpload(u.StorageService, file, progress)
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

func (u *UploadOperation) expandSourceFileList() error {

	// START HERE

	// TODO: Set a flag describing whether this has already been done,
	// so we don't do it over and over. Also, consider using util.RecursiveFileList,
	// since we're going to have to stat everything anyway.
	//
	// Also note that when uploading loose files to S3, they all go into
	// the top-level dir. They should instead mirror the local struction.
	//
	// SFTP uploads seem to create the necessary directories, but
	// they're created as files, and then
	// all of the files go into the top-level directory. Fix that.
	//
	// Also, the progress bar for SFTP multifile uploads is schizo.
	// It's not calculating progress against the total upload size.

	u.ExpandedSourceList = make([]string, 0)
	for _, filePath := range u.SourceFiles {
		if util.IsDirectory(filePath) {
			filesInSource, err := listDirRecursive(filePath)
			if err != nil {
				return err
			}
			u.ExpandedSourceList = append(u.ExpandedSourceList, filesInSource...)
		} else {
			u.ExpandedSourceList = append(u.ExpandedSourceList, filePath)
		}
	}
	return nil
}

func listDirRecursive(dir string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(dir, func(filePath string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() || f.IsDir() {
			files = append(files, filePath)
		}
		return nil
	})
	return files, err
}
