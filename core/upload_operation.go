package core

type UploadOperation struct {
	Errors           map[string]string  `json:"errors"`
	PayloadSize      int64              `json:"payloadSize"`
	Results          []*OperationResult `json:"results"`
	SourceFiles      []string           `json:"sourceFiles"`
	StorageServiceID string             `json:"storageServiceId"`
}
