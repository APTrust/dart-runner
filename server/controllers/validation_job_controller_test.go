package controllers_test

import (
	"testing"

	"github.com/APTrust/dart-runner/core"
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
}

func TestValidationJobShowFiles(t *testing.T) {

}

func TestValidationJobAddFile(t *testing.T) {

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
