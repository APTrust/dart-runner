package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/machinebox/progress"
	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHConnect returns an ssh connection or an error. If the StorageService
// specifies LoginExtra, this will treat that as the path to the SSH key;
// otherwise, it will use ss.Password as the password. In either case, it
// uses ss.Login as the user name.
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
	config := &goph.Config{
		Auth:     auth,
		User:     ss.Login,
		Addr:     ss.Host,
		Port:     uint(ss.Port),
		Timeout:  5 * time.Second,
		Callback: ssh.InsecureIgnoreHostKey(), // TODO: Change this for production
	}
	return goph.NewConn(config)
}

// SFTPConnect returns an sftp connection or an error. If the StorageService
// specifies LoginExtra, this will treat that as the path to the SSH key;
// otherwise, it will use ss.Password as the password. In either case, it
// uses ss.Login as the user name.
func SFTPConnect(ss *StorageService) (*sftp.Client, error) {
	conn, err := SSHConnect(ss)
	if err != nil {
		Dart.Log.Errorf("Connection to SFTP service %s at %s failed: %v", ss.Name, ss.HostAndPort(), err)
		return nil, err
	}
	return conn.NewSftp()
}

// SFTPUpload uploads a file to the SFTP server described in the
// StorageService param. localPath is the path to the local file
// that you want to upload to the remote server. The uploadProgress
// param should be nil except when running a job from the UI.
// For jobs launched from the UI, the uploadProgress object will
// pass progress info back to the front end. Be sure to set
// uploadProgress.Total to the size of the file and make sure
// the MessageChannel is initialized.
func SFTPUpload(ss *StorageService, localPath string, uploadProgress *StreamProgress) (int64, error) {
	localFile, err := os.Open(localPath)
	if err != nil {
		Dart.Log.Errorf("SFTPUpload can't open file %s: %v", localPath, err)
		return 0, err
	}
	client, err := SFTPConnect(ss)
	if err != nil {
		return 0, err
	}

	remoteFileName := path.Join(ss.Bucket, path.Base(localPath))

	err = client.MkdirAll(ss.Bucket)
	if err != nil {
		message := fmt.Sprintf("SFTPUpload cannot create remote parent directories for '%s': %v", ss.Bucket, err)
		Dart.Log.Error(message)
		return 0, fmt.Errorf(message)
	}

	remoteFile, err := client.Create(remoteFileName)
	if err != nil {
		message := fmt.Sprintf("SFTPUpload cannot create remote file '%s': %v", remoteFileName, err)
		Dart.Log.Error(message)
		return 0, fmt.Errorf(message)
	}
	defer remoteFile.Close()

	Dart.Log.Infof("SFTPUpload: uploading %s to %s", localPath, remoteFileName)

	bytesWritten := int64(0)
	if uploadProgress != nil {
		progressWriter := progress.NewWriter(remoteFile)
		go func() {
			progressChan := progress.NewTicker(context.Background(), progressWriter, uploadProgress.Total, 1*time.Second)
			for p := range progressChan {
				// TODO: This causes a panic by trying to write to a
				// closed channel. The JobRunController closes the channel
				// when the upload is complete, but the progressChan runs
				// on a ticker that has no idea the messageChannel is closed.
				//
				// UPDATE 2023-11-06: Is this still an issue? I haven't seen
				// it happen in months.
				uploadProgress.SetTotalBytesCompleted(p.N())
			}
		}()
		bytesWritten, err = io.Copy(progressWriter, localFile)
	} else {
		bytesWritten, err = io.Copy(remoteFile, localFile)
	}
	Dart.Log.Infof("SFTPUpload: uploaded %d bytes to %s", bytesWritten, remoteFileName)
	return bytesWritten, err
}
