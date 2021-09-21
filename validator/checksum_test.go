package validator_test

import (
	"testing"

	"github.com/APTrust/dart-runner/validator"
	"github.com/stretchr/testify/assert"
)

func TestNewChecksum(t *testing.T) {
	cs := validator.NewChecksum(validator.SourcePayloadFile, validator.AlgMd5, "1234abcd")
	assert.Equal(t, validator.SourcePayloadFile, cs.Source)
	assert.Equal(t, validator.AlgMd5, cs.Algorithm)
	assert.Equal(t, "1234abcd", cs.Digest)
}

func TestChecksumSourceName(t *testing.T) {
	cs := validator.NewChecksum(validator.SourcePayloadFile, validator.AlgMd5, "1234abcd")
	assert.Equal(t, "payload file", cs.SourceName())

	cs.Source = validator.SourceManifest
	assert.Equal(t, "manifest-md5.txt", cs.SourceName())

	cs.Algorithm = validator.AlgSha1
	assert.Equal(t, "manifest-sha1.txt", cs.SourceName())

	cs.Algorithm = validator.AlgSha256
	assert.Equal(t, "manifest-sha256.txt", cs.SourceName())

	cs.Algorithm = validator.AlgSha512
	assert.Equal(t, "manifest-sha512.txt", cs.SourceName())
}
