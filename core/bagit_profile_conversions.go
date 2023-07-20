package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/stew/slice"
)

/*
	This file contains additional BagItProfile functions for converting
	to and from other BagIt Profile formats. We don't want to clutter
	bagit_profile.go with these functions.
*/

// ToStandardFormat converts this BagIt Profile to a version 1.3.0
// BagIt Profile. See https://bagit-profiles.github.io/bagit-profiles-specification/
// and https://github.com/bagit-profiles/bagit-profiles-specification
func (p *BagItProfile) ToStandardFormat() *StandardProfile {
	sp := NewStandardProfile()
	sp.AcceptBagItVersion = p.AcceptBagItVersion
	sp.AcceptSerialization = p.AcceptSerialization
	sp.AllowFetchTxt = p.AllowFetchTxt
	sp.BagItProfileInfo = StandardProfileInfo{
		BagItProfileIdentifier: p.BagItProfileInfo.BagItProfileIdentifier,
		BagItProfileVersion:    p.BagItProfileInfo.BagItProfileVersion,
		ContactEmail:           p.BagItProfileInfo.ContactEmail,
		ContactName:            p.BagItProfileInfo.ContactName,
		ExternalDescription:    p.BagItProfileInfo.ExternalDescription,
		SourceOrganization:     p.BagItProfileInfo.SourceOrganization,
		Version:                p.BagItProfileInfo.Version,
	}
	sp.ManifestsAllowed = p.ManifestsAllowed
	sp.ManifestsRequired = p.ManifestsRequired
	sp.Serialization = p.Serialization
	sp.TagFilesAllowed = p.TagFilesAllowed
	sp.TagFilesRequired = p.TagFilesRequired
	sp.TagManifestsAllowed = p.TagManifestsAllowed
	sp.TagManifestsRequired = p.TagManifestsRequired

	for _, tagDef := range p.Tags {
		if tagDef.TagFile == "bagit.txt" {
			continue
		}
		if tagDef.TagFile == "bag-info.txt" {
			sp.BagInfo[tagDef.TagName] = StandardProfileTagDef{
				Required:    tagDef.Required,
				Values:      tagDef.Values,
				Description: tagDef.Help,
				Recommended: strings.Contains(tagDef.Help, "Recommended"),
			}
		} else {
			// We can't specify tag info outside of bag-info.txt,
			// but if we find a required tag in another file,
			// we can assume that tag file is required.
			if tagDef.Required && !slice.Contains(p.TagFilesRequired, tagDef.TagFile) {
				p.TagFilesRequired = append(p.TagFilesRequired, tagDef.TagFile)
			}
			if !slice.Contains(p.TagFilesAllowed, tagDef.TagFile) {
				p.TagFilesRequired = append(p.TagFilesAllowed, tagDef.TagFile)
			}
		}
	}
	return sp
}

// GuessProfileTypeFromJson tries to determine the type of a BagIt profile based
// on its structure.
func GuessProfileTypeFromJson(jsonBytes []byte) (string, error) {
	obj := make(map[string]interface{})
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return constants.ProfileTypeUnknown, err
	}
	return GuessProfileType(obj), nil
}

// GuessProfileType tries to determine the type of a BagIt profile based
// on its structure.
func GuessProfileType(obj map[string]interface{}) string {
	profileType := constants.ProfileTypeUnknown
	if util.IsListType(obj["tags"]) {
		profileType = constants.ProfileTypeDart
	} else if util.IsListType(obj["ordered"]) {
		profileType = constants.ProfileTypeLOCOrdered
	} else if util.IsMapType(obj["Bag-Info"]) {
		profileType = constants.ProfileTypeStandard
	} else {
		everythingLooksLikeATag := true
		for _, value := range obj {
			if !util.IsMapType(value) {
				everythingLooksLikeATag = false
				break
			}
			item, ok := value.(map[string]interface{})
			if !ok {
				everythingLooksLikeATag = false
				break
			}
			// Tags have at least one of these properties.
			_, hasFieldRequired := item["fieldRequired"]
			_, hasRequiredValue := item["requiredValue"]
			if !hasFieldRequired && !hasRequiredValue {
				everythingLooksLikeATag = false
				break
			}
		}
		if everythingLooksLikeATag {
			profileType = constants.ProfileTypeLOCUnordered
		}
	}
	return profileType
}

// ProfileFromLOCOrdered converts the JSON representation of an
// ordered Library of Congress BagIt profile to a DART BagIt profile.
func ProfileFromLOCOrdered(jsonBytes []byte, sourceUrl string) (*BagItProfile, error) {
	locOrderedProfile := &LOCOrderedProfile{}
	err := json.Unmarshal(jsonBytes, locOrderedProfile)
	if err != nil {
		return nil, err
	}
	profile := NewBagItProfile()
	profile.Name = getProfileName(sourceUrl)
	profile.Description = profile.Name
	for _, tagMap := range locOrderedProfile.Tags {
		for tagName, locTagDef := range tagMap {
			err = convertLOCTag(profile, tagName, locTagDef)
			if err != nil {
				return profile, err
			}
		}
	}
	return profile, nil
}

// ProfileFromLOCUnordered converts the JSON representation of an
// unordered Library of Congress BagIt profile to a DART BagIt profile.
func ProfileFromLOCUnordered(jsonBytes []byte, sourceUrl string) (*BagItProfile, error) {
	locUnorderedProfile := make(map[string]LOCTagDef)
	err := json.Unmarshal(jsonBytes, &locUnorderedProfile)
	if err != nil {
		return nil, err
	}
	profile := NewBagItProfile()
	profile.Name = getProfileName(sourceUrl)
	profile.Description = profile.Name
	for tagName, locTagDef := range locUnorderedProfile {
		err = convertLOCTag(profile, tagName, locTagDef)
		if err != nil {
			return profile, err
		}
	}
	return profile, nil
}

// convertLOCTag converts a Library of Congress tag definition to a DART tag def.
func convertLOCTag(profile *BagItProfile, tagName string, locTagDef LOCTagDef) error {
	var tagDef *TagDefinition

	// There should never be an error here, unless we search on an
	// unsupported field. TagName is supported.
	matchingTags, err := profile.FindMatchingTags("TagName", tagName)
	if err != nil {
		return err
	}

	// LOC profiles only define tags for the bag-info.txt file.
	// If our DART profile already defines that tag in bag-info.txt,
	// let's edit that tag definition instead of creating and appending
	// a new/redundant tag def.
	for _, t := range matchingTags {
		if t.TagFile == "bag-info.txt" {
			tagDef = t
			break
		}
	}
	if tagDef == nil {
		// Not found in existing DART profile, so we'll create a new tag def.
		tagDef = &TagDefinition{
			ID:      uuid.NewString(),
			TagFile: "bag-info.txt",
			TagName: tagName,
			Values:  make([]string, len(locTagDef.Values)),
		}
		profile.Tags = append(profile.Tags, tagDef)
	}

	// Now set the values
	tagDef.Required = locTagDef.Required
	tagDef.DefaultValue = locTagDef.DefaultValue
	copy(tagDef.Values, locTagDef.Values)
	if locTagDef.RequiredValue != "" {
		tagDef.Required = true
		tagDef.Values = []string{locTagDef.RequiredValue}
		tagDef.DefaultValue = locTagDef.RequiredValue
	}

	return nil
}

// getProfileName returns a placeholder name for a newly imported profile.
// We want to return a unique name to prevent unique constraint violation
// on dart.obj_type + dart.obj_name.
func getProfileName(sourceUrl string) string {
	name := fmt.Sprintf("Imported Profile %s", time.Now().Format(time.Stamp))
	if sourceUrl != "" {
		name = fmt.Sprintf("Profile imported from %s (%s)", sourceUrl, time.Now().Format(time.Stamp))
	}
	return name
}

// ConvertProfile converts a BagIt profile from a known format to a
// DART BagIt profile. It does not try to save the profile, because some
// imported profiles may not have all required info, and that will cause
// an error. User should be able to convert the profile and then edit it
// to correct missing or invalid properties. Accepted profile
// formats include Dart, Standard, LOC Ordered, and LOC Unordered.
func ConvertProfile(jsonBytes []byte, sourceUrl string) (*BagItProfile, error) {
	dartProfile := &BagItProfile{}
	profileType, err := GuessProfileTypeFromJson(jsonBytes)
	if err != nil {
		return nil, err
	}
	switch profileType {
	case constants.ProfileTypeDart:
		err = json.Unmarshal(jsonBytes, dartProfile)
	case constants.ProfileTypeStandard:
		standardProfile := &StandardProfile{}
		err = json.Unmarshal(jsonBytes, standardProfile)
		if err != nil {
			return nil, err
		}
		dartProfile = standardProfile.ToDartProfile()
	case constants.ProfileTypeLOCOrdered:
		dartProfile, err = ProfileFromLOCOrdered(jsonBytes, sourceUrl)
	case constants.ProfileTypeLOCUnordered:
		dartProfile, err = ProfileFromLOCUnordered(jsonBytes, sourceUrl)
	default:
		err = fmt.Errorf("Cannot convert unrecognized BagIt profile type.")
	}
	dartProfile.EnsureMinimumRequirements()
	return dartProfile, err
}
