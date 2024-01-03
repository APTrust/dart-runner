package controllers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/stretchr/testify/require"
)

func TestValidationJobNew(t *testing.T) {
	defer core.ClearDartTable()
	expected := []string{
		"jumpMenu",
		"showHiddenFiles",
		"file-browser-item",
		"dropZone",
		"fileTotals",
		"attachDragAndDropEvents",
	}
	DoGetTestWithRedirect(t, "/validation_jobs/new", "/validation_jobs/files/", expected)

	valJob := core.ObjList(constants.TypeValidationJob, "obj_name", 1, 0).ValidationJob()
	require.NotNil(t, valJob)

	testValidationJobShowFiles(t, valJob.ID)
	testValidationJobAddFile(t, valJob.ID)
}

func testValidationJobShowFiles(t *testing.T, id string) {
	expected := []string{
		"jumpMenu",
		"showHiddenFiles",
		"file-browser-item",
		"dropZone",
		"fileTotals",
		"attachDragAndDropEvents",
	}
	DoSimpleGetTest(t, fmt.Sprintf("/validation_jobs/files/%s", id), expected)
}

func testValidationJobAddFile(t *testing.T, id string) {
	fileToAdd := filepath.Join(util.PathToTestData(), "aptrust-unit-test-job.json")
	params := url.Values{}
	params.Add("fullPath", fileToAdd)
	settings := PostTestSettings{
		EndpointUrl:              fmt.Sprintf("/validation_jobs/add_file/%s", id),
		Params:                   params,
		ExpectedResponseCode:     http.StatusFound,
		ExpectedRedirectLocation: fmt.Sprintf("/validation_jobs/files/%s?directory=", id),
		ExpectedContent:          []string{fileToAdd},
	}
	DoPostTestWithRedirect(t, settings)
}

func TestValidationJobDeleteFile(t *testing.T) {

}

func TestValidationJobShowProfiles(t *testing.T) {

}

func TestValidationJobSaveProfile(t *testing.T) {

}

func TestValidationJobReview(t *testing.T) {

}

func TestValidationJobRun(t *testing.T) {

}
