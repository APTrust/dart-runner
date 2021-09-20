package core

type PackageOperation struct {
	BagItSerialization string            `json:"bagItSerialization"`
	Errors             map[string]string `json:"errors"`
	OutputPath         string            `json:"outputPath"`
	PackageFormat      string            `json:"packageFormat"`
	PackageName        string            `json:"packageName"`
	PayloadSize        int64             `json:"payloadSize"`
	PluginId           string            `json:"pluginId"`
	Result             *OperationResult  `json:"result"`
	SkipFiles          []string          `json:"skipFiles"`
	SourceFiles        []string          `json:"sourceFiles"`
	TrimLeadingPaths   bool              `json:"_trimLeadingPaths"`
}
