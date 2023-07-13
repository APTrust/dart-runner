package core

type LOCOrderedProfile struct {
	Tags []map[string]LOCTagDef `json:"ordered"`
}

type LOCTagDef struct {
	Required      bool     `json:"fieldRequired,omitempty"`
	DefaultValue  string   `json:"defaultValue,omitempty"`
	Values        []string `json:"valueList,omitempty"`
	RequiredValue string   `json:"requiredValue,omitempty"`
}
