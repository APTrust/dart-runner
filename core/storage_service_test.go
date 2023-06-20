package core_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/APTrust/dart-runner/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageService(t *testing.T) {
	ss := &core.StorageService{}
	assert.False(t, ss.Validate())
	assert.Equal(t, 6, len(ss.Errors))
	assert.Equal(t, "StorageService requires a valid ID.", ss.Errors["StorageService.ID"])
	assert.Equal(t, "StorageService requires a protocol (s3, sftp, etc).", ss.Errors["StorageService.Protocol"])
	assert.Equal(t, "StorageService requires a hostname or IP address.", ss.Errors["StorageService.Host"])
	assert.Equal(t, "StorageService requires a bucket or folder name.", ss.Errors["StorageService.Bucket"])
	assert.Equal(t, "StorageService requires a login name or access key id.", ss.Errors["StorageService.Login"])
	assert.Equal(t, "StorageService requires a password or secret access key.", ss.Errors["StorageService.Password"])

	ss = &core.StorageService{
		ID:       uuid.NewString(),
		Host:     "example.com",
		Bucket:   "uploads",
		Login:    "user@example.com",
		Password: "secret",
		Protocol: "s3",
	}
	assert.True(t, ss.Validate())
	assert.Empty(t, ss.Errors)
	assert.Equal(t, "s3://example.com/uploads/bag.tar", ss.URL("bag.tar"))
}

func TestStorageServiceSensitiveData(t *testing.T) {
	creds := map[string]string{
		"RUNNER_UNIT_TEST_SS_LOGIN": "user555@example.com",
		"RUNNER_UNIT_TEST_SS_PWD":   "Secret! Shh!!",
	}
	for key, value := range creds {
		os.Setenv(key, value)
	}
	defer func() {
		for key := range creds {
			os.Unsetenv(key)
		}
	}()

	ss := &core.StorageService{}
	ss.Login = "insecure@example.com"
	ss.Password = "not secret"

	assert.Equal(t, "insecure@example.com", ss.GetLogin())
	assert.Equal(t, "not secret", ss.GetPassword())

	ss.Login = "env:RUNNER_UNIT_TEST_SS_LOGIN"
	ss.Password = "env:RUNNER_UNIT_TEST_SS_PWD"

	assert.Equal(t, creds["RUNNER_UNIT_TEST_SS_LOGIN"], ss.GetLogin())
	assert.Equal(t, creds["RUNNER_UNIT_TEST_SS_PWD"], ss.GetPassword())
}

func TestStorageServiceURL(t *testing.T) {
	ss := &core.StorageService{}
	ss.Protocol = "s3"
	ss.Host = "example.com"
	ss.Bucket = "bucky"
	assert.Equal(t, "s3://example.com/bucky/document.pdf", ss.URL("document.pdf"))

	ss.Port = 9999
	assert.Equal(t, "s3://example.com:9999/bucky/document.pdf", ss.URL("document.pdf"))
}

func TestStorageServiceHostAndPort(t *testing.T) {
	ss := &core.StorageService{}
	ss.Host = "example.com"
	assert.Equal(t, "example.com", ss.HostAndPort())

	ss.Port = 9999
	assert.Equal(t, "example.com:9999", ss.HostAndPort())
}

func getSampleStorageService() *core.StorageService {
	return &core.StorageService{
		ID:             uuid.NewString(),
		AllowsDownload: true,
		AllowsUpload:   true,
		Bucket:         "chum.bucket",
		Description:    "Everyone loves the Krusty Krab!",
		Errors:         make(map[string]string),
		Host:           "s3.example.com",
		Login:          "user1",
		LoginExtra:     "nothing to see here",
		Name:           "example bucket",
		Password:       "secret-password",
		Port:           999,
		Protocol:       "s3",
	}
}

func TestStorgeServiceCopy(t *testing.T) {
	ss := getSampleStorageService()
	ssCopy := ss.Copy()
	require.NotNil(t, ssCopy)

	// Make sure internal struct values are equal...
	assert.Equal(t, ss, ssCopy)

	// ...but these pointers don't point to the same address
	assert.NotSame(t, ss, ssCopy)
}

func TestStorageServicePersistence(t *testing.T) {

	// Clean up when test completes.
	defer core.ClearDartTable()

	// Insert three records for testing.
	ss1 := getSampleStorageService()
	ss1.Name = "Storage Service 1"
	ss2 := ss1.Copy()
	ss2.ID = uuid.NewString()
	ss2.Name = "Storage Service 2"
	ss3 := ss1.Copy()
	ss3.ID = uuid.NewString()
	ss3.Name = "Storage Service 3"

	assert.Nil(t, core.ObjSave(ss1))
	assert.Nil(t, core.ObjSave(ss2))
	assert.Nil(t, core.ObjSave(ss3))

	// Make sure S1 was saved as expected.
	ss1Reload, err := core.StorageServiceFind(ss1.ID)
	require.Nil(t, err)
	require.NotNil(t, ss1Reload)
	assert.Equal(t, ss1.ID, ss1Reload.ID)
	assert.Equal(t, ss1.Name, ss1Reload.Name)
	assert.Equal(t, ss1.Host, ss1Reload.Host)

	// Make sure order, offset and limit work on list query.
	settings, err := core.StorageServiceList("obj_name", 1, 0)
	require.Nil(t, err)
	require.Equal(t, 1, len(settings))
	assert.Equal(t, ss1.ID, settings[0].ID)

	// Make sure we can get all results.
	settings, err = core.StorageServiceList("obj_name", 100, 0)
	require.Nil(t, err)
	require.Equal(t, 3, len(settings))
	assert.Equal(t, ss1.ID, settings[0].ID)
	assert.Equal(t, ss2.ID, settings[1].ID)
	assert.Equal(t, ss3.ID, settings[2].ID)

	// Make sure delete works. Should return no error.
	assert.Nil(t, core.ObjDelete(ss1))

	// Make sure the record was truly deleted.
	deletedRecord, err := core.AppSettingFind(ss1.ID)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, deletedRecord)
}
