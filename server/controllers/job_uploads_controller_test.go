package controllers_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/stretchr/testify/require"
)

type JobUploadTestInfo struct {
	Job             *core.Job
	Services        []*core.StorageService
	ServiceNames    []string
	ServiceIds      []string
	ExpectedContent []string
}

func TestJobShowUpload(t *testing.T) {
	defer core.ClearDartTable()
	info := GetJobUploadTestInfo(t)
	DoSimpleGetTest(t, fmt.Sprintf("/jobs/upload/%s", info.Job.ID), info.ExpectedContent)
}

func TestJobSaveUpload(t *testing.T) {

}

func TestGetUploadTargetsForm(t *testing.T) {

}

func TestAlreadySelectedTargets(t *testing.T) {

}

func GetJobUploadTestInfo(t *testing.T) JobUploadTestInfo {
	job := loadTestJob(t)

	serviceNames := make([]string, 0)
	serviceIds := make([]string, 0)

	// Make sure our DB contains this job's storage services.
	for _, op := range job.UploadOps {
		if op.StorageService != nil {
			require.NoError(t, core.ObjSave(op.StorageService))
			serviceIds = append(serviceIds, op.StorageService.ID)
			serviceNames = append(serviceNames, op.StorageService.Name)
		}
	}
	require.NoError(t, core.ObjSave(job))

	// Add some more services to the DB
	services := CreateStorageServices(t, 5)
	for _, ss := range services {
		serviceIds = append(serviceIds, ss.ID)
		serviceNames = append(serviceNames, ss.Name)
	}

	expected := append(serviceIds, serviceNames...)

	return JobUploadTestInfo{
		Job:             job,
		Services:        services,
		ServiceNames:    serviceNames,
		ServiceIds:      serviceIds,
		ExpectedContent: expected,
	}
}

func CreateStorageServices(t *testing.T, count int) []*core.StorageService {
	services := make([]*core.StorageService, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("Service %d", i+1)
		host := fmt.Sprintf("service-%d.example.com", i+1)
		ss := getFakeService(name, host)
		require.NoError(t, core.ObjSave(ss))
		services[i] = ss
	}
	return services
}
