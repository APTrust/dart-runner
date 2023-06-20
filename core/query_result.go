package core

type QueryResult struct {
	AppSetting         *AppSetting
	AppSettings        []*AppSetting
	Error              error
	InternalSetting    *InternalSetting
	InternalSettings   []*InternalSetting
	Limit              int
	ObjCount           int
	ObjType            string
	Offset             int
	OrderBy            string
	RemoteRepository   *RemoteRepository
	RemoteRepositories []*RemoteRepository
	StorageService     *StorageService
	StorageServices    []*StorageService
}

func NewQueryResult(objType string) *QueryResult {
	return &QueryResult{
		ObjType: objType,
	}
}
