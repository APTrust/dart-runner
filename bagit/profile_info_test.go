package bagit_test

import (
	"testing"

	"github.com/APTrust/dart-runner/bagit"
	"github.com/stretchr/testify/assert"
)

func TestCopyProfileInfo(t *testing.T) {
	info := bagit.ProfileInfo{
		BagItProfileIdentifier: "ident1",
		BagItProfileVersion:    "version1",
		ContactEmail:           "user@example.com",
		ContactName:            "giuseppe",
		ExternalDescription:    "external desc",
		SourceOrganization:     "source org",
	}
	copyOfInfo := bagit.CopyProfileInfo(info)
	assert.Equal(t, info.BagItProfileVersion, copyOfInfo.BagItProfileVersion)
	assert.Equal(t, info.BagItProfileIdentifier, copyOfInfo.BagItProfileIdentifier)
	assert.Equal(t, info.ContactEmail, copyOfInfo.ContactEmail)
	assert.Equal(t, info.ContactName, copyOfInfo.ContactName)
	assert.Equal(t, info.ExternalDescription, copyOfInfo.ExternalDescription)
	assert.Equal(t, info.SourceOrganization, copyOfInfo.SourceOrganization)
}
