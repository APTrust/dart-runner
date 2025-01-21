package core_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageService(t *testing.T) {
	ss := &core.StorageService{}
	assert.False(t, ss.Validate())
	assert.Equal(t, 6, len(ss.Errors))
	assert.Equal(t, "StorageService requires a valid ID.", ss.Errors["ID"])
	assert.Equal(t, "StorageService requires a name.", ss.Errors["Name"])
	assert.Equal(t, "StorageService requires a protocol (s3, sftp, etc).", ss.Errors["Protocol"])
	assert.Equal(t, "StorageService requires a hostname or IP address.", ss.Errors["Host"])
	assert.Equal(t, "StorageService requires a login name or access key id.", ss.Errors["Login"])
	assert.Equal(t, "StorageService requires a password or secret access key, or the path to your SSH private key.", ss.Errors["Password"])

	// Bucket name is not required for SFTP
	ss.Protocol = constants.ProtocolSFTP
	ss.Validate()
	assert.Empty(t, ss.Errors["Bucket"])

	// But it is for S3
	ss.Protocol = constants.ProtocolS3
	ss.Validate()
	assert.Equal(t, "StorageService requires a bucket name when the protocol is S3.", ss.Errors["Bucket"])

	ss = &core.StorageService{
		ID:       uuid.NewString(),
		Name:     "Spongebob",
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
	result := core.ObjFind(ss1.ID)
	require.Nil(t, result.Error)
	ss1Reload := result.StorageService()
	require.NotNil(t, ss1Reload)
	assert.Equal(t, ss1.ID, ss1Reload.ID)
	assert.Equal(t, ss1.Name, ss1Reload.Name)
	assert.Equal(t, ss1.Host, ss1Reload.Host)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeStorageService, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.StorageServices))
	assert.Equal(t, ss1.ID, result.StorageServices[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeStorageService, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.StorageServices))
	assert.Equal(t, ss1.ID, result.StorageServices[0].ID)
	assert.Equal(t, ss2.ID, result.StorageServices[1].ID)
	assert.Equal(t, ss3.ID, result.StorageServices[2].ID)

	// Make sure delete works. Should return no error.
	assert.Nil(t, core.ObjDelete(ss1))

	// Make sure the record was truly deleted.
	result = core.ObjFind(ss1.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)
	assert.Nil(t, result.StorageService())
}

func TestStorageServiceToForm(t *testing.T) {
	ss := core.NewStorageService()
	ss.Name = "test service"
	ss.AllowsUpload = true
	ss.AllowsDownload = true
	ss.Bucket = "the chum bucket"
	ss.Description = "yadda yadda yadda"
	ss.Host = "sftp.example.com"
	ss.Login = "spongebob"
	ss.LoginExtra = "login-xtra"
	ss.Password = "patrick star"
	ss.Port = 8080
	ss.Protocol = constants.ProtocolSFTP

	form := ss.ToForm()
	assert.Equal(t, 12, len(form.Fields))
	assert.True(t, form.UserCanDelete)
	assert.Equal(t, ss.ID, form.Fields["ID"].Value)
	assert.Equal(t, ss.Name, form.Fields["Name"].Value)
	assert.Equal(t, "true", form.Fields["AllowsUpload"].Value)
	assert.Equal(t, "true", form.Fields["AllowsDownload"].Value)
	assert.Equal(t, ss.Bucket, form.Fields["Bucket"].Value)
	assert.Equal(t, ss.Description, form.Fields["Description"].Value)
	assert.Equal(t, ss.Host, form.Fields["Host"].Value)
	assert.Equal(t, ss.Login, form.Fields["Login"].Value)
	assert.Equal(t, ss.LoginExtra, form.Fields["LoginExtra"].Value)
	assert.Equal(t, ss.Password, form.Fields["Password"].Value)
	assert.Equal(t, "8080", form.Fields["Port"].Value)
	assert.Equal(t, ss.Protocol, form.Fields["Protocol"].Value)

	assert.True(t, form.Fields["ID"].Required)
	assert.True(t, form.Fields["Name"].Required)
	assert.True(t, form.Fields["Protocol"].Required)
	assert.True(t, form.Fields["Host"].Required)
	assert.True(t, form.Fields["Bucket"].Required)
	assert.True(t, form.Fields["Login"].Required)
	assert.True(t, form.Fields["Password"].Required)
}

func TestStorageServicePersistentObject(t *testing.T) {
	ss := core.NewStorageService()
	ss.Name = "test repo"

	assert.Equal(t, constants.TypeStorageService, ss.ObjType())
	assert.Equal(t, "StorageService", ss.ObjType())
	assert.Equal(t, ss.ID, ss.ObjID())
	assert.True(t, util.LooksLikeUUID(ss.ObjID()))
	assert.True(t, ss.IsDeletable())
	assert.Equal(t, "test repo", ss.ObjName())
	assert.Equal(t, "StorageService: 'test repo'", ss.String())
	assert.Empty(t, ss.GetErrors())

	ss.Errors = map[string]string{
		"Error 1": "Message 1",
		"Error 2": "Message 2",
	}

	assert.Equal(t, 2, len(ss.GetErrors()))
	assert.Equal(t, "Message 1", ss.GetErrors()["Error 1"])
	assert.Equal(t, "Message 2", ss.GetErrors()["Error 2"])
}

func TestStorageServiceConnectionS3(t *testing.T) {
	// If you're running this test without using ./scripts/test.rb,
	// start the local minio server with the following command first.
	// You would run this from the dart-runner project root directory.
	ss, err := core.LoadStorageServiceFixture("storage_service_local_minio.json")
	require.NoError(t, err)
	assert.NoError(t, ss.TestConnection())
}

func TestHasPlaintextPassword(t *testing.T) {
	ss := &core.StorageService{}

	// False because password is empty
	assert.False(t, ss.HasPlaintextPassword())

	// False because password is empty
	ss.Password = "    "
	assert.False(t, ss.HasPlaintextPassword())

	// False because password uses env variable
	ss.Password = "env:SS_PASSWORD"
	assert.False(t, ss.HasPlaintextPassword())

	// True because password is non-empty, and
	// not an ENV variable.
	ss.Password = "this-here-is-secret"
	assert.True(t, ss.HasPlaintextPassword())
}
