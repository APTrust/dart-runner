package core

type ExportQuestion struct {
	ID      string `json:"id"`
	Prompt  string `json:"prompt"`
	ObjType string `json:"objType"`
	ObjID   string `json:"objId"`
	Field   string `json:"field"`
}
