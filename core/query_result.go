package core

type QueryResult struct {
	AppSetting         *AppSetting
	AppSettings        []*AppSetting
	InternalSetting    *InternalSetting
	InternalSettings   []*InternalSetting
	ObjType            string
	RemoteRepository   *RemoteRepository
	RemoteRepositories []*RemoteRepository
}

func NewQueryResult(objType string) *QueryResult {
	return &QueryResult{
		ObjType: objType,
	}
}
