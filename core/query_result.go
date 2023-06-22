package core

import (
	"github.com/APTrust/dart-runner/constants"
)

type QueryResult struct {
	AppSettings        []*AppSetting
	Error              error
	InternalSettings   []*InternalSetting
	Limit              int
	ObjCount           int
	ObjType            string
	Offset             int
	OrderBy            string
	RemoteRepositories []*RemoteRepository
	ResultType         string
	StorageServices    []*StorageService
}

func NewQueryResult(resultType string) *QueryResult {
	qr := &QueryResult{
		ResultType: resultType,
	}
	return qr
}

func (qr *QueryResult) AppSetting() *AppSetting {
	if len(qr.AppSettings) == 1 {
		return qr.AppSettings[0]
	}
	return nil
}

func (qr *QueryResult) InternalSetting() *InternalSetting {
	if len(qr.InternalSettings) == 1 {
		return qr.InternalSettings[0]
	}
	return nil
}

func (qr *QueryResult) RemoteRepository() *RemoteRepository {
	if len(qr.RemoteRepositories) == 1 {
		return qr.RemoteRepositories[0]
	}
	return nil
}

func (qr *QueryResult) StorageService() *StorageService {
	if len(qr.StorageServices) == 1 {
		return qr.StorageServices[0]
	}
	return nil
}

func (qr *QueryResult) GetForm() (*Form, error) {
	if qr.ResultType != constants.ResultTypeSingle || qr.ObjCount != 1 {
		return nil, constants.ErrWrongTypeForForm
	}
	var form *Form
	var err error
	switch qr.ObjType {
	case constants.TypeAppSetting:
		form = qr.AppSetting().ToForm()
	case constants.TypeInternalSetting:
		form = qr.InternalSetting().ToForm()
	case constants.TypeStorageService:
		form = qr.StorageService().ToForm()
	case constants.TypeRemoteRepository:
		form = qr.RemoteRepository().ToForm()
	default:
		err = constants.ErrUnknownType
	}
	return form, err
}
