package server

import (
	"html/template"
	"io"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/server/controllers"
	"github.com/APTrust/dart-runner/util"
	"github.com/gin-gonic/gin"
)

// Run runs the Registry application. This is called from main() to start
// the app.
func Run() {
	r := InitAppEngine(false)
	r.Run()
}

// InitAppEngine sets up the whole Gin application, loading templates and
// middleware and defining routes. The test suite can use this to get an
// instance of the Gin engine to bind to.
//
// Set param discardStdOut during unit/integration tests to suppress
// Gin's STDOUT logging. Those log statements are useful in development,
// but can be verbose and clutter the test output.
func InitAppEngine(discardStdOut bool) *gin.Engine {
	var r *gin.Engine
	if discardStdOut {
		r = gin.New()
		r.Use(gin.Recovery())
		gin.DefaultWriter = io.Discard
	} else {
		r = gin.Default()
	}
	initTemplates(r)
	initRoutes(r)
	return r
}

// initTemplates loads templates and sets up template helper functions.
func initTemplates(router *gin.Engine) {

	// Note: We can't put workflowList in util with the
	// other template helpers because it creates a cyclical
	// import cycle between core and util. So we define the
	// body of that helper here inline.

	router.SetFuncMap(template.FuncMap{
		"dateISO":        util.DateISO,
		"dateTimeISO":    util.DateTimeISO,
		"dateTimeUS":     util.DateTimeUS,
		"dateUS":         util.DateUS,
		"dict":           util.Dict,
		"dirStats":       util.DirStats,
		"displayDate":    util.DisplayDate,
		"escapeAttr":     util.EscapeAttr,
		"escapeHTML":     util.EscapeHTML,
		"fileIconFor":    util.FileIconFor,
		"humanSize":      util.HumanSize,
		"strEq":          util.StrEq,
		"truncate":       util.Truncate,
		"truncateMiddle": util.TruncateMiddle,
		"truncateStart":  util.TruncateStart,
		"unixToISO":      util.UnixToISO,
		"workflowList":   func() []core.NameIDPair { return core.ObjNameIdList(constants.TypeWorkflow) },
		"yesNo":          util.YesNo,
	})

	// Load the view templates
	// If we're running from main, templates will come
	// from ./views. When running tests, templates come
	// from ../../views because http tests run from web
	// from ../../../views for member api and admin api
	// sub directory.
	if util.FileExists("./views") {
		router.LoadHTMLGlob("./views/**/*.html")
	} else if util.FileExists("./server/views") {
		router.LoadHTMLGlob("./server/views/**/*.html")
	} else if util.FileExists("../server/views") {
		router.LoadHTMLGlob("../server/views/**/*.html")
	} else {
		router.LoadHTMLGlob("../../server/views/**/*.html")
	}
}

func initRoutes(router *gin.Engine) {

	// This ensures that routes match even when they contain
	// extraneous slashes.
	router.RedirectFixedPath = true

	router.StaticFile("/favicon.ico", "./server/assets/img/favicon.ico")
	router.Static("/assets", "./server/assets")

	// About
	router.GET("/", controllers.DashboardShow)
	router.GET("/about", controllers.AboutShow)
	router.GET("/open_external", controllers.OpenExternalUrl)
	router.GET("/open_log", controllers.OpenLog)
	router.GET("/open_log_folder", controllers.OpenLogFolder)
	router.GET("/open_data_folder", controllers.OpenDataFolder)

	// App Settings
	router.GET("/app_settings", controllers.AppSettingIndex)
	router.GET("/app_settings/new", controllers.AppSettingNew)
	router.POST("/app_settings/new", controllers.AppSettingSave)
	router.GET("/app_settings/edit/:id", controllers.AppSettingEdit)
	router.PUT("/app_settings/edit/:id", controllers.AppSettingSave)
	router.POST("/app_settings/edit/:id", controllers.AppSettingSave)
	router.DELETE("/app_settings/delete/:id", controllers.AppSettingDelete)
	router.POST("/app_settings/delete/:id", controllers.AppSettingDelete)

	// BagIt Profiles
	router.GET("/profiles", controllers.BagItProfileIndex)
	router.GET("/profiles/new", controllers.BagItProfileNew)
	router.POST("/profiles/new", controllers.BagItProfileCreate)
	router.GET("/profiles/edit/:id", controllers.BagItProfileEdit)
	router.PUT("/profiles/edit/:id", controllers.BagItProfileSave)
	router.POST("/profiles/edit/:id", controllers.BagItProfileSave)
	router.PUT("/profiles/delete/:id", controllers.BagItProfileDelete)
	router.POST("/profiles/delete/:id", controllers.BagItProfileDelete)
	router.GET("/profiles/import", controllers.BagItProfileImportStart)
	router.POST("/profiles/import", controllers.BagItProfileImport)
	router.GET("/profiles/export/:id", controllers.BagItProfileExport)

	// BagIt Profile Tags & Tag Files
	router.GET("/profiles/new_tag/:profile_id/:tag_file", controllers.BagItProfileNewTag)
	router.GET("/profiles/edit_tag/:profile_id/:tag_id", controllers.BagItProfileEditTag)
	router.PUT("/profiles/edit_tag/:profile_id/:tag_id", controllers.BagItProfileSaveTag)
	router.POST("/profiles/edit_tag/:profile_id/:tag_id", controllers.BagItProfileSaveTag)
	router.POST("/profiles/delete_tag/:profile_id/:tag_id", controllers.BagItProfileDeleteTag)
	router.PUT("/profiles/delete_tag/:profile_id/:tag_id", controllers.BagItProfileDeleteTag)
	router.GET("/profiles/new_tag_file/:profile_id", controllers.BagItProfileNewTagFile)
	router.POST("/profiles/new_tag_file/:profile_id", controllers.BagItProfileCreateTagFile)
	router.POST("/profiles/delete_tag_file/:profile_id", controllers.BagItProfileDeleteTagFile)
	router.PUT("/profiles/delete_tag_file/:profile_id", controllers.BagItProfileDeleteTagFile)

	// Internal Settings
	router.GET("/internal_settings", controllers.InternalSettingIndex)

	// Jobs
	router.GET("/jobs", controllers.JobIndex)
	router.GET("/jobs/new", controllers.JobNew)
	router.PUT("/jobs/delete/:id", controllers.JobDelete)
	router.POST("/jobs/delete/:id", controllers.JobDelete)
	router.GET("/jobs/packaging/:id", controllers.JobShowPackaging)
	router.POST("/jobs/packaging/:id", controllers.JobSavePackaging)
	router.GET("/jobs/metadata/:id", controllers.JobShowMetadata)
	router.POST("/jobs/metadata/:id", controllers.JobSaveMetadata)
	router.GET("/jobs/add_tag/:id", controllers.JobAddTag)
	router.POST("/jobs/add_tag/:id", controllers.JobSaveTag)
	router.POST("/jobs/delete_tag/:id", controllers.JobDeleteTag)
	router.GET("/jobs/upload/:id", controllers.JobShowUpload)
	router.POST("/jobs/upload/:id", controllers.JobSaveUpload)
	router.GET("/jobs/files/:id", controllers.JobShowFiles)
	router.POST("/jobs/add_file/:id", controllers.JobAddFile)
	router.POST("/jobs/delete_file/:id", controllers.JobDeleteFile)
	router.GET("/jobs/summary/:id", controllers.JobRunShow)
	router.GET("/jobs/run/:id", controllers.JobRunExecute)

	// Remote Repositories
	router.GET("/remote_repositories", controllers.RemoteRepositoryIndex)
	router.GET("/remote_repositories/new", controllers.RemoteRepositoryNew)
	router.POST("/remote_repositories/new", controllers.RemoteRepositorySave)
	router.GET("/remote_repositories/edit/:id", controllers.RemoteRepositoryEdit)
	router.PUT("/remote_repositories/edit/:id", controllers.RemoteRepositorySave)
	router.POST("/remote_repositories/edit/:id", controllers.RemoteRepositorySave)
	router.PUT("/remote_repositories/delete/:id", controllers.RemoteRepositoryDelete)
	router.POST("/remote_repositories/delete/:id", controllers.RemoteRepositoryDelete)

	// Strorage Services
	router.GET("/storage_services", controllers.StorageServiceIndex)
	router.GET("/storage_services/new", controllers.StorageServiceNew)
	router.POST("/storage_services/new", controllers.StorageServiceSave)
	router.GET("/storage_services/edit/:id", controllers.StorageServiceEdit)
	router.PUT("/storage_services/edit/:id", controllers.StorageServiceSave)
	router.POST("/storage_services/edit/:id", controllers.StorageServiceSave)
	router.PUT("/storage_services/delete/:id", controllers.StorageServiceDelete)
	router.POST("/storage_services/delete/:id", controllers.StorageServiceDelete)
	router.POST("/storage_services/test/:id", controllers.StorageServiceTestConnection)

	// Workflows
	router.GET("/workflows", controllers.WorkflowIndex)
	router.GET("/workflows/new", controllers.WorkflowNew)
	router.GET("/workflows/edit/:id", controllers.WorkflowEdit)
	router.PUT("/workflows/edit/:id", controllers.WorkflowSave)
	router.POST("/workflows/edit/:id", controllers.WorkflowSave)
	router.GET("/workflows/export/:id", controllers.WorkflowExport)
	router.PUT("/workflows/delete/:id", controllers.WorkflowDelete)
	router.POST("/workflows/delete/:id", controllers.WorkflowDelete)
	router.POST("/workflows/from_job/:jobId", controllers.WorkflowCreateFromJob)
	router.POST("/workflows/run/:id", controllers.WorkflowRun)
	router.GET("/workflows/batch/choose", controllers.WorkflowShowBatchForm)
	router.POST("/workflows/batch/validate", controllers.WorkflowBatchValidate)
	router.POST("/workflows/batch/run", controllers.WorkflowRunBatch)
}
