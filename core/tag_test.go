package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/assert"
)

func TestTagFQName(t *testing.T) {
	tagDef := &core.Tag{
		TagFile: "bag-info.txt",
		TagName: "Source-Organization",
	}
	assert.Equal(t, "bag-info.txt/Source-Organization", tagDef.FullyQualifiedName())
}
