package server

import (
	"html/template"
	"io"

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

	router.SetFuncMap(template.FuncMap{
		"dict": util.Dict,
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
	router.GET("/", controllers.AboutShow)
	router.GET("/about", controllers.AboutShow)

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
	router.GET("/profiles/import_start", controllers.BagItProfileImport)
	router.POST("/profiles/import", controllers.BagItProfileImport)
	router.GET("/profiles/export/:id", controllers.BagItProfileExport)

	// BagIt Profile Tags & Tag Files
	router.GET("/profiles/new_tag/:profile_id/:tag_file", controllers.BagItProfileNewTag)
	router.POST("/profiles/new_tag/:profile_id", controllers.BagItProfileCreateTag)
	router.GET("/profiles/edit_tag/:profile_id/:tag_id", controllers.BagItProfileEditTag)
	router.PUT("/profiles/edit_tag/:profile_id/:tag_id", controllers.BagItProfileSaveTag)
	router.POST("/profiles/edit_tag/:profile_id/:tag_id", controllers.BagItProfileSaveTag)
	router.POST("/profiles/delete_tag/:profile_id/:tag_id", controllers.BagItProfileDeleteTag)
	router.PUT("/profiles/delete_tag/:profile_id/:tag_id", controllers.BagItProfileDeleteTag)
	router.POST("/profiles/new_tag_file/:profile_id", controllers.BagItProfileCreateTagFile)
	router.POST("/profiles/delete_tag_file/:profile_id", controllers.BagItProfileDeleteTagFile)
	router.PUT("/profiles/delete_tag_file/:profile_id", controllers.BagItProfileDeleteTagFile)

	// Internal Settings
	router.GET("/internal_settings", controllers.InternalSettingIndex)

	// Jobs
	router.GET("/jobs", controllers.JobIndex)
	router.GET("/jobs/new", controllers.JobNew)
	router.POST("/jobs/new", controllers.JobCreate)
	router.GET("/jobs/edit/:id", controllers.JobEdit)
	router.PUT("/jobs/edit/:id", controllers.JobUpdate)
	router.POST("/jobs/edit/:id", controllers.JobUpdate)
	router.PUT("/jobs/delete/:id", controllers.JobDelete)
	router.POST("/jobs/delete/:id", controllers.JobDelete)
	router.GET("/jobs/:id", controllers.JobShow)

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

	// Workflows
	router.GET("/workflows", controllers.WorkflowIndex)
	router.GET("/workflows/new", controllers.WorkflowNew)
	router.POST("/workflows/new", controllers.WorkflowCreate)
	router.GET("/workflows/edit/:id", controllers.WorkflowEdit)
	router.PUT("/workflows/edit/:id", controllers.WorkflowUpdate)
	router.POST("/workflows/edit/:id", controllers.WorkflowUpdate)
	router.PUT("/workflows/delete/:id", controllers.WorkflowDelete)
	router.POST("/workflows/delete/:id", controllers.WorkflowDelete)
	router.GET("/workflows/:id", controllers.WorkflowShow)

}
