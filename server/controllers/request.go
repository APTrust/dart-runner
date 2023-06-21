package controllers

import (
	"regexp"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

type Request struct {
	ginCtx         *gin.Context
	Handler        string
	ObjType        string
	Action         string
	QueryResult    *core.QueryResult
	IsListResponse bool
	TemplateData   gin.H
}

// Regex to help us extract object type from handler name.
// E.g. AppSettingIndex -> AppSetting, StorageServiceEdit -> StorageService.
var routeSuffix = regexp.MustCompile(`Index|New|Save|Edit|Delete$`)

func NewRequest(c *gin.Context) *Request {
	request := &Request{
		ginCtx:         c,
		IsListResponse: false,
		TemplateData: gin.H{
			"currentUrl":  c.Request.URL,
			"showAsModal": c.Query("modal") == "true",
		},
	}
	request.initFromHandlerName()
	request.loadObjects()

	// TODO: Init pager if this is a list request

	// TODO: Figure out response template?
	//       Run save & delete here
	//       Run ToForm here
	//       Set error on request object and figure out status code?

	return request
}

/*

TODO:

  * Set ObjType (AppSetting, RemoteRepo, etc)
  * Set Action (new, edit, save, list, delete)
  * Select template
  * Fetch object or list
  * Populate template data
  * Set up pager (if this is a list page)
  * Figure out redirects and status code
  * Set current menu highlight
  * Set flash message
  * If single item, set template var to indicate whether ObjExists(), so we can show or hide delete button

  * Index, New, Save, Edit, Delete

*/

func (r *Request) initFromHandlerName() {
	nameParts := strings.Split(r.ginCtx.HandlerName(), ".")
	if len(nameParts) > 1 {
		r.Handler = nameParts[len(nameParts)-1]
		if strings.HasSuffix(r.Handler, "Index") {
			r.IsListResponse = true
		}
		r.ObjType = routeSuffix.ReplaceAllString(r.Handler, "")
		r.Action = strings.Replace(r.Handler, r.ObjType, "", 1)
	}
}

func (r *Request) loadObjects() {
	objId := r.ginCtx.Param("id")
	if !r.IsListResponse {
		r.QueryResult = core.ObjFind(objId)
		r.TemplateData["objectExistsInDB"], _ = core.ObjExists(objId)
	} else {
		orderBy := r.ginCtx.DefaultQuery("orderBy", "obj_name")
		offset := r.ginCtx.GetInt("offset")
		limit := r.ginCtx.GetInt("limit")
		if limit < 1 {
			limit = 25
		}
		r.QueryResult = core.ObjList(r.ObjType, orderBy, limit, offset)
	}
}
