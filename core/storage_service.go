package core

type StorageService struct {
	AllowsDownload bool   `json:"allowsDownload"`
	AllowsUpload   bool   `json:"allowsUpload"`
	Bucket         string `json:"bucket"`
	Description    string `json:"description"`
	Host           string `json:"host"`
	Login          string `json:"port"`
	LoginExtra     string `json:"loginExtra"`
	Name           string `json:"name"`
	Password       string `json:"password"`
	Port           int    `json:"port"`
	Protocol       string `json:"protocol"`
}
