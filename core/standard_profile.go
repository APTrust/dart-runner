package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StandardProfile represents a BagIt Profile in BagIt Profile Spec version 1.3.0.
// See https://bagit-profiles.github.io/bagit-profiles-specification/
// and https://github.com/bagit-profiles/bagit-profiles-specification
//
// The BagIt Profile spec can describe a limited subset of the properties DART
// BagIt profiles can describe.
type StandardProfile struct {
	AcceptBagItVersion   []string                         `json:"Accept-BagIt-Version"`
	AcceptSerialization  []string                         `json:"Accept-Serialization"`
	AllowFetchTxt        bool                             `json:"Allow-Fetch.txt"`
	Serialization        string                           `json:"Serialization"`
	ManifestsAllowed     []string                         `json:"Manifests-Allowed"`
	ManifestsRequired    []string                         `json:"Manifests-Required"`
	TagManifestsAllowed  []string                         `json:"Tag-Manifests-Allowed"`
	TagManifestsRequired []string                         `json:"Tag-Manifests-Required"`
	TagFilesAllowed      []string                         `json:"Tag-Files-Allowed"`
	TagFilesRequired     []string                         `json:"Tag-Files-Required"`
	BagItProfileInfo     StandardProfileInfo              `json:"BagIt-Profile-Info"`
	BagInfo              map[string]StandardProfileTagDef `json:"Bag-Info"`
}

// StandardProfileInfo is structurally identical to BagItProfileInfo, but it
// serializes to and from JSON differently.
type StandardProfileInfo struct {
	BagItProfileIdentifier string `json:"BagIt-Profile-Identifier"`
	BagItProfileVersion    string `json:"BagIt-Profile-Version"`
	ContactEmail           string `json:"Contact-Email"`
	ContactName            string `json:"Contact-Name"`
	ExternalDescription    string `json:"External-Description"`
	SourceOrganization     string `json:"Source-Organization"`
	Version                string `json:"Version"`
}

// StandardProfileTagDef represents a tag definition in
// BagIt Profile Spec version 1.3.0.
type StandardProfileTagDef struct {
	Required    bool     `json:"required"`
	Recommended bool     `json:"recommended"`
	Values      []string `json:"values"`
	Description string   `json:"description"`
}

// NewStandardProfile creates a new StandardProfile object with all
// internal structs, slices, and maps allocated and initialized to empty values.
func NewStandardProfile() *StandardProfile {
	return &StandardProfile{
		AcceptBagItVersion:   make([]string, 0),
		AcceptSerialization:  make([]string, 0),
		ManifestsAllowed:     make([]string, 0),
		ManifestsRequired:    make([]string, 0),
		TagManifestsAllowed:  make([]string, 0),
		TagManifestsRequired: make([]string, 0),
		TagFilesAllowed:      make([]string, 0),
		TagFilesRequired:     make([]string, 0),
		BagItProfileInfo:     StandardProfileInfo{},
		BagInfo:              make(map[string]StandardProfileTagDef),
	}
}

// StandardProfileFromJson converts the JSON in jsonBytes into a StandardProfile object.
func StandardProfileFromJson(jsonBytes []byte) (*StandardProfile, error) {
	p := NewStandardProfile()
	err := json.Unmarshal(jsonBytes, p)
	return p, err
}

// ToDartProfile converts a StandardProfile to a DART BagItProfile object.
func (sp *StandardProfile) ToDartProfile() *BagItProfile {
	p := NewBagItProfile()
	p.AllowFetchTxt = sp.AllowFetchTxt
	p.BagItProfileInfo = ProfileInfo{
		BagItProfileIdentifier: sp.BagItProfileInfo.BagItProfileIdentifier,
		BagItProfileVersion:    sp.BagItProfileInfo.BagItProfileVersion,
		ContactEmail:           sp.BagItProfileInfo.ContactEmail,
		ContactName:            sp.BagItProfileInfo.ContactName,
		ExternalDescription:    sp.BagItProfileInfo.ExternalDescription,
		SourceOrganization:     sp.BagItProfileInfo.SourceOrganization,
		Version:                sp.BagItProfileInfo.Version,
	}
	p.Description = sp.BagItProfileInfo.ExternalDescription
	p.ID = uuid.NewString()
	p.IsBuiltIn = false
	p.Serialization = sp.Serialization

	p.AcceptBagItVersion = make([]string, len(sp.AcceptBagItVersion))
	p.AcceptSerialization = make([]string, len(sp.AcceptSerialization))
	p.ManifestsAllowed = make([]string, len(sp.ManifestsAllowed))
	p.ManifestsRequired = make([]string, len(sp.ManifestsRequired))
	p.TagFilesAllowed = make([]string, len(sp.TagFilesAllowed))
	p.TagFilesRequired = make([]string, len(sp.TagFilesRequired))
	p.TagManifestsAllowed = make([]string, len(sp.TagManifestsAllowed))
	p.TagManifestsRequired = make([]string, len(sp.TagManifestsRequired))

	copy(p.AcceptBagItVersion, sp.AcceptBagItVersion)
	copy(p.AcceptSerialization, sp.AcceptSerialization)
	copy(p.ManifestsAllowed, sp.ManifestsAllowed)
	copy(p.ManifestsRequired, sp.ManifestsRequired)
	copy(p.TagFilesAllowed, sp.TagFilesAllowed)
	copy(p.TagFilesRequired, sp.TagFilesRequired)
	copy(p.TagManifestsAllowed, sp.TagManifestsAllowed)
	copy(p.TagManifestsRequired, sp.TagManifestsRequired)

	if sp.BagItProfileInfo.SourceOrganization != "" && sp.BagItProfileInfo.Version != "" {
		p.Name = fmt.Sprintf("%s (version %s)", sp.BagItProfileInfo.SourceOrganization, sp.BagItProfileInfo.Version)
	} else {
		p.Name = fmt.Sprintf("Imported Profile - %s", time.Now().Format(time.Stamp))
	}

	// Standard profile can define tags only for bag-info.txt
	tagDefs := make([]*TagDefinition, 0)
	for name, tag := range sp.BagInfo {
		help := tag.Description
		if tag.Recommended {
			help = fmt.Sprintf("(Recommended) %s", tag.Description)
		}
		tagDef := &TagDefinition{
			TagFile:  "bag-info.txt",
			TagName:  name,
			Required: tag.Required,
			Values:   tag.Values,
			Help:     help,
			EmptyOK:  !tag.Required,
		}
		tagDefs = append(tagDefs, tagDef)
	}
	p.Tags = tagDefs

	return p
}

// ToJSON returns the StandardProfile object in pretty-printed JSON format.
func (sp *StandardProfile) ToJSON() (string, error) {
	data, err := json.MarshalIndent(sp, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), err
}
