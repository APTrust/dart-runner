package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageOperation(t *testing.T) {
	op := core.NewPackageOperation("", "", []string{})
	assert.NotNil(t, op.Result)
	assert.False(t, op.Validate())
	assert.Equal(t, 3, len(op.Errors))
	assert.Equal(t, "Package name is required.", op.Errors["PackageOperation.PackageName"])
	assert.Equal(t, "Output path is required.", op.Errors["PackageOperation.OutputPath"])
	assert.Equal(t, "Specify at least one file or directory to package.", op.Errors["PackageOperation.SourceFiles"])

	op.SourceFiles = append(op.SourceFiles, "file-does-not-exist")
	assert.False(t, op.Validate())
	assert.Equal(t, 3, len(op.Errors))
	assert.Equal(t, "Specify at least one file or directory to package.", op.Errors["PackageOperation.SourceFiles"])

	sourceFiles := []string{
		util.PathToUnitTestBag("test.edu.btr_good_sha256.tar"),
		util.PathToUnitTestBag("test.edu.btr_good_sha512.tar"),
	}
	op = core.NewPackageOperation("bag.tar", "/path/to/output", sourceFiles)
	assert.True(t, op.Validate())
	assert.Equal(t, 0, len(op.Errors))
}

func TestPackageOperationToForm(t *testing.T) {
	sourceFiles := []string{
		"/usr/local/photos",
		"/usr/local/pdfs",
		"/usr/local/music",
	}
	op := core.NewPackageOperation("photos.tar", "/home/josie/photos.tar", sourceFiles)
	op.BagItSerialization = constants.SerializationRequired
	op.PackageFormat = constants.PackageFormatBagIt

	form := op.ToForm()
	require.NotNil(t, form)
	assert.Equal(t, 4, len(form.Fields["BagItSerialization"].Choices))
	assert.Equal(t, constants.SerializationRequired, form.Fields["BagItSerialization"].Value)

	assert.Equal(t, "/home/josie/photos.tar", form.Fields["OutputPath"].Value)
	assert.Equal(t, "photos.tar", form.Fields["PackageName"].Value)

	assert.Equal(t, 2, len(form.Fields["PackageFormat"].Choices))
	assert.Equal(t, constants.PackageFormatBagIt, form.Fields["PackageFormat"].Value)

	assert.Equal(t, sourceFiles, form.Fields["SourceFiles"].Values)
}

func TestPruneSourceFiles(t *testing.T) {
	op := core.NewPackageOperation("", "", []string{})
	op.SourceFiles = append(op.SourceFiles, "file-does-not-exist")
	assert.Equal(t, 1, len(op.SourceFiles))

	// Prune should remove the one file that does not exist.
	op.PruneSourceFiles()
	assert.Equal(t, 0, len(op.SourceFiles))

	btr256 := util.PathToUnitTestBag("test.edu.btr_good_sha256.tar")
	sourceFiles := []string{
		btr256,
		btr256,
	}
	op = core.NewPackageOperation("bag.tar", "/path/to/output", sourceFiles)
	assert.Equal(t, 2, len(op.SourceFiles))

	// Prune should remove duplicate entries from SourceFiles.
	op.PruneSourceFiles()
	assert.Equal(t, 1, len(op.SourceFiles))
}
