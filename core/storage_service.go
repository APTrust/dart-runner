package core

import (
	"fmt"
	"os"
	"strings"
)

type StorageService struct {
	AllowsDownload bool              `json:"allowsDownload"`
	AllowsUpload   bool              `json:"allowsUpload"`
	Bucket         string            `json:"bucket"`
	Description    string            `json:"description"`
	Errors         map[string]string `json:"-"`
	Host           string            `json:"host"`
	Login          string            `json:"login"`
	LoginExtra     string            `json:"loginExtra"`
	Name           string            `json:"name"`
	Password       string            `json:"password"`
	Port           int               `json:"port"`
	Protocol       string            `json:"protocol"`
}

func NewStorageService() *StorageService {
	return &StorageService{
		Errors: make(map[string]string),
	}
}

// URL returns the URL to which the file will be uploaded.
func (ss *StorageService) URL(filename string) string {
	return fmt.Sprintf("%s://%s/%s/%s", ss.Protocol, ss.HostAndPort(), ss.Bucket, filename)
}

// HostAndPort returns the host and port for connecting to a remote service.
// Use this when creating a connection to a remote S3 service.
func (ss *StorageService) HostAndPort() string {
	port := ""
	if ss.Port > 0 {
		port = fmt.Sprintf(":%d", ss.Port)
	}
	return fmt.Sprintf("%s%s", ss.Host, port)
}

func (ss *StorageService) Validate() bool {
	ss.Errors = make(map[string]string)
	if strings.TrimSpace(ss.Protocol) == "" {
		ss.Errors["StorageService.Protocol"] = "StorageService requires a protocol (s3, sftp, etc)."
	}
	if strings.TrimSpace(ss.Host) == "" {
		ss.Errors["StorageService.Host"] = "StorageService requires a hostname or IP address."
	}
	if strings.TrimSpace(ss.Bucket) == "" {
		ss.Errors["StorageService.Bucket"] = "StorageService requires a bucket or folder name."
	}
	if strings.TrimSpace(ss.Login) == "" {
		ss.Errors["StorageService.Login"] = "StorageService requires a login name or access key id."
	}
	if strings.TrimSpace(ss.Password) == "" {
		ss.Errors["StorageService.Password"] = "StorageService requires a password or secret access key."
	}
	return len(ss.Errors) == 0
}

// GetLogin returns the login name or AccessKeyID to connect to this
// storage service. Per the DART docts, if the login begins with "env:",
// we fetch it from the environment. For example, "env:MY_SS_LOGIN"
// causes us to fetch the env var "MY_SS_LOGIN". This allows us to
// copy Workflow info across the wire without exposing sensitive credentials.
//
// If the login does not begin with "env:", this returns it verbatim.
func (ss *StorageService) GetLogin() string {
	if strings.HasPrefix(ss.Login, "env:") {
		return ss.getEnv(ss.Login)
	}
	return ss.Login
}

// GetPassword returns this storage service's password from the
// StorageService record or from the environment as necessary. See the
// documentation for StorageService.GetLogin() for more info.
func (ss *StorageService) GetPassword() string {
	if strings.HasPrefix(ss.Password, "env:") {
		return ss.getEnv(ss.Password)
	}
	return ss.Password
}

// getEnv returns the value of an environment variable, minus the
// "ENV:" prefix.
func (ss *StorageService) getEnv(varname string) string {
	parts := strings.SplitN(varname, ":", 2)
	return os.Getenv(parts[1])
}
