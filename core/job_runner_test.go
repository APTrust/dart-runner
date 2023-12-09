package core_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRunnerTestJob(t *testing.T, bagName string) *core.Job {
	workflow := loadJsonWorkflow(t)
	files := []string{
		filepath.Join(util.PathToTestData(), "files"),
	}
	outputPath := filepath.Join(os.TempDir(), bagName)
	tags := getTestTags()
	jobParams := core.NewJobParams(workflow, bagName, outputPath, files, tags)
	return jobParams.ToJob()
}

func testJobRunner(t *testing.T, bagName string, withCleanup bool) {
	job := getRunnerTestJob(t, bagName)
	outputDir := path.Dir(job.PackageOp.OutputPath)
	defer func() {
		if withCleanup && util.LooksSafeToDelete(job.PackageOp.OutputPath, 12, 2) {
			os.Remove(job.PackageOp.OutputPath)
			fileName := strings.TrimSuffix(filepath.Base(job.PackageOp.OutputPath), path.Ext(job.PackageOp.OutputPath))
			artifactsDir := filepath.Join(path.Dir(job.PackageOp.OutputPath), fileName+"_artifacts")
			os.RemoveAll(artifactsDir)
		}
	}()

	require.True(t, job.Validate(), job.Errors)
	retVal := core.RunJob(job, withCleanup, false)
	assert.Equal(t, constants.ExitOK, retVal)

	assert.True(t, job.PackageOp.Result.Succeeded())
	assert.True(t, job.ValidationOp.Result.Succeeded())
	for _, op := range job.UploadOps {
		assert.True(t, op.Result.Succeeded())
	}

	lastUpload := job.UploadOps[len(job.UploadOps)-1]
	if withCleanup {
		assert.Contains(t, lastUpload.Result.Info, "was deleted at")
	} else {
		assert.Contains(t, lastUpload.Result.Info, "Bag file(s) remain")
	}

	assertArtifactsWereSaved(t, job, outputDir)
}

func assertArtifactsWereSaved(t *testing.T, job *core.Job, outputDir string) {
	// In GUI mode, artifacts go into the SQLite DB
	if core.Dart.RuntimeMode == constants.ModeDartGUI {
		artifacts, err := core.ArtifactListByJobID(job.ID)
		require.NoError(t, err)
		for _, alg := range job.BagItProfile.ManifestsRequired {
			manifestName := fmt.Sprintf("manifest-%s.txt", alg)
			found := false
			for _, artifact := range artifacts {
				if artifact.FileName == manifestName {
					found = true
					break
				}
			}
			assert.True(t, found, manifestName)
		}
		foundBagIt := false
		foundBagInfo := false
		for _, artifact := range artifacts {
			if artifact.FileName == "bagit.txt" {
				foundBagIt = true
			}
			if artifact.FileName == "bag-info.txt" {
				foundBagInfo = true
			}
		}
		assert.True(t, foundBagIt, "bagit.txt")
		assert.True(t, foundBagInfo, "bag-info.txt")
	} else {
		// For dart runner and apt-cmd, artifacts go into output dir.
		for _, alg := range job.BagItProfile.ManifestsRequired {
			manifestName := fmt.Sprintf("manifest-%s.txt", alg)
			assert.True(t, util.FileExists(filepath.Join(job.ArtifactsDir, manifestName)))
		}
		assert.True(t, util.FileExists(filepath.Join(job.ArtifactsDir, "bagit.txt")), filepath.Join(job.ArtifactsDir, "bagit.txt"))
		assert.True(t, util.FileExists(filepath.Join(job.ArtifactsDir, "bag-info.txt")), filepath.Join(job.ArtifactsDir, "bag-info.txt"))
	}

}

func TestJobRunnerWithCleanup(t *testing.T) {
	testJobRunner(t, "bag_with_cleanup.tar", true)
}

func TestJobRunnerNoCleanup(t *testing.T) {
	testJobRunner(t, "bag_without_cleanup.tar", false)
}
