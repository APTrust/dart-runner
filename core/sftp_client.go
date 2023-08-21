package core

import (
	"context"
	"io"
	"os"
	"path"
	"time"

	"github.com/machinebox/progress"
	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
)

func SSHConnect(ss *StorageService) (*goph.Client, error) {
	var auth goph.Auth
	var err error
	if ss.LoginExtra != "" {
		auth, err = goph.Key(ss.LoginExtra, "")
	} else {
		auth = goph.Password(ss.Password)
	}
	if err != nil {
		return nil, err
	}
	return goph.New(ss.Login, ss.HostAndPort(), auth)
}

func SFTPConnect(ss *StorageService) (*sftp.Client, error) {
	conn, err := SSHConnect(ss)
	if err != nil {
		return nil, err
	}
	return conn.NewSftp()
}

func SFTPUpload(ss *StorageService, localPath string, uploadProgress *S3UploadProgress) (int64, error) {
	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, err
	}
	client, err := SFTPConnect(ss)
	if err != nil {
		return 0, err
	}
	remoteFileName := path.Join(ss.Bucket, path.Base(localPath))
	remoteFile, err := client.Create(remoteFileName)
	if err != nil {
		return 0, err
	}
	defer remoteFile.Close()

	bytesWritten := int64(0)
	if uploadProgress != nil {
		progressWriter := progress.NewWriter(remoteFile)
		go func() {
			progressChan := progress.NewTicker(context.Background(), progressWriter, uploadProgress.Total, 1*time.Second)
			for p := range progressChan {
				uploadProgress.SetTotalBytesCompleted(p.N())
			}
		}()
		bytesWritten, err = io.Copy(progressWriter, localFile)
	} else {
		bytesWritten, err = io.Copy(remoteFile, localFile)
	}
	return bytesWritten, err
}
