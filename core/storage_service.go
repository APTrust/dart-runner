package core

import (
	"fmt"
	"strings"
)

type StorageService struct {
	AllowsDownload bool              `json:"allowsDownload"`
	AllowsUpload   bool              `json:"allowsUpload"`
	Bucket         string            `json:"bucket"`
	Description    string            `json:"description"`
	Errors         map[string]string `json:"-"`
	Host           string            `json:"host"`
	Login          string            `json:"port"`
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
	port := ""
	if ss.Port > 0 {
		port = fmt.Sprintf(":%d", ss.Port)
	}
	return fmt.Sprintf("%s://%s%s/%s/%s", ss.Protocol, ss.Host, port, ss.Bucket, filename)
}

func (ss *StorageService) Validate() bool {
	ss.Errors = make(map[string]string)
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
		ss.Errors["StorageService.Password"] = "StorageService requires password or secret access key."
	}
	return len(ss.Errors) == 0
}
