package bagit

import (
	"github.com/APTrust/dart-runner/util"
)

// TagDefinition describes a tag in a BagItProfile, whether it's
// required, what values are allowed, etc.
type TagDefinition struct {
	DefaultValue string   `json:"defaultValue"`
	Help         string   `json:"help"`
	ID           string   `json:"id"`
	Required     bool     `json:"required"`
	TagFile      string   `json:"tagFile"`
	TagName      string   `json:"tagName"`
	UserValue    string   `json:"userValue"`
	Values       []string `json:"values"`
}

// IsLegalValue returns true if val is a legal value for this tag definition.
// If TagDefinition.Values is empty, all values are legal.
func (t *TagDefinition) IsLegalValue(val string) bool {
	if t.Values == nil || len(t.Values) == 0 {
		return true
	}
	return util.StringListContains(t.Values, val)
}
