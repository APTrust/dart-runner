package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
)

func TestPackageOperation(t *testing.T) {
	op := core.NewPackageOperation("", "", []string{})
	assert.NotNil(t, op.Result)
	assert.False(t, op.Validate())
	assert.Equal(t, 3, len(op.Errors))
	assert.Equal(t, "Package name is required.", op.Errors["PackageOperation.PackageName"])
	assert.Equal(t, "Output path is required.", op.Errors["PackageOperation.OutputPath"])
	assert.Equal(t, "Specify at least one file or directory to package.", op.Errors["PackageOperation.sourceFiles"])

	op.SourceFiles = append(op.SourceFiles, "file-does-not-exist")
	assert.False(t, op.Validate())
	assert.Equal(t, 3, len(op.Errors))
	assert.Equal(t, "The following files are missing: file-does-not-exist", op.Errors["PackageOperation.sourceFiles"])

	btr256 := util.PathToUnitTestBag("test.edu.btr_good_sha256.tar")
	btr512 := util.PathToUnitTestBag("test.edu.btr_good_sha512.tar")

	sourceFiles := []string{
		btr256,
		btr256,
	}
	op = core.NewPackageOperation("bag.tar", "/path/to/output", sourceFiles)
	assert.False(t, op.Validate())
	assert.Equal(t, "The following files are included more than once. Please remove duplicates: "+btr256, op.Errors["PackageOperation.sourceFiles"])

	sourceFiles = []string{
		btr256,
		btr512,
	}
	op = core.NewPackageOperation("bag.tar", "/path/to/output", sourceFiles)
	assert.True(t, op.Validate())
	assert.Equal(t, 0, len(op.Errors))
}
