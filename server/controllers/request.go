package controllers

import (
	"regexp"
	"strings"

	"github.com/APTrust/dart-runner/core"
	"github.com/gin-gonic/gin"
)

type Request struct {
	ginCtx         *gin.Context
	Action         string
	Errors         []error
	Handler        string
	ObjType        string
	Path           string
	PathAndQuery   string
	QueryResult    *core.QueryResult
	IsListResponse bool
	TemplateData   gin.H
}

// Regex to help us extract object type from handler name.
// E.g. AppSettingIndex -> AppSetting, StorageServiceEdit -> StorageService.
var routeSuffix = regexp.MustCompile(`Index|New|Save|Edit|Delete$`)

func NewRequest(c *gin.Context) *Request {
	pathAndQuery := c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		pathAndQuery = c.Request.URL.Path + "?" + c.Request.URL.RawQuery
	}
	request := &Request{
		ginCtx:         c,
		Errors:         make([]error, 0),
		IsListResponse: false,
		Path:           c.Request.URL.Path,
		PathAndQuery:   pathAndQuery,
		TemplateData: gin.H{
			"currentUrl":  c.Request.URL.Path,
			"showAsModal": c.Query("modal") == "true",
		},
	}
	request.initFromHandlerName()
	request.loadObjects()
	return request
}

func (r *Request) HasErrors() bool {
	return len(r.Errors) > 0
}

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
		r.TemplateData["objectExistsInDB"] = (r.QueryResult.ObjCount == 1)
		if r.QueryResult.Error != nil {
			r.Errors = append(r.Errors, r.QueryResult.Error)
			return
		}
		form, err := r.QueryResult.GetForm()
		if err != nil {
			r.Errors = append(r.Errors, err)
		} else {
			r.TemplateData["form"] = form
		}
	} else {
		orderBy := r.ginCtx.DefaultQuery("orderBy", "obj_name")
		offset := r.ginCtx.GetInt("offset")
		limit := r.ginCtx.GetInt("limit")
		if limit < 1 {
			limit = 25
		}
		r.QueryResult = core.ObjList(r.ObjType, orderBy, limit, offset)
		if r.QueryResult.Error != nil {
			r.Errors = append(r.Errors, r.QueryResult.Error)
			return
		}
		pager, err := NewPager(r.ginCtx, r.Path, 25)
		if err != nil {
			r.Errors = append(r.Errors, err)
			return
		}
		totalObjectCount, err := core.ObjCount(r.QueryResult.ObjType)
		if err != nil {
			r.Errors = append(r.Errors, err)
			return
		}
		pager.SetCounts(totalObjectCount, r.QueryResult.ObjCount)
		r.TemplateData["pager"] = pager
	}
}
