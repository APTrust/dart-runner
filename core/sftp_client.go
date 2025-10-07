package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPClient struct {
	client *sftp.Client
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

// NewSFTPClient creates a new SFTP client connection
func NewSFTPClient(ss *StorageService) (*SFTPClient, error) {
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

	return &SFTPClient{client: client}, nil
}

// Close closes the SFTP client connection
func (sc *SFTPClient) Close() error {
	return sc.client.Close()
}

// Upload uploads a file or directory from source to destination on the remote server
func (sc *SFTPClient) Upload(source, destination string) error {
	// Get file info to determine if source is a file or directory
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if info.IsDir() {
		return sc.uploadDirectory(source, destination)
	}
	return sc.uploadFile(source, destination, info)
}

// uploadFile uploads a single file to the remote server
func (sc *SFTPClient) uploadFile(localPath, remotePath string, info os.FileInfo) error {
	// Open local file
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

	fmt.Printf("Uploaded file: %s -> %s\n", localPath, remotePath)
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
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Convert to forward slashes for remote path
		remoteDest := filepath.ToSlash(filepath.Join(remotePath, relPath))

		if info.IsDir() {
			// Create directory on remote server
			err = sc.client.MkdirAll(remoteDest)
			if err != nil {
				return fmt.Errorf("failed to create remote directory: %w", err)
			}

			// Preserve directory permissions
			err = sc.client.Chmod(remoteDest, info.Mode())
			if err != nil {
				return fmt.Errorf("failed to set directory permissions: %w", err)
			}

			fmt.Printf("Created directory: %s\n", remoteDest)
		} else {
			// Upload file
			err = sc.uploadFile(path, remoteDest, info)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
