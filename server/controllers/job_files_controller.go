package controllers

import (
	"fmt"
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

// POST /jobs/add_file/:id
func JobAddFile(c *gin.Context) {
	// TODO: Make this AJAX or preserve file browser path.
	fileToAdd := c.PostForm("fullPath")
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	index := -1
	for i, filename := range job.PackageOp.SourceFiles {
		if fileToAdd == filename {
			index = i
			break
		}
	}
	if index < 0 {
		job.PackageOp.SourceFiles = append(job.PackageOp.SourceFiles, fileToAdd)
		err := core.ObjSave(job)
		if err != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, err)
			return
		}
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/jobs/files/%s", job.ID))
}

// POST /jobs/delete_file/:id
func JobDeleteFile(c *gin.Context) {
	// TODO: Make this AJAX or preserve file browser path.
	fileToDelete := c.PostForm("fullPath")
	result := core.ObjFind(c.Param("id"))
	if result.Error != nil {
		AbortWithErrorHTML(c, http.StatusNotFound, result.Error)
		return
	}
	job := result.Job()
	index := -1
	for i, filename := range job.PackageOp.SourceFiles {
		if fileToDelete == filename {
			index = i
			break
		}
	}
	if index >= 0 {
		util.RemoveFromSlice[string](job.PackageOp.SourceFiles, index)
		err := core.ObjSave(job)
		if err != nil {
			AbortWithErrorHTML(c, http.StatusNotFound, err)
			return
		}
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/jobs/files/%s", job.ID))
}

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
