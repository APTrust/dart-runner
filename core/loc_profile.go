package core

// LOCOrderedProfile represents a Library of Congress BagIt profile
// containing an ordered list of tags. In this format, each entry
// in the list is a map containing a single key and value, where key
// is the tag name and value is the LOC tag definition.
type LOCOrderedProfile struct {
	Tags []map[string]LOCTagDef `json:"ordered"`
}

// LOCTagDef represents a Library of Congress tag definition. These
// may appear in both ordered and unordered LOC profiles. Unordered
// LOC profiles are simply a map in format map[string]LOCTagDef
type LOCTagDef struct {
	Required      bool     `json:"fieldRequired,omitempty"`
	DefaultValue  string   `json:"defaultValue,omitempty"`
	Values        []string `json:"valueList,omitempty"`
	RequiredValue string   `json:"requiredValue,omitempty"`
}
