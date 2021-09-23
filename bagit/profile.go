package bagit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
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
	AcceptBagItVersion   []string          `json:"acceptBagItVersion"`
	AcceptSerialization  []string          `json:"acceptSerialization"`
	AllowFetchTxt        bool              `json:"allowFetchTxt"`
	BagItProfileInfo     ProfileInfo       `json:"bagItProfileInfo"`
	Description          string            `json:"description"`
	Errors               map[string]string `json:"-"`
	ManifestsAllowed     []string          `json:"manifestsAllowed"`
	ManifestsRequired    []string          `json:"manifestsRequired"`
	Name                 string            `json:"name"`
	Serialization        string            `json:"serialization"`
	TagFilesAllowed      []string          `json:"tagFilesAllowed"`
	TagManifestsAllowed  []string          `json:"tagManifestsAllowed"`
	TagManifestsRequired []string          `json:"tagManifestsRequired"`
	Tags                 []*TagDefinition  `json:"tags"`
}

func NewProfile() *Profile {
	profile := &Profile{
		AllowFetchTxt:        false,
		BagItProfileInfo:     ProfileInfo{},
		Errors:               make(map[string]string),
		ManifestsAllowed:     make([]string, 0),
		ManifestsRequired:    make([]string, 0),
		Serialization:        constants.SerializationOptional,
		TagFilesAllowed:      make([]string, 0),
		TagManifestsAllowed:  make([]string, 0),
		TagManifestsRequired: make([]string, 0),
		Tags:                 make([]*TagDefinition, 0),
	}
	copy(profile.AcceptBagItVersion, constants.AcceptBagItVersion)
	copy(profile.AcceptSerialization, constants.AcceptSerialization)
	return profile
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

func (p *Profile) HasTagFile(name string) bool {
	tagDef, _ := p.FirstMatchingTag("TagFile", name)
	return tagDef != nil
}

// IsValid returns true if this profile is valid. This is not to be
// confused with bag validation. We're just making sure the profile itself
// is complete and makes sense.
func (p *Profile) IsValid() bool {
	p.Errors = make(map[string]string)
	if util.IsEmptyStringList(p.AcceptBagItVersion) {
		p.Errors["BagItProfile.AcceptBagItVersion"] = "Profile must accept at least one BagIt version."
	}
	if util.IsEmptyStringList(p.ManifestsAllowed) {
		p.Errors["BagItProfile.ManifestsAllowed"] = "Profile must allow at least one manifest algorithm."
	}
	if !p.HasTagFile("bagit.txt") {
		p.Errors["BagItProfile.BagIt"] = "Profile lacks requirements for bagit.txt tag file."
	}
	if !p.HasTagFile("bag-info.txt") {
		p.Errors["BagItProfile.BagInfo"] = "Profile lacks requirements for bag-info.txt tag file."
	}
	if !util.StringListContains(constants.SerializationOptions, p.Serialization) {
		p.Errors["BagItProfile.Serialization"] = fmt.Sprintf("Serialization must be one of: %s.", strings.Join(constants.SerializationOptions, ","))
	}
	if p.Serialization == constants.SerializationOptional || p.Serialization == constants.SerializationRequired {
		if util.IsEmptyStringList(p.AcceptSerialization) {
			p.Errors["BagItProfile.AcceptSerialization"] = "When serialization is allowed, you must specify at least one serialization format."
		}
	}
	return len(p.Errors) == 0
}
