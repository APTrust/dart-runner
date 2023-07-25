package controllers

import (
	"net/http"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// GET /jobs/files/:id
func JobShowFiles(c *gin.Context) {
	job, items, err := GetJobAndDirList(c.Param("id"), c.Query("directory"))
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
	}
	data := gin.H{
		"job":   job,
		"items": items,
	}
	c.HTML(http.StatusOK, "partials/file_browser.html", data)
}

// POST /jobs/files/:id
func JobSaveFiles(c *gin.Context) {

}

/*

-- Linux --

Home
Documents
Downloads
Music
Pictures
Videos
Trash

-- Mac --

Home
Desktop
Documents
Download

-- Windows --

Desktop
Downloads
Documents
Pictures
Videos
Root

Also show attached volumes

*/

func GetJobAndDirList(jobId, dirname string) (*core.Job, []*util.ExtendedFileInfo, error) {
	var entries []*util.ExtendedFileInfo
	var err error
	if dirname == "" {
		entries, err = core.Dart.Paths.DefaultPaths()
	} else {
		entries, err = util.ListDirectory(dirname)
	}
	if err != nil {
		return nil, nil, err
	}
	result := core.ObjFind(jobId)
	if result.Error != nil {
		return nil, entries, result.Error
	}
	return result.Job(), entries, nil
}
