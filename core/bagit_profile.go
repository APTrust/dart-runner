package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

// BagItProfile represents a DART-type BagItProfile, as described at
// https://aptrust.github.io/dart/BagItProfile.html. This format differs
// slightly from the profiles at
// https://github.com/bagit-profiles/bagit-profiles-specification. The
// DART specification is richer and can describe requirements that the
// other profile format cannot. DART can convert between the two formats
// as described in https://aptrust.github.io/dart-docs/users/bagit/importing/
// and https://aptrust.github.io/dart-docs/users/bagit/exporting/.
type BagItProfile struct {
	ID                   string            `json:"id"`
	AcceptBagItVersion   []string          `json:"acceptBagItVersion"`
	AcceptSerialization  []string          `json:"acceptSerialization"`
	AllowFetchTxt        bool              `json:"allowFetchTxt"`
	BagItProfileInfo     ProfileInfo       `json:"bagItProfileInfo"`
	BaseProfileID        string            `json:"baseProfileId"`
	Description          string            `json:"description"`
	Errors               map[string]string `json:"-"`
	IsBuiltIn            bool              `json:"isBuiltIn"`
	ManifestsAllowed     []string          `json:"manifestsAllowed"`
	ManifestsRequired    []string          `json:"manifestsRequired"`
	Name                 string            `json:"name"`
	Serialization        string            `json:"serialization"`
	TagFilesAllowed      []string          `json:"tagFilesAllowed"`
	TagFilesRequired     []string          `json:"tagFilesRequired"`
	TagManifestsAllowed  []string          `json:"tagManifestsAllowed"`
	TagManifestsRequired []string          `json:"tagManifestsRequired"`
	Tags                 []*TagDefinition  `json:"tags"`
	TarDirMustMatchName  bool              `json:"tarDirMustMatchName"`
}

func NewBagItProfile() *BagItProfile {
	profile := &BagItProfile{
		AcceptBagItVersion:   make([]string, len(constants.AcceptBagItVersion)),
		AcceptSerialization:  make([]string, len(constants.AcceptSerialization)),
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

func BagItProfileClone(p *BagItProfile) *BagItProfile {
	profile := &BagItProfile{
		ID:                   uuid.NewString(),
		AcceptBagItVersion:   make([]string, len(p.AcceptBagItVersion)),
		AcceptSerialization:  make([]string, len(p.AcceptSerialization)),
		AllowFetchTxt:        false,
		BaseProfileID:        p.BaseProfileID,
		Description:          p.Description,
		Errors:               make(map[string]string),
		ManifestsAllowed:     make([]string, len(p.ManifestsAllowed)),
		ManifestsRequired:    make([]string, len(p.ManifestsRequired)),
		Name:                 p.Name,
		Serialization:        p.Serialization,
		TagFilesAllowed:      make([]string, len(p.TagFilesAllowed)),
		TagManifestsAllowed:  make([]string, len(p.TagManifestsAllowed)),
		TagManifestsRequired: make([]string, len(p.TagManifestsRequired)),
		Tags:                 make([]*TagDefinition, len(p.Tags)),
	}
	profile.BagItProfileInfo = CopyProfileInfo(p.BagItProfileInfo)
	copy(profile.AcceptBagItVersion, p.AcceptBagItVersion)
	copy(profile.AcceptSerialization, p.AcceptSerialization)
	copy(profile.ManifestsAllowed, p.ManifestsAllowed)
	copy(profile.ManifestsRequired, p.ManifestsRequired)
	copy(profile.TagFilesAllowed, p.TagFilesAllowed)
	copy(profile.TagManifestsAllowed, p.TagManifestsAllowed)
	copy(profile.TagManifestsRequired, p.TagManifestsRequired)
	for i, tag := range p.Tags {
		profile.Tags[i] = tag.Copy() // These are TagDefinition objects
	}
	return profile
}

// BagItProfileLoad loads a BagIt Profile from the specified file.
func BagItProfileLoad(filename string) (*BagItProfile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return BagItProfileFromJSON(string(data))
}

// BagItProfileFromJSON converts a JSON representation of a BagIt Profile
// to a Profile object.
func BagItProfileFromJSON(jsonData string) (*BagItProfile, error) {
	p := &BagItProfile{}
	err := json.Unmarshal([]byte(jsonData), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ToJSON returns a JSON representation of this object.
func (p *BagItProfile) ToJSON() (string, error) {
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
func (p *BagItProfile) GetTagDef(tagFile, tagName string) *TagDefinition {
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

func (p *BagItProfile) FindMatchingTags(property, value string) ([]*TagDefinition, error) {
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

func (p *BagItProfile) FirstMatchingTag(property, value string) (*TagDefinition, error) {
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

func (p *BagItProfile) HasTagFile(name string) bool {
	tagDef, _ := p.FirstMatchingTag("TagFile", name)
	return tagDef != nil
}

// Validate returns true if this profile is valid. This is not to be
// confused with bag validation. We're just making sure the profile itself
// is complete and makes sense.
func (p *BagItProfile) Validate() bool {
	p.Errors = make(map[string]string)
	if !util.LooksLikeUUID(p.ID) {
		p.Errors["ID"] = "Profile ID is missing."
	}
	if strings.TrimSpace(p.Name) == "" {
		p.Errors["Name"] = "Profile requires a name."
	}
	if util.IsEmptyStringList(p.AcceptBagItVersion) {
		p.Errors["AcceptBagItVersion"] = "Profile must accept at least one BagIt version."
	}
	if util.IsEmptyStringList(p.ManifestsAllowed) {
		p.Errors["ManifestsAllowed"] = "Profile must allow at least one manifest algorithm."
	}
	if !p.HasTagFile("bagit.txt") {
		p.Errors["BagIt"] = "Profile lacks requirements for bagit.txt tag file."
	}
	if !p.HasTagFile("bag-info.txt") {
		p.Errors["BagInfo"] = "Profile lacks requirements for bag-info.txt tag file."
	}
	if !util.StringListContains(constants.SerializationOptions, p.Serialization) {
		p.Errors["Serialization"] = fmt.Sprintf("Serialization must be one of: %s.", strings.Join(constants.SerializationOptions, ","))
	}
	if p.Serialization == constants.SerializationOptional || p.Serialization == constants.SerializationRequired {
		if util.IsEmptyStringList(p.AcceptSerialization) {
			p.Errors["AcceptSerialization"] = "When serialization is allowed, you must specify at least one serialization format."
		}
	}
	return len(p.Errors) == 0
}

// TagFileNames returns the names of the tag files for which we
// have actual tag definitions. The bag may require other tag
// files, but we can't produce them if we don't have tag defs.
func (p *BagItProfile) TagFileNames() []string {
	distinct := make(map[string]bool)
	for _, tagDef := range p.Tags {
		distinct[tagDef.TagFile] = true
	}
	names := make([]string, len(distinct))
	i := 0
	for name, _ := range distinct {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

// GetTagFileContents returns the generated contents of the specified
// tag file.
func (p *BagItProfile) GetTagFileContents(tagFileName string) (string, error) {
	tags, err := p.FindMatchingTags("TagFile", tagFileName)
	if err != nil {
		return "", err
	}
	contents := make([]string, len(tags))
	for i, tag := range tags {
		contents[i] = tag.ToFormattedString()
	}
	return strings.Join(contents, "\n") + "\n", nil
}

// SetTagValue sets the value of the specified tag in the specified
// file. It creates the tag if it doesn't already exist in the profile.
// This currently supports only one instance of each tag in each file.
func (p *BagItProfile) SetTagValue(tagFile, tagName, value string) {
	tag := p.GetTagDef(tagFile, tagName)
	if tag == nil {
		tag = &TagDefinition{
			ID:        uuid.New().String(),
			TagFile:   tagFile,
			TagName:   tagName,
			UserValue: value,
		}
		p.Tags = append(p.Tags, tag)
	} else {
		tag.UserValue = value
	}
}

func (p *BagItProfile) GetErrors() map[string]string {
	return p.Errors
}

func (p *BagItProfile) IsDeletable() bool {
	return !p.IsBuiltIn
}

func (p *BagItProfile) ObjID() string {
	return p.ID
}

func (p *BagItProfile) ObjName() string {
	return p.Name
}

func (p *BagItProfile) ObjType() string {
	return constants.TypeBagItProfile
}

func (p *BagItProfile) String() string {
	return fmt.Sprintf("BagItProfile: %s", p.Name)
}

func (p *BagItProfile) ToForm() *Form {
	form := NewForm(constants.TypeBagItProfile, p.ID, p.Errors)
	form.UserCanDelete = p.IsDeletable()

	form.AddField("ID", "ID", p.ID, true)

	bagitVersionField := form.AddMultiValueField("AcceptBagItVersion", "AcceptBagItVersion", p.AcceptBagItVersion, true)
	bagitVersionField.Choices = MakeMultiChoiceList(constants.AcceptBagItVersion, p.AcceptBagItVersion)

	acceptSerializationField := form.AddMultiValueField("AcceptSerialization", "AcceptSerialization", p.AcceptSerialization, true)
	acceptSerializationField.Choices = MakeMultiChoiceList(constants.AcceptSerialization, p.AcceptSerialization)

	form.AddField("AllowFetchTxt", "AllowFetchTxt", strconv.FormatBool(p.AllowFetchTxt), true)
	form.AddField("BaseProfileID", "BaseProfileID", p.BaseProfileID, true)
	form.AddField("Description", "Description", p.Description, true)
	form.AddField("IsBuiltIn", "IsBuiltIn", strconv.FormatBool(p.IsBuiltIn), true)

	manifestsAllowedField := form.AddMultiValueField("ManifestsAllowed", "ManifestsAllowed", p.ManifestsAllowed, true)
	manifestsAllowedField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.ManifestsAllowed)

	manifestsRequiredField := form.AddMultiValueField("ManifestsRequired", "ManifestsRequired", p.ManifestsRequired, true)
	manifestsRequiredField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.ManifestsRequired)

	nameField := form.AddField("Name", "Name", p.Name, true)
	if p.IsBuiltIn {
		nameField.Attrs["readonly"] = "readonly"
	}

	serlializationField := form.AddField("Serialization", "Serialization", p.Serialization, true)
	serlializationField.Choices = MakeChoiceList(constants.SerializationOptions, p.Serialization)

	form.AddMultiValueField("TagFilesAllowed", "TagFilesAllowed", p.TagFilesAllowed, true)
	form.AddMultiValueField("TagFilesRequired", "TagFilesRequired", p.TagFilesRequired, true)

	tagManifestsAllowedField := form.AddMultiValueField("TagManifestsAllowed", "TagManifestsAllowed", p.TagManifestsAllowed, true)
	tagManifestsAllowedField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.TagManifestsAllowed)

	tagManifestsRequiredField := form.AddMultiValueField("TagManifestsRequired", "TagManifestsRequired", p.TagManifestsRequired, true)
	tagManifestsRequiredField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.TagManifestsRequired)

	tarDirMustMatchField := form.AddField("TarDirMustMatchName", "TarDirMustMatchName", strconv.FormatBool(p.TarDirMustMatchName), true)
	tarDirMustMatchField.Choices = YesNoChoices(p.TarDirMustMatchName)

	// BagItProfileInfo
	form.AddField("InfoIdentifier", "Identifier", p.BagItProfileInfo.BagItProfileIdentifier, false)
	form.AddField("InfoContactEmail", "Contact Email", p.BagItProfileInfo.ContactEmail, false)
	form.AddField("InfoContactName", "Contact Name", p.BagItProfileInfo.ContactName, false)
	form.AddField("InfoExternalDescription", "External Description", p.BagItProfileInfo.ExternalDescription, false)
	form.AddField("InfoSourceOrganization", "Source Organization", p.BagItProfileInfo.SourceOrganization, false)
	form.AddField("InfoVersion", "Version", p.BagItProfileInfo.Version, false)

	// Tags

	return form
}
