package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// GET /validation_jobs/files/:id
func ValidationJobShowFiles(c *gin.Context) {
	templateData, err := InitFileChooser(c)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	valJob := result.ValidationJob()

	directory := c.Query("directory")
	if directory == "" {
		directory, _ = core.GetAppSetting("Bagging Directory")
	}
	items, err := GetDirList(directory)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	templateData["job"] = valJob
	templateData["items"] = items
	templateData["showJobFiles"] = len(valJob.PathsToValidate) > 0

	templateData["dragDropInstructions"] = "Drag and drop the items from the left that you want to validate."
	templateData["fileDeletionUrl"] = fmt.Sprintf("/validation_jobs/delete_file/%s", valJob.ID)
	templateData["jobDeletionUrl"] = fmt.Sprintf("/validation_jobs/delete/%s", valJob.ID)
	templateData["nextButtonUrl"] = fmt.Sprintf("/validation_jobs/profile/%s", valJob.ID)
	templateData["addFileUrl"] = fmt.Sprintf("/validation_jobs/add_file/%s", valJob.ID)

	c.HTML(http.StatusOK, "job/files.html", templateData)
}

// POST /validation_jobs/add_file/:id
func ValidationJobAddFile(c *gin.Context) {
	fileToAdd := c.PostForm("fullPath")
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	valJob := result.ValidationJob()
	index := -1
	for i, filename := range valJob.PathsToValidate {
		if fileToAdd == filename {
			index = i
			break
		}
	}
	if index < 0 {
		valJob.PathsToValidate = append(valJob.PathsToValidate, fileToAdd)
		err := core.ObjSaveWithoutValidation(valJob)
		if err != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, err)
			return
		}
	}
	fileBrowserPath := c.PostForm("directory")
	values := url.Values{}
	values.Set("directory", fileBrowserPath)
	c.Redirect(http.StatusFound, fmt.Sprintf("/validation_jobs/files/%s?%s", valJob.ID, values.Encode()))
}

// POST /validation_jobs/delete_file/:id
func ValidationJobDeleteFile(c *gin.Context) {
	fileToDelete := c.PostForm("fullPath")
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	valJob := result.ValidationJob()
	index := -1
	for i, filename := range valJob.PathsToValidate {
		if fileToDelete == filename {
			index = i
			break
		}
	}
	if index >= 0 {
		valJob.PathsToValidate = util.RemoveFromSlice[string](valJob.PathsToValidate, index)
		err := core.ObjSaveWithoutValidation(valJob)
		if err != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, err)
			return
		}
	}
	fileBrowserPath := c.PostForm("directory")
	values := url.Values{}
	values.Set("directory", fileBrowserPath)
	c.Redirect(http.StatusFound, fmt.Sprintf("/validation_jobs/files/%s?%s", valJob.ID, values.Encode()))
}

// GET /validation_jobs/profiles/:id
func ValidationJobShowProfiles(c *gin.Context) {
	// Show list of BagIt profiles
}

// POST /validation_jobs/profiles/:id
func ValidationJobSaveProfile(c *gin.Context) {
	// Save the selected profile to ValidationJob
	// and redirect to ValidationJobReview
}

// GET /validation_jobs/review/:id
func ValidationJobReview(c *gin.Context) {
	// Show validation job details and Run button.
}

// GET /validation_jobs/run/:id
//
// By REST standards, this should be a POST. However, the Server
// Send Events standard for JavaScript only supports GET.
func ValidationJobRun(c *gin.Context) {
	// Run this job using Server Sent Events.
	// See JobRunExecute()
}
