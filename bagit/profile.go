package bagit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Profile represents a DART-type Profile, as described at
// https://aptrust.github.io/dart/Profile.html. This format differs
// slightly from the profiles at
// https://github.com/bagit-profiles/bagit-profiles-specification. The
// DART specification is richer and can describe requirements that the
// other profile format cannot. DART can convert between the two formats
// as described in https://aptrust.github.io/dart-docs/users/bagit/importing/
// and https://aptrust.github.io/dart-docs/users/bagit/exporting/.
type Profile struct {
	AcceptBagItVersion   []string         `json:"acceptBagItVersion"`
	AcceptSerialization  []string         `json:"acceptSerialization"`
	AllowFetchTxt        bool             `json:"allowFetchTxt"`
	BagItProfileInfo     ProfileInfo      `json:"bagItProfileInfo"`
	Description          string           `json:"description"`
	ManifestsAllowed     []string         `json:"manifestsAllowed"`
	ManifestsRequired    []string         `json:"manifestsRequired"`
	Name                 string           `json:"name"`
	Serialization        string           `json:"serialization"`
	TagFilesAllowed      []string         `json:"tagFilesAllowed"`
	TagManifestsAllowed  []string         `json:"tagManifestsAllowed"`
	TagManifestsRequired []string         `json:"tagManifestsRequired"`
	Tags                 []*TagDefinition `json:"tags"`
}

// ProfileLoad loads a BagIt Profile from the specified file.
func ProfileLoad(filename string) (*Profile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return ProfileFromJSON(string(data))
}

// ProfileFromJSON converts a JSON representation of a BagIt Profile
// to a Profile object.
func ProfileFromJSON(jsonData string) (*Profile, error) {
	p := &Profile{}
	err := json.Unmarshal([]byte(jsonData), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

//ToJSON returns a JSON representation of this object.
func (p *Profile) ToJSON() (string, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetTagDef returns the TagDefinition for the specified tag file
// and tag name.
// Note: BagIt spec section 2.2.2 says tag names are case-insensitive.
// https://tools.ietf.org/html/rfc8493#section-2.2.2
func (p *Profile) GetTagDef(tagFile, tagName string) *TagDefinition {
	for _, tagDef := range p.Tags {
		// Try exact match first
		if tagDef.TagFile == tagFile && tagDef.TagName == tagName {
			return tagDef
		}
		// Try case-insensitve match
		if tagDef.TagFile == tagFile && strings.EqualFold(tagDef.TagName, tagName) {
			return tagDef
		}
	}
	return nil
}

func (p *Profile) FindMatchingTags(property, value string) ([]*TagDefinition, error) {
	matches := make([]*TagDefinition, 0)
	for _, tagDef := range p.Tags {
		var match *TagDefinition
		switch property {
		case "DefaultValue":
			if tagDef.DefaultValue == value {
				match = tagDef
			}
		case "ID":
			if tagDef.ID == value {
				match = tagDef
			}
		case "TagFile":
			if tagDef.TagFile == value {
				match = tagDef
			}
		case "TagName":
			if tagDef.TagName == value {
				match = tagDef
			}
		case "UserValue":
			if tagDef.UserValue == value {
				match = tagDef
			}
		default:
			return nil, fmt.Errorf("search property not supported")
		}
		if match != nil {
			matches = append(matches, match)
		}
	}
	return matches, nil
}

func (p *Profile) FirstMatchingTag(property, value string) (*TagDefinition, error) {
	for _, tagDef := range p.Tags {
		switch property {
		case "DefaultValue":
			if tagDef.DefaultValue == value {
				return tagDef, nil
			}
		case "ID":
			if tagDef.ID == value {
				return tagDef, nil
			}
		case "TagFile":
			if tagDef.TagFile == value {
				return tagDef, nil
			}
		case "TagName":
			if tagDef.TagName == value {
				return tagDef, nil
			}
		case "UserValue":
			if tagDef.UserValue == value {
				return tagDef, nil
			}
		default:
			return nil, fmt.Errorf("search property not supported")
		}
	}
	return nil, nil
}
