package controllers

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// GET /jobs/files/:id
func JobShowFiles(c *gin.Context) {
	directory := c.Query("directory")
	job, items, err := GetJobAndDirList(c.Param("id"), directory)
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
	}
	defaultPaths, err := core.Dart.Paths.DefaultPaths()
	if err != nil {
		AbortWithErrorHTML(c, http.StatusInternalServerError, err)
	}
	parentDir, parentDirShortName := GetParentDir(directory)
	showParentDirLink := directory != "" && directory != parentDir
	showJumpMenu := directory != ""
	data := gin.H{
		"job":                job,
		"items":              items,
		"parentDir":          parentDir,
		"parentDirShortName": parentDirShortName,
		"showParentDirLink":  showParentDirLink,
		"defaultPaths":       defaultPaths,
		"showJumpMenu":       showJumpMenu,
		"currentDir":         directory,
	}
	c.HTML(http.StatusOK, "job/files.html", data)
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

func GetParentDir(dirName string) (string, string) {
	parentDir := path.Dir(dirName)
	var parentDirShortName string
	parts := strings.Split(parentDir, string(os.PathSeparator))
	if len(parts) > 0 {
		parentDirShortName = parts[len(parts)-1]
	}
	if parentDirShortName == "" {
		parentDirShortName = parentDir
	}
	if parentDirShortName == "." {
		parentDirShortName = "Default Menu"
		parentDir = ""
	}
	return parentDir, parentDirShortName
}
