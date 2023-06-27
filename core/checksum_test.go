package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestNewChecksum(t *testing.T) {
	cs := core.NewChecksum(constants.FileTypePayload, constants.AlgMd5, "1234abcd")
	assert.Equal(t, constants.FileTypePayload, cs.Source)
	assert.Equal(t, constants.AlgMd5, cs.Algorithm)
	assert.Equal(t, "1234abcd", cs.Digest)
}

func TestChecksumSourceName(t *testing.T) {
	cs := core.NewChecksum(constants.FileTypePayload, constants.AlgMd5, "1234abcd")
	assert.Equal(t, "payload file", cs.SourceName())

	cs.Source = constants.FileTypeManifest
	assert.Equal(t, "manifest-md5.txt", cs.SourceName())

	cs.Algorithm = constants.AlgSha1
	assert.Equal(t, "manifest-sha1.txt", cs.SourceName())

	cs.Algorithm = constants.AlgSha256
	assert.Equal(t, "manifest-sha256.txt", cs.SourceName())

	cs.Algorithm = constants.AlgSha512
	assert.Equal(t, "manifest-sha512.txt", cs.SourceName())
}
