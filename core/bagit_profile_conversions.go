package core

import (
	"encoding/json"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
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
