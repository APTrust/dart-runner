package core

import (
	"github.com/APTrust/dart-runner/constants"
)

type QueryResult struct {
	AppSettings        []*AppSetting
	BagItProfiles      []*BagItProfile
	Error              error
	ExportSettings     []*ExportSettings
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
	UploadJobs         []*UploadJob
	ValidationJobs     []*ValidationJob
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

func (qr *QueryResult) ExportSetting() *ExportSettings {
	if len(qr.ExportSettings) > 0 {
		return qr.ExportSettings[0]
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

func (qr *QueryResult) UploadJob() *UploadJob {
	if len(qr.UploadJobs) > 0 {
		return qr.UploadJobs[0]
	}
	return nil
}

func (qr *QueryResult) ValidationJob() *ValidationJob {
	if len(qr.ValidationJobs) > 0 {
		return qr.ValidationJobs[0]
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

// Result count returns the total number of records
// returned by the query. This is different from ObjectCount,
// which is the total number of objects of a specified type
// in the DB.
func (qr *QueryResult) ResultCount() int {
	count := 0
	switch qr.ObjType {
	case constants.TypeAppSetting:
		count = len(qr.AppSettings)
	case constants.TypeBagItProfile:
		count = len(qr.BagItProfiles)
	case constants.TypeExportSettings:
		count = len(qr.ExportSettings)
	case constants.TypeInternalSetting:
		count = len(qr.InternalSettings)
	case constants.TypeJob:
		count = len(qr.Jobs)
	case constants.TypeRemoteRepository:
		count = len(qr.RemoteRepositories)
	case constants.TypeStorageService:
		count = len(qr.StorageServices)
	case constants.TypeUploadJob:
		count = len(qr.UploadJobs)
	case constants.TypeValidationJob:
		count = len(qr.ValidationJobs)
	case constants.TypeWorkflow:
		count = len(qr.Workflows)
	case constants.TypeWorkflowBatch:
		count = len(qr.WorkflowBatches)
	}
	return count
}

func (qr *QueryResult) GetForm() (*Form, error) {
	if qr.ResultType != constants.ResultTypeSingle || qr.ObjCount < 1 {
		return nil, constants.ErrWrongTypeForForm
	}
	var form *Form
	var err error

	// Note that there's no form for type Job
	// because Job form is actually multiple forms.
	switch qr.ObjType {
	case constants.TypeAppSetting:
		form = qr.AppSetting().ToForm()
	case constants.TypeExportSettings:
		form = qr.ExportSetting().ToForm()
	case constants.TypeBagItProfile:
		form = qr.BagItProfile().ToForm()
	case constants.TypeInternalSetting:
		form = qr.InternalSetting().ToForm()
	case constants.TypeRemoteRepository:
		form = qr.RemoteRepository().ToForm()
	case constants.TypeStorageService:
		form = qr.StorageService().ToForm()
	case constants.TypeUploadJob:
		form = qr.UploadJob().ToForm()
	case constants.TypeValidationJob:
		form = qr.ValidationJob().ToForm()
	case constants.TypeWorkflow:
		form = qr.Workflow().ToForm()
	case constants.TypeWorkflowBatch:
		form = qr.WorkflowBatch().ToForm()
	default:
		err = constants.ErrUnknownType
	}
	return form, err
}
