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
		ID:                   uuid.NewString(),
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
	profile.initBagitTxt()
	profile.initBagInfoTxt()
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
		Name:                 fmt.Sprintf("Copy of %s", p.Name),
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

// EnsureMinimumRequirements creates the minimum required attributes
// for this profile to pass validation, if those attributes are not
// already present. This function is required to flesh out some barebones
// profiles that user may import from elsewhere. For example, traditional
// Library of Congress profiles do not include any tag info for the
// bagit.txt file, nor do they include info about required or allowed
// manifest algorithms.
func (p *BagItProfile) EnsureMinimumRequirements() {
	if p.AcceptBagItVersion == nil || len(p.AcceptBagItVersion) < 1 {
		copy(p.AcceptBagItVersion, constants.AcceptBagItVersion)
	}
	if p.ManifestsAllowed == nil || len(p.ManifestsAllowed) < 1 {
		p.ManifestsAllowed = make([]string, len(constants.PreferredAlgsInOrder))
		copy(p.ManifestsAllowed, constants.PreferredAlgsInOrder)
	}
	if p.TagManifestsAllowed == nil || len(p.TagManifestsAllowed) < 1 {
		p.TagManifestsAllowed = make([]string, len(constants.PreferredAlgsInOrder))
		copy(p.TagManifestsAllowed, constants.PreferredAlgsInOrder)
	}
	if p.AcceptSerialization == nil || len(p.AcceptSerialization) < 1 {
		p.AcceptSerialization = make([]string, len(constants.AcceptSerialization))
		copy(p.AcceptSerialization, constants.AcceptSerialization)
	}
	if p.Serialization == "" {
		p.Serialization = constants.SerializationOptional
	}
	p.initBagitTxt()
	p.initBagInfoTxt()
}

func (p *BagItProfile) initBagitTxt() {
	// BagIt spec says these two tags in bagit.txt file
	// are always required.
	if p.GetTagDef("bagit.txt", "Bagit-Version") == nil {
		var version = &TagDefinition{
			ID:           uuid.NewString(),
			TagFile:      "bagit.txt",
			TagName:      "BagIt-Version",
			Required:     true,
			DefaultValue: "1.0",
			Help:         "Which version of the BagIt specification describes this bag's format?",
		}
		copy(version.Values, constants.AcceptBagItVersion)
		p.Tags = append(p.Tags, version)
	}
	if p.GetTagDef("bagit.txt", "Tag-File-Character-Encoding") == nil {
		var encoding = &TagDefinition{
			ID:           uuid.NewString(),
			TagFile:      "bagit.txt",
			TagName:      "Tag-File-Character-Encoding",
			Required:     true,
			DefaultValue: "UTF-8",
			Help:         "How are this bag's plain-text tag files encoded? (Hint: usually UTF-8)",
		}
		p.Tags = append(p.Tags, encoding)
	}
}

func (p *BagItProfile) initBagInfoTxt() {
	tags := []string{
		"Bag-Count",
		"Bag-Group-Identifier",
		"Bag-Size",
		"Bagging-Date",
		"Bagging-Software",
		"Contact-Email",
		"Contact-Name",
		"Contact-Phone",
		"External-Description",
		"External-Identifier",
		"Internal-Sender-Description",
		"Internal-Sender-Identifier",
		"Organization-Address",
		"Payload-Oxum",
		"Source-Organization",
	}
	for _, tagName := range tags {
		if p.GetTagDef("bag-info.txt", tagName) == nil {
			tag := &TagDefinition{
				ID:       uuid.NewString(),
				TagFile:  "bag-info.txt",
				TagName:  tagName,
				Required: false,
			}
			p.Tags = append(p.Tags, tag)
		}
	}

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

func (p *BagItProfile) GetTagByFullyQualifiedName(fullyQualifiedName string) *TagDefinition {
	parts := strings.SplitN(fullyQualifiedName, "/", 2)
	if len(parts) == 2 {
		return p.GetTagDef(parts[0], parts[1])
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
// The list will include user-added tag files.
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

// FlagUserAddedTagFiles sets IsUserAddedFile to true on tag definitions
// where the user added the tag file to the profile. The user can do this
// on a per-job basis through the jobs/metadata UI page and through the
// BagIt Profile editor.
func (p *BagItProfile) FlagUserAddedTagFiles() {
	fileHasOnlyUserAddedTags := make(map[string]bool)
	for _, tagDef := range p.Tags {
		if tagDef.TagFile == "bagit.txt" || tagDef.TagFile == "bag-info.txt" {
			continue
		}
		if _, ok := fileHasOnlyUserAddedTags[tagDef.TagFile]; !ok {
			fileHasOnlyUserAddedTags[tagDef.TagFile] = tagDef.IsUserAddedTag
			continue
		}
		if !tagDef.IsUserAddedTag {
			fileHasOnlyUserAddedTags[tagDef.TagFile] = false
		}
	}
	for fileName, isUserAddedFile := range fileHasOnlyUserAddedTags {
		if isUserAddedFile {
			tagDefs, _ := p.FindMatchingTags("TagFile", fileName)
			for _, tagDef := range tagDefs {
				tagDef.IsUserAddedFile = true
			}
		}
	}
}

// TagsInFile returns all of the tag definition objects in tagFileName,
// sorted by name. This is equivalent to calling FindMatchingTags("TagFile", tagFileName),
// except it returns results in sorted order.
func (p *BagItProfile) TagsInFile(tagFileName string) []*TagDefinition {
	matches := make([]*TagDefinition, 0)
	for _, tagDef := range p.Tags {
		if tagDef.TagFile == tagFileName {
			matches = append(matches, tagDef)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].TagName < matches[j].TagName
	})
	return matches
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

	bagitVersionField := form.AddMultiValueField("AcceptBagItVersion", "Accept BagIt Version", p.AcceptBagItVersion, true)
	bagitVersionField.Choices = MakeMultiChoiceList(p.AcceptBagItVersion, p.AcceptBagItVersion)
	bagitVersionField.Help = "Which BagIt versions are allowed in this profile?"

	acceptSerializationField := form.AddMultiValueField("AcceptSerialization", "AcceptSerialization", p.AcceptSerialization, true)
	acceptSerializationField.Choices = MakeMultiChoiceList(p.AcceptSerialization, p.AcceptSerialization)
	acceptSerializationField.Help = "If bags using this profile can be serialized to tar, zip or other formats, enter the mime types for those formats here. E.g. application/x-tar, application/zip, etc. See https://en.wikipedia.org/wiki/List_of_archive_formats for a full list."

	allowFetch := form.AddField("AllowFetchTxt", "AllowFetchTxt", strconv.FormatBool(p.AllowFetchTxt), true)
	allowFetch.Choices = YesNoChoices(p.AllowFetchTxt)
	allowFetch.Help = "Does this profile allow the fetch.txt file to specify that some contents should be fetched from a URL rather than being packaged inside the bag?"

	form.AddField("BaseProfileID", "Base Profile", p.BaseProfileID, true)
	form.AddField("Description", "Description", p.Description, true)
	form.AddField("IsBuiltIn", "IsBuiltIn", strconv.FormatBool(p.IsBuiltIn), true)

	manifestsAllowedField := form.AddMultiValueField("ManifestsAllowed", "ManifestsAllowed", p.ManifestsAllowed, true)
	manifestsAllowedField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.ManifestsAllowed)
	manifestsAllowedField.Help = "Which manifest algorithms are allowed in this profile? E.g. md5, sha1, sha256, sha512, etc."

	manifestsRequiredField := form.AddMultiValueField("ManifestsRequired", "ManifestsRequired", p.ManifestsRequired, true)
	manifestsRequiredField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.ManifestsRequired)
	manifestsRequiredField.Help = "Which manifests MUST be present in this profile? Leave empty if users can choose from any of the allowed manifests."

	nameField := form.AddField("Name", "Name", p.Name, true)
	if p.IsBuiltIn {
		nameField.Attrs["readonly"] = "readonly"
	}

	serlializationField := form.AddField("Serialization", "Serialization", p.Serialization, true)
	serlializationField.Choices = MakeChoiceList(constants.SerializationOptions, p.Serialization)
	serlializationField.Help = "Can bags using this profile be serialized to tar, zip, or some other format?"

	tagFilesAllowed := strings.Join(p.TagFilesAllowed, "\n")
	tfaField := form.AddField("TagFilesAllowed", "TagFilesAllowed", tagFilesAllowed, true)
	tfaField.Help = "List tag files allowed, one item per line, or enter * to allow any tag files."

	tagFilesRequired := form.AddMultiValueField("TagFilesRequired", "TagFilesRequired", p.TagFilesRequired, true)
	tagFilesRequired.Help = "Which tag files MUST be included in bags conforming to this profile? List them one per line. You don't need to include bagit.txt or bag-info.txt, as those are assumed."

	tagManifestsAllowedField := form.AddMultiValueField("TagManifestsAllowed", "TagManifestsAllowed", p.TagManifestsAllowed, true)
	tagManifestsAllowedField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.TagManifestsAllowed)
	tagManifestsAllowedField.Help = "Which tag manifest algorithms are allowed in this profile? E.g. md5, sha1, sha256, sha512, etc."

	tagManifestsRequiredField := form.AddMultiValueField("TagManifestsRequired", "TagManifestsRequired", p.TagManifestsRequired, true)
	tagManifestsRequiredField.Choices = MakeMultiChoiceList(constants.PreferredAlgsInOrder, p.TagManifestsRequired)
	tagManifestsRequiredField.Help = "Which tag manifests MUST be present in this profile? Leave empty if users can choose from any of the allowed manifests."

	tarDirMustMatchField := form.AddField("TarDirMustMatchName", "TarDirMustMatchName", strconv.FormatBool(p.TarDirMustMatchName), true)
	tarDirMustMatchField.Choices = YesNoChoices(p.TarDirMustMatchName)

	// BagItProfileInfo
	form.AddField("InfoIdentifier", "Identifier", p.BagItProfileInfo.BagItProfileIdentifier, false)
	form.AddField("InfoContactEmail", "Contact Email", p.BagItProfileInfo.ContactEmail, false)
	form.AddField("InfoContactName", "Contact Name", p.BagItProfileInfo.ContactName, false)
	form.AddField("InfoExternalDescription", "External Description", p.BagItProfileInfo.ExternalDescription, false)
	form.AddField("InfoSourceOrganization", "Source Organization", p.BagItProfileInfo.SourceOrganization, false)
	form.AddField("InfoVersion", "Version", p.BagItProfileInfo.Version, false)

	for field, errMsg := range p.Errors {
		form.Fields[field].Error = errMsg
	}

	return form
}

// NewBagItProfileCreationForm returns a form for starting the
// BagIt Profile creation process. This form contains only a single
// element: a select list containing a list of base profiles on
// which to base the new profile.
func NewBagItProfileCreationForm() (*Form, error) {
	form := NewForm(constants.TypeBagItProfile, constants.EmptyUUID, nil)
	result := ObjList(constants.TypeBagItProfile, "obj_name", 10000, 0)
	if result.Error != nil {
		return nil, result.Error
	}
	choices := make([]Choice, len(result.BagItProfiles))
	for i, profile := range result.BagItProfiles {
		choices[i] = Choice{
			Label:    profile.Name,
			Value:    profile.ID,
			Selected: false,
		}
	}
	baseProfileID := form.AddField("BaseProfileID", "Base this profile on...", "", true)
	baseProfileID.Choices = choices
	return form, nil
}
