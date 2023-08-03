package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// GET /jobs/upload/:id
func JobShowUpload(c *gin.Context) {
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	form, err := GetUploadTargetsForm(job)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	data := gin.H{
		"job":  job,
		"form": form,
	}
	c.HTML(http.StatusOK, "job/uploads.html", data)
}

// POST /jobs/upload/:id
func JobSaveUpload(c *gin.Context) {

}

func GetUploadTargetsForm(job *core.Job) (*core.Form, error) {
	selectedTargets := AlreadySelectedTargets(job)
	form := core.NewForm("", "", nil)
	targetsField := form.AddMultiValueField("UploadTargets", "Upload Targets", selectedTargets, false)
	targetChoices, err := GetAvailableUploadTargets(selectedTargets)
	if err != nil {
		return nil, err
	}
	targetsField.Choices = targetChoices
	return form, nil
}

func GetAvailableUploadTargets(selectedTargets []string) ([]core.Choice, error) {
	result := core.ObjList(constants.TypeStorageService, "obj_name", 1000, 0)
	if result.Error != nil {
		return nil, result.Error
	}
	targets := make([]core.Choice, 0)
	for _, ss := range result.StorageServices {
		isValid := ss.Validate()
		if !isValid {
			core.Dart.Log.Warn("Omitting storage service '%s' from upload targets due to validation errors: %v", ss.Name, ss.Errors)
			continue
		}
		if !ss.AllowsUpload {
			core.Dart.Log.Warn("Omitting storage service '%s' from upload targets because it does not allow uploads", ss.Name)
		} else {
			choice := core.Choice{
				Label:    ss.Name,
				Value:    ss.ID,
				Selected: util.StringListContains(selectedTargets, ss.ID),
			}
			targets = append(targets, choice)
		}
	}
	return targets, nil
}

func AlreadySelectedTargets(job *core.Job) []string {
	selected := make([]string, 0)
	for _, uploadOp := range job.UploadOps {
		if uploadOp.StorageService != nil {
			selected = append(selected, uploadOp.StorageService.ID)
		}
	}
	return selected
}
