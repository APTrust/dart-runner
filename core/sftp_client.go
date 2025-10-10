package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPClient struct {
	client             *sftp.Client
	totalBytesToUpload int64
	bytesUploaded      int64
	filesUploaded      int64
	uploadProgress     *StreamProgress
}

// GetSFTPAuthMethod returns an authentication method to be used
// when connecting to the remote server. If ss.LoginExtra is not
// empty, this returns a ssh.PublicKey AuthMethod using the key
// found in the file pointed to by ss.LoginExtra. Otherwise, it
// returns an ssh.Password() AuthMethod using ss.Password.
func GetSFTPAuthMethod(ss *StorageService) (ssh.AuthMethod, error) {
	// ss.LoginExtra will be the path to the SSH key file
	// needed to authenticate this connection.
	if ss.LoginExtra != "" {
		key, err := os.ReadFile(ss.LoginExtra)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}

		// Parse private key
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		return ssh.PublicKeys(signer), nil
	}
	// If ss.LoginExtra is empty, use password authentication
	return ssh.Password(ss.Password), nil
}

// GetUploadPayloadSize returns the number of bytes to be uploaded
// from pathToDir and all of its subdirectories. This counts bytes for
// regular files only, excluding directories, symlinks, pipes, devices
// and anything else that cannot be copied as a simple byte stream
// across an SFTP connection.
func GetUploadPayloadSize(pathToDir string) (int64, error) {
	byteCount := int64(0)
	err := filepath.Walk(pathToDir, func(filePath string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			byteCount += f.Size()
		}
		return nil
	})
	return byteCount, err
}

// NewSFTPClient creates a new SFTP client connection that
// will connect to the specified StorageService.
// If param uploadProgress is not nil, this will
// update the progress bar as each file is uploaded. Param
// uploadProgress should be nil unless we're running in DART 3 GUI
// mode.
func NewSFTPClient(ss *StorageService, uploadProgress *StreamProgress) (*SFTPClient, error) {
	authMethod, err := GetSFTPAuthMethod(ss)
	if err != nil {
		return nil, err
	}
	// Configure SSH client
	config := &ssh.ClientConfig{
		User: ss.Login,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use proper host key checking in production
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", ss.Host, ss.Port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	// Create SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	sftpClient := &SFTPClient{
		client:             client,
		totalBytesToUpload: int64(0),
		bytesUploaded:      int64(0),
		uploadProgress:     uploadProgress,
	}

	return sftpClient, nil
}

// Close closes the SFTP client connection
func (sc *SFTPClient) Close() error {
	return sc.client.Close()
}

// Upload uploads a file or directory from source to destination
// on the remote server.
func (sc *SFTPClient) Upload(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	err = sc.ensureDestinationRootExists(destination)
	if err != nil {
		return err
	}

	sc.totalBytesToUpload = int64(0)
	sc.bytesUploaded = int64(0)
	if info.IsDir() {
		sc.totalBytesToUpload, err = GetUploadPayloadSize(source)
		if err != nil {
			return err
		}
		return sc.uploadDirectory(source, destination)
	}
	return sc.uploadFile(source, destination, info)
}

// ensureDestinationRootExists ensures that the target root directory exists
// on the remote SFTP server. The StorageService.Bucket attribute contains the
// remote bucket name for S3 storage services, and the remote upload directory
// name for SFTP storage services. If StorageService.Bucket contains a directory
// name that does not exist on the remote server, any attempt to store a file
// in that directory will fail.
//
// SFTP servers typically map user logins to individual user home directories.
// This means most users *should* be able to leave StorageService.BucketName
// empty for SFTP services. If the user happens to have a non-empty bucket name,
// we want to make sure it's there to receive whatever files we send.
//
// Note that uploadDirectory automatically creates directories as necessary,
// but uploadFile does not, so we do that here.
func (sc *SFTPClient) ensureDestinationRootExists(destination string) error {
	remoteParentDir := filepath.Dir(destination)
	if remoteParentDir == "" {
		return nil // nothing to do
	}
	_, err := sc.client.Lstat(remoteParentDir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err = sc.client.MkdirAll(remoteParentDir)
		if err != nil {
			return fmt.Errorf("failed to create remote directory %s: %w", remoteParentDir, err)
		} else {
			Dart.Log.Infof("SFTP client created remote directory %s", remoteParentDir)
		}
	} else if err != nil {
		return fmt.Errorf("error checking whether parent directory %s exists on remote server: %w", remoteParentDir, err)
	}
	return nil
}

// uploadFile uploads a single file to the remote server
func (sc *SFTPClient) uploadFile(localPath, remotePath string, info os.FileInfo) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer srcFile.Close()

	// Create remote file
	dstFile, err := sc.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Preserve file permissions
	err = sc.client.Chmod(remotePath, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Housekeeping
	sc.bytesUploaded += info.Size()
	if sc.uploadProgress != nil {
		sc.uploadProgress.SetTotalBytesCompleted(sc.bytesUploaded)
	}

	Dart.Log.Infof("Uploaded file: %s -> %s", localPath, remotePath)
	return nil
}

// uploadDirectory recursively uploads a directory to the remote server
func (sc *SFTPClient) uploadDirectory(localPath, remotePath string) error {
	// Walk through the local directory
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for local path %s: %w", localPath, err)
		}

		// Convert to forward slashes for remote path
		remoteDest := filepath.ToSlash(filepath.Join(remotePath, relPath))

		if info.IsDir() {
			// Create directory on remote server
			err = sc.client.MkdirAll(remoteDest)
			if err != nil {
				return fmt.Errorf("failed to create remote directory %s: %w", remoteDest, err)
			}

			// Preserve directory permissions
			err = sc.client.Chmod(remoteDest, info.Mode())
			if err != nil {
				return fmt.Errorf("failed to set directory permissions on %s: %w", remoteDest, err)
			}

			Dart.Log.Infof("SFTP client created remote directory: %s\n", remoteDest)
		} else if info.Mode().IsRegular() {
			// Upload file
			err = sc.uploadFile(path, remoteDest, info)
			if err != nil {
				detailedErr := fmt.Errorf("SFTP client: error uploading %s: %w", path, err)
				return detailedErr
			} else {
				sc.filesUploaded += 1
				Dart.Log.Infof("SFTP client: uploaded %s", path)
			}
		} else {
			Dart.Log.Warningf("SFTP client is skipping upload of %s because it's not a regular file", path)
		}

		return nil
	})
}

func (sc *SFTPClient) FilesUploaded() int64 {
	return sc.filesUploaded
}

func (sc *SFTPClient) BytesUploaded() int64 {
	return sc.bytesUploaded
}

func (sc *SFTPClient) PayloadSize() int64 {
	return sc.totalBytesToUpload
}
