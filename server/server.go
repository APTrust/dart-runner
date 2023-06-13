package server

import (
	"io"
	"text/template"

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
		"dict": Dict,
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
	} else {
		router.LoadHTMLGlob("../server/views/**/*.html")
	}
}

func initRoutes(router *gin.Engine) {

	// This ensures that routes match even when they contain
	// extraneous slashes.
	router.RedirectFixedPath = true
	router.Static("/static", "./static")
	router.Static("/favicon.ico", "./static/img/favicon.png")

	// About
	router.GET("/", controllers.AboutShow)
	router.GET("/about", controllers.AboutShow)

	// App Settings
	router.GET("/app_settings", controllers.AppSettingIndex)
	router.GET("/app_settings/new", controllers.AppSettingNew)
	router.POST("/app_settings/new", controllers.AppSettingCreate)
	router.GET("/app_settings/edit/:id", controllers.AppSettingEdit)
	router.PUT("/app_settings/edit/:id", controllers.AppSettingUpdate)
	router.POST("/app_settings/edit/:id", controllers.AppSettingUpdate)
	router.PUT("/app_settings/delete/:id", controllers.AppSettingDelete)
	router.POST("/app_settings/delete/:id", controllers.AppSettingDelete)
	router.GET("/app_settings/:id", controllers.AppSettingShow)

	// BagIt Profiles
	router.GET("/profiles", controllers.ProfileIndex)
	router.GET("/profiles/new", controllers.ProfileNew)
	router.POST("/profiles/new", controllers.ProfileCreate)
	router.GET("/profiles/edit/:id", controllers.ProfileEdit)
	router.PUT("/profiles/edit/:id", controllers.ProfileUpdate)
	router.POST("/profiles/edit/:id", controllers.ProfileUpdate)
	router.PUT("/profiles/delete/:id", controllers.ProfileDelete)
	router.POST("/profiles/delete/:id", controllers.ProfileDelete)
	router.GET("/profiles/:id", controllers.ProfileShow)

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
	router.POST("/remote_repositories/new", controllers.RemoteRepositoryCreate)
	router.GET("/remote_repositories/edit/:id", controllers.RemoteRepositoryEdit)
	router.PUT("/remote_repositories/edit/:id", controllers.RemoteRepositoryUpdate)
	router.POST("/remote_repositories/edit/:id", controllers.RemoteRepositoryUpdate)
	router.PUT("/remote_repositories/delete/:id", controllers.RemoteRepositoryDelete)
	router.POST("/remote_repositories/delete/:id", controllers.RemoteRepositoryDelete)
	router.GET("/remote_repositories/:id", controllers.RemoteRepositoryShow)

	// Strorage Services
	router.GET("/storage_services", controllers.StorageServiceIndex)
	router.GET("/storage_services/new", controllers.StorageServiceNew)
	router.POST("/storage_services/new", controllers.StorageServiceCreate)
	router.GET("/storage_services/edit/:id", controllers.StorageServiceEdit)
	router.PUT("/storage_services/edit/:id", controllers.StorageServiceUpdate)
	router.POST("/storage_services/edit/:id", controllers.StorageServiceUpdate)
	router.PUT("/storage_services/delete/:id", controllers.StorageServiceDelete)
	router.POST("/storage_services/delete/:id", controllers.StorageServiceDelete)
	router.GET("/storage_services/:id", controllers.StorageServiceShow)

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
