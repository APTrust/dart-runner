package core

// ProfileInfo contains meta info about the profile.
type ProfileInfo struct {
	BagItProfileIdentifier string `json:"bagItProfileIdentifier" form:"InfoIdentifier"`
	BagItProfileVersion    string `json:"bagItProfileVersion" form:"InfoProfileVersion"`
	ContactEmail           string `json:"contactEmail" form:"InfoContactEmail"`
	ContactName            string `json:"contactName" form:"InfoContactName"`
	ExternalDescription    string `json:"externalDescription" form:"InfoExternalDescription"`
	SourceOrganization     string `json:"sourceOrganization" form:"InfoSourceOrganization"`
	Version                string `json:"version" form:"InfoVersion"`
}

func CopyProfileInfo(info ProfileInfo) ProfileInfo {
	return ProfileInfo{
		BagItProfileIdentifier: info.BagItProfileIdentifier,
		BagItProfileVersion:    info.BagItProfileVersion,
		ContactEmail:           info.ContactEmail,
		ContactName:            info.ContactName,
		ExternalDescription:    info.ExternalDescription,
		SourceOrganization:     info.SourceOrganization,
		Version:                info.Version,
	}
}
