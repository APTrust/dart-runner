package bagit

// ProfileInfo contains meta info about the profile.
type ProfileInfo struct {
	BagItProfileIdentifier string `json:"bagItProfileIdentifier"`
	BagItProfileVersion    string `json:"bagItProfileVersion"`
	ContactEmail           string `json:"contactEmail"`
	ContactName            string `json:"contactName"`
	ExternalDescription    string `json:"externalDescription"`
	SourceOrganization     string `json:"sourceOrganization"`
}

func CopyProfileInfo(info ProfileInfo) ProfileInfo {
	return ProfileInfo{
		BagItProfileIdentifier: info.BagItProfileIdentifier,
		BagItProfileVersion:    info.BagItProfileVersion,
		ContactEmail:           info.ContactEmail,
		ContactName:            info.ContactName,
		ExternalDescription:    info.ExternalDescription,
		SourceOrganization:     info.SourceOrganization,
	}
}
