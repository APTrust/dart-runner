package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// GET /upload_jobs/new
func UploadJobNew(c *gin.Context) {
	uploadJob := core.NewUploadJob()
	err := core.ObjSaveWithoutValidation(uploadJob)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/upload_jobs/files/%s", uploadJob.ID))

}

// GET /upload_jobs/files/:id
func UploadJobShowFiles(c *gin.Context) {
	templateData, err := InitFileChooser(c)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
		return
	}
	uploadJob, err := loadUploadJob(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}

	directory := c.Query("directory")
	if directory == "" {
		directory, _ = core.GetAppSetting("Bagging Directory")
	}
	items, err := GetDirList(directory)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	templateData["job"] = uploadJob
	templateData["items"] = items
	templateData["showJobFiles"] = len(uploadJob.PathsToUpload) > 0
	templateData["sourceFiles"] = uploadJob.PathsToUpload
	templateData["showJumpMenu"] = true

	templateData["dragDropInstructions"] = "Drag and drop the items from the left that you want to upload."
	templateData["fileDeletionUrl"] = fmt.Sprintf("/upload_jobs/delete_file/%s", uploadJob.ID)
	templateData["jobDeletionUrl"] = fmt.Sprintf("/upload_jobs/delete/%s", uploadJob.ID)
	templateData["nextButtonUrl"] = fmt.Sprintf("/upload_jobs/profiles/%s", uploadJob.ID)
	templateData["addFileUrl"] = fmt.Sprintf("/upload_jobs/add_file/%s", uploadJob.ID)

	c.HTML(http.StatusOK, "job/files.html", templateData)

}

// POST /upload_jobs/add_file/:id
func UploadJobAddFile(c *gin.Context) {
	fileToAdd := c.PostForm("fullPath")
	uploadJob, err := loadUploadJob(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	index := -1
	for i, filename := range uploadJob.PathsToUpload {
		if fileToAdd == filename {
			index = i
			break
		}
	}
	if index < 0 {
		uploadJob.PathsToUpload = append(uploadJob.PathsToUpload, fileToAdd)
		err := core.ObjSaveWithoutValidation(uploadJob)
		if err != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, err)
			return
		}
	}
	fileBrowserPath := c.PostForm("directory")
	values := url.Values{}
	values.Set("directory", fileBrowserPath)
	c.Redirect(http.StatusFound, fmt.Sprintf("/upload_jobs/files/%s?%s", uploadJob.ID, values.Encode()))
}

// POST /upload_jobs/delete_file/:id
func UploadJobDeleteFile(c *gin.Context) {
	fileToDelete := c.PostForm("fullPath")
	uploadJob, err := loadUploadJob(c.Param("id"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, err)
		return
	}
	index := -1
	for i, filename := range uploadJob.PathsToUpload {
		if fileToDelete == filename {
			index = i
			break
		}
	}
	if index >= 0 {
		uploadJob.PathsToUpload = util.RemoveFromSlice[string](uploadJob.PathsToUpload, index)
		err := core.ObjSaveWithoutValidation(uploadJob)
		if err != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, err)
			return
		}
	}
	fileBrowserPath := c.PostForm("directory")
	values := url.Values{}
	values.Set("directory", fileBrowserPath)
	c.Redirect(http.StatusFound, fmt.Sprintf("/upload_jobs/files/%s?%s", uploadJob.ID, values.Encode()))
}

// GET /upload_jobs/profiles/:id
func UploadJobShowTargets(c *gin.Context) {

}

// POST /upload_jobs/profiles/:id
func UploadJobSaveTarget(c *gin.Context) {

}

// GET /upload_jobs/review/:id
func UploadJobReview(c *gin.Context) {

}

// GET /upload_jobs/run/:id
//
// By REST standards, this should be a POST. However, the Server
// Send Events standard for JavaScript only supports GET.
func UploadJobRun(c *gin.Context) {

}

func loadUploadJob(uploadJobID string) (*core.UploadJob, error) {
	result := core.ObjFind(uploadJobID)
	return result.UploadJob(), result.Error
}
