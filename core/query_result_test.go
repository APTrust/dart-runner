package core_test

import (
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryResult(t *testing.T) {
	qr := core.NewQueryResult(constants.ResultTypeList)
	assert.Equal(t, constants.ResultTypeList, qr.ResultType)
}

func TestQueryResultTypeGetters(t *testing.T) {
	// AppSetting
	qr := core.NewQueryResult(constants.ResultTypeSingle)
	qr.ObjType = constants.TypeAppSetting
	assert.Nil(t, qr.AppSetting())
	assert.Nil(t, qr.InternalSetting())
	assert.Nil(t, qr.RemoteRepository())
	assert.Nil(t, qr.StorageService())

	qr.AppSettings = []*core.AppSetting{
		{ID: uuid.NewString(), Name: "Setting 1", Value: "Value 1"},
	}
	require.NotNil(t, qr.AppSetting())
	assert.Equal(t, qr.AppSettings[0].ID, qr.AppSetting().ID)

	qr.AppSettings = append(qr.AppSettings, &core.AppSetting{ID: uuid.NewString(), Name: "Setting 2", Value: "Value 2"})
	require.NotNil(t, qr.AppSetting())
	assert.Equal(t, qr.AppSettings[0].ID, qr.AppSetting().ID)

	// InternalSetting
	qr = core.NewQueryResult(constants.ResultTypeSingle)
	qr.ObjType = constants.TypeInternalSetting
	assert.Nil(t, qr.AppSetting())
	assert.Nil(t, qr.InternalSetting())
	assert.Nil(t, qr.RemoteRepository())
	assert.Nil(t, qr.StorageService())

	qr.InternalSettings = []*core.InternalSetting{
		{ID: uuid.NewString(), Name: "Setting 1", Value: "Value 1"},
	}
	require.NotNil(t, qr.InternalSetting())
	assert.Equal(t, qr.InternalSettings[0].ID, qr.InternalSetting().ID)

	qr.InternalSettings = append(qr.InternalSettings, &core.InternalSetting{ID: uuid.NewString(), Name: "Setting 2", Value: "Value 2"})
	require.NotNil(t, qr.InternalSetting())
	assert.Equal(t, qr.InternalSettings[0].ID, qr.InternalSetting().ID)

	// RemoteRepository
	qr = core.NewQueryResult(constants.ResultTypeSingle)
	qr.ObjType = constants.TypeRemoteRepository
	assert.Nil(t, qr.AppSetting())
	assert.Nil(t, qr.InternalSetting())
	assert.Nil(t, qr.RemoteRepository())
	assert.Nil(t, qr.StorageService())

	qr.RemoteRepositories = []*core.RemoteRepository{
		core.NewRemoteRepository(),
	}
	require.NotNil(t, qr.RemoteRepository())
	assert.Equal(t, qr.RemoteRepositories[0].ID, qr.RemoteRepository().ID)

	qr.RemoteRepositories = append(qr.RemoteRepositories, core.NewRemoteRepository())
	require.NotNil(t, qr.RemoteRepository())
	assert.Equal(t, qr.RemoteRepositories[0].ID, qr.RemoteRepository().ID)

	// StorageService
	qr = core.NewQueryResult(constants.ResultTypeSingle)
	qr.ObjType = constants.TypeStorageService
	assert.Nil(t, qr.AppSetting())
	assert.Nil(t, qr.InternalSetting())
	assert.Nil(t, qr.RemoteRepository())
	assert.Nil(t, qr.StorageService())

	qr.StorageServices = []*core.StorageService{
		core.NewStorageService(),
	}
	require.NotNil(t, qr.StorageService())
	assert.Equal(t, qr.StorageServices[0].ID, qr.StorageService().ID)

	qr.StorageServices = append(qr.StorageServices, core.NewStorageService())
	require.NotNil(t, qr.StorageService())
	assert.Equal(t, qr.StorageServices[0].ID, qr.StorageService().ID)
}

func TestQueryResultToForm(t *testing.T) {
	qr := core.NewQueryResult(constants.ResultTypeSingle)
	qr.ObjCount = 1
	qr.AppSettings = []*core.AppSetting{
		{ID: uuid.NewString(), Name: "Setting 1", Value: "Value 1"},
	}
	qr.InternalSettings = []*core.InternalSetting{
		{ID: uuid.NewString(), Name: "Setting 1", Value: "Value 1"},
	}
	qr.RemoteRepositories = []*core.RemoteRepository{
		{ID: uuid.NewString(), Name: "Test Repo 1"},
	}
	qr.StorageServices = []*core.StorageService{
		{ID: uuid.NewString(), Name: "Storage Service 1"},
	}

	qr.ObjType = constants.TypeAppSetting
	form, err := qr.GetForm()
	require.Nil(t, err)
	require.NotNil(t, form)
	assert.Equal(t, qr.AppSetting().ID, form.Fields["ID"].Value)
	assert.Equal(t, qr.AppSetting().Name, form.Fields["Name"].Value)

	qr.ObjType = constants.TypeInternalSetting
	form, err = qr.GetForm()
	require.Nil(t, err)
	require.NotNil(t, form)
	assert.Equal(t, qr.InternalSetting().ID, form.Fields["ID"].Value)
	assert.Equal(t, qr.InternalSetting().Name, form.Fields["Name"].Value)

	qr.ObjType = constants.TypeRemoteRepository
	form, err = qr.GetForm()
	require.Nil(t, err)
	require.NotNil(t, form)
	assert.Equal(t, qr.RemoteRepository().ID, form.Fields["ID"].Value)
	assert.Equal(t, qr.RemoteRepository().Name, form.Fields["Name"].Value)

	qr.ObjType = constants.TypeStorageService
	form, err = qr.GetForm()
	require.Nil(t, err)
	require.NotNil(t, form)
	assert.Equal(t, qr.StorageService().ID, form.Fields["ID"].Value)
	assert.Equal(t, qr.StorageService().Name, form.Fields["Name"].Value)

}
