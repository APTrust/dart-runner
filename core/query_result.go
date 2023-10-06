package core

import (
	"github.com/APTrust/dart-runner/constants"
)

type QueryResult struct {
	AppSettings        []*AppSetting
	BagItProfiles      []*BagItProfile
	Error              error
	InternalSettings   []*InternalSetting
	Jobs               []*Job
	Limit              int
	ObjCount           int
	ObjType            string
	Offset             int
	OrderBy            string
	RemoteRepositories []*RemoteRepository
	ResultType         string
	StorageServices    []*StorageService
	Workflows          []*Workflow
	WorkflowBatches    []*WorkflowBatch
}

func NewQueryResult(resultType string) *QueryResult {
	qr := &QueryResult{
		ResultType: resultType,
	}
	return qr
}

func (qr *QueryResult) AppSetting() *AppSetting {
	if len(qr.AppSettings) > 0 {
		return qr.AppSettings[0]
	}
	return nil
}

func (qr *QueryResult) BagItProfile() *BagItProfile {
	if len(qr.BagItProfiles) > 0 {
		return qr.BagItProfiles[0]
	}
	return nil
}

func (qr *QueryResult) InternalSetting() *InternalSetting {
	if len(qr.InternalSettings) > 0 {
		return qr.InternalSettings[0]
	}
	return nil
}

func (qr *QueryResult) Job() *Job {
	if len(qr.Jobs) > 0 {
		return qr.Jobs[0]
	}
	return nil
}

func (qr *QueryResult) RemoteRepository() *RemoteRepository {
	if len(qr.RemoteRepositories) > 0 {
		return qr.RemoteRepositories[0]
	}
	return nil
}

func (qr *QueryResult) StorageService() *StorageService {
	if len(qr.StorageServices) > 0 {
		return qr.StorageServices[0]
	}
	return nil
}

func (qr *QueryResult) Workflow() *Workflow {
	if len(qr.Workflows) > 0 {
		return qr.Workflows[0]
	}
	return nil
}

func (qr *QueryResult) WorkflowBatch() *WorkflowBatch {
	if len(qr.WorkflowBatches) > 0 {
		return qr.WorkflowBatches[0]
	}
	return nil
}

func (qr *QueryResult) GetForm() (*Form, error) {
	if qr.ResultType != constants.ResultTypeSingle || qr.ObjCount < 1 {
		return nil, constants.ErrWrongTypeForForm
	}
	var form *Form
	var err error
	switch qr.ObjType {
	case constants.TypeAppSetting:
		form = qr.AppSetting().ToForm()
	case constants.TypeBagItProfile:
		form = qr.BagItProfile().ToForm()
	case constants.TypeInternalSetting:
		form = qr.InternalSetting().ToForm()
	case constants.TypeRemoteRepository:
		form = qr.RemoteRepository().ToForm()
	case constants.TypeStorageService:
		form = qr.StorageService().ToForm()
	case constants.TypeWorkflow:
		form = qr.Workflow().ToForm()
	case constants.TypeWorkflowBatch:
		form = qr.WorkflowBatch().ToForm()
	default:
		err = constants.ErrUnknownType
	}
	return form, err
}
