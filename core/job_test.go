package core_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/core"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestJob(t *testing.T) *core.Job {
	workflow := getTestWorkflow(t)
	files := getTestFileList()
	tags := getTestTags()
	params := core.NewJobParams(workflow, "job_unit_test.tar", os.TempDir(), files, tags)
	return params.ToJob()
}

func TestJobValidate(t *testing.T) {
	job := core.NewJob()
	assert.False(t, job.Validate())
	//fmt.Println(job.Errors)
	require.Equal(t, 5, len(job.Errors))
	assert.Equal(t, "Validation requires a BagItProfile.", job.Errors["Job.Validate.BagItProfile"])
	assert.Equal(t, "Validation requires a file or bag to validate.", job.Errors["Job.Validate.PathToBag"])
	assert.Equal(t, "Output path is required.", job.Errors["PackageOperation.OutputPath"])
	assert.Equal(t, "Package name is required.", job.Errors["PackageOperation.PackageName"])
	assert.Equal(t, "Specify at least one file or directory to package.", job.Errors["PackageOperation.SourceFiles"])

	//assert.Equal(t, "Job has nothing to package, validate, or upload.", job.Errors["Job"])

	job = getTestJob(t)
	assert.Equal(t, 0, len(job.Errors))

	// BagItJob without Profile should cause error
	job.BagItProfile = nil
	assert.False(t, job.Validate())

	assert.Equal(t, 2, len(job.Errors))
	assert.Equal(t, "BagIt packaging requires a BagItProfile.", job.Errors["Job.Package.BagItProfile"])
	assert.Equal(t, "Validation requires a BagItProfile.", job.Errors["Job.Validate.BagItProfile"])

	// Let's cause some more problems, shall we?
	for _, uploadOp := range job.UploadOps {
		uploadOp.StorageService.Login = ""
	}
	assert.False(t, job.Validate())
	assert.Equal(t, 3, len(job.Errors))
	assert.Equal(t, "StorageService requires a login name or access key id.", job.Errors["Job.StorageService.Login"])

	for _, uploadOp := range job.UploadOps {
		uploadOp.StorageService = nil
	}
	assert.False(t, job.Validate())
	assert.Equal(t, 4, len(job.Errors))
	assert.Equal(t, "UploadOperation requires a StorageService", job.Errors["UploadOp[0].StorageService"])
	assert.Equal(t, "UploadOperation requires a StorageService", job.Errors["UploadOp[1].StorageService"])
}

func TestJobFromJson(t *testing.T) {
	pathToJobFile := filepath.Join(util.PathToTestData(), "files", "sample_job.json")
	job, err := core.JobFromJson(pathToJobFile)
	require.Nil(t, err)
	require.NotNil(t, job)

	assert.True(t, job.HasPackageOp())
	assert.True(t, job.HasUploadOps())

	job.PackageOp = nil
	job.UploadOps = make([]*core.UploadOperation, 0)

	assert.False(t, job.HasPackageOp())
	assert.False(t, job.HasUploadOps())
}

func TestJobPersistence(t *testing.T) {
	// Clean up when test completes.
	defer core.ClearDartTable()

	// Insert three records for testing.
	job1 := getTestJob(t)
	job1.ID = uuid.NewString()
	job1.PackageOp.PackageName = "Job1.tar"
	job2 := getTestJob(t)
	job2.ID = uuid.NewString()
	job2.PackageOp.PackageName = "Job2.tar"
	job3 := getTestJob(t)
	job3.ID = uuid.NewString()
	job3.PackageOp.PackageName = "Job3.tar"
	assert.Nil(t, core.ObjSave(job1))
	assert.Nil(t, core.ObjSave(job2))
	assert.Nil(t, core.ObjSave(job3))

	// Make sure S1 was saved as expected.
	result := core.ObjFind(job1.ID)
	require.Nil(t, result.Error)
	job1Reload := result.Job()
	require.NotNil(t, job1Reload)
	assert.Equal(t, job1.ID, job1Reload.ID)
	assert.Equal(t, job1.Name(), job1Reload.Name())
	assert.Equal(t, job1.BagItProfile.BagItProfileInfo.BagItProfileIdentifier, job1Reload.BagItProfile.BagItProfileInfo.BagItProfileIdentifier)

	// Make sure order, offset and limit work on list query.
	result = core.ObjList(constants.TypeJob, "obj_name", 1, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 1, len(result.Jobs))
	assert.Equal(t, job1.ID, result.Jobs[0].ID)

	// Make sure we can get all results.
	result = core.ObjList(constants.TypeJob, "obj_name", 100, 0)
	require.Nil(t, result.Error)
	require.Equal(t, 3, len(result.Jobs))
	assert.Equal(t, job1.ID, result.Jobs[0].ID)
	assert.Equal(t, job2.ID, result.Jobs[1].ID)
	assert.Equal(t, job3.ID, result.Jobs[2].ID)

	// Make sure delete works. Should return no error.
	assert.NoError(t, core.ObjDelete(job1))

	// Make sure the record was truly deleted.
	result = core.ObjFind(job1.ID)
	assert.Equal(t, sql.ErrNoRows, result.Error)
	assert.Nil(t, result.Job())
}

func TestJobFromWorkflow(t *testing.T) {
	defer core.ClearDartTable()
	workflow := loadJsonWorkflow(t)
	assert.True(t, workflow.Validate())
	assert.Empty(t, workflow.Errors)
	assert.NotNil(t, workflow.BagItProfile)

	job := core.JobFromWorkflow(workflow)
	require.NotNil(t, job)
	assert.True(t, util.LooksLikeUUID(job.ID))
	assert.Equal(t, workflow.ID, job.WorkflowID)
	assert.NotNil(t, job.PackageOp)
	assert.NotNil(t, job.ValidationOp)
	assert.NotNil(t, job.BagItProfile)
	assert.Equal(t, workflow.BagItProfile.ID, job.BagItProfile.ID)
	assert.Equal(t, len(workflow.StorageServices), len(job.UploadOps))
	assert.Equal(t, workflow.PackageFormat, job.PackageOp.PackageFormat)
}

func TestJobToForm(t *testing.T) {
	defer core.ClearDartTable()
	job := getTestJob(t)
	form := job.ToForm()
	assert.NotNil(t, form)

	assert.Equal(t, constants.TypeJob, form.ObjType)
	assert.Equal(t, job.ID, form.ObjectID)
	assert.True(t, form.UserCanDelete)

	assert.Equal(t, 6, len(form.Fields))
	assert.Equal(t, job.ID, form.Fields["ID"].Value)
	assert.Equal(t, job.BagItProfile.ID, form.Fields["BagItProfileID"].Value)
	assert.Equal(t, constants.PackageFormatBagIt, form.Fields["PackageFormat"].Value)
	assert.Equal(t, ".tar", form.Fields["BagItSerialization"].Value)
	assert.Equal(t, "job_unit_test.tar", form.Fields["PackageName"].Value)

	expectedOutputPath := filepath.Join(os.TempDir(), job.PackageOp.PackageName)
	assert.Equal(t, expectedOutputPath, form.Fields["OutputPath"].Value)

	// Test some specifics of OutputPath
	baggingDir := filepath.Join("home", "someone", "dart")
	setting := core.NewAppSetting("Bagging Directory", baggingDir)
	assert.NoError(t, core.ObjSave(setting))

	// If no output path is specified, and we have a package name,
	// the form should automatically set the output path to
	// bagging dir + package name.
	job.PackageOp.OutputPath = ""
	form = job.ToForm()
	expectedOutputPath = filepath.Join(baggingDir, "job_unit_test.tar")
	assert.Equal(t, expectedOutputPath, form.Fields["OutputPath"].Value)

	// This following emulates a common case in which the user
	// runs a job as an instance of a workflow. When we convert
	// the workflow to a job, it has no package name, path to bag,
	// or source files. In this case, we want to be sure that the
	// output path on the form includes a trailing slash (or backslash
	// on Windows), which vastly eases some path parsing and logic
	// on the front end, where a JavaScript function tries to sync
	// the bag name and output path.
	//
	// First, strip out any information that Job may use to find
	// the bag name and output path.
	job.PackageOp.PackageName = ""
	job.PackageOp.OutputPath = ""
	job.ValidationOp.PathToBag = ""
	for i, _ := range job.UploadOps {
		op := job.UploadOps[i]
		op.SourceFiles = make([]string, 0)
	}
	for i, _ := range job.BagItProfile.Tags {
		tagDef := job.BagItProfile.Tags[i]
		tagDef.DefaultValue = ""
		tagDef.UserValue = ""
	}

	// Now rebuild the form and test our output path.
	// It should have the trailing slash.
	form = job.ToForm()
	expectedOutputPath = filepath.Join("home", "someone", "dart") + string(os.PathSeparator)
	assert.Equal(t, expectedOutputPath, form.Fields["OutputPath"].Value)
}

func TestJobPackageFormat(t *testing.T) {
	job := core.NewJob()
	job.PackageOp = nil

	// No package operation = format none
	assert.Equal(t, constants.PackageFormatNone, job.PackageFormat())

	// If pacakage op doesn't specify format, format = none
	job.PackageOp = &core.PackageOperation{}
	assert.Empty(t, job.PackageOp.PackageFormat)
	assert.Equal(t, constants.PackageFormatNone, job.PackageFormat())

	// If job has a package op with an explicitly defined format,
	// we should get that from job.PackageFormat()
	job.PackageOp.PackageFormat = constants.PackageFormatBagIt
	assert.Equal(t, constants.PackageFormatBagIt, job.PackageFormat())
}

func TestJobOutcome(t *testing.T) {
	job := core.NewJob()

	outcome := job.Outcome()
	assert.Equal(t, job.Name(), outcome.JobName)
	assert.False(t, outcome.JobWasRun)
	assert.False(t, outcome.JobSucceeded)
	assert.Equal(t, "Job has not run", outcome.Message)
	assert.Equal(t, constants.StagePreRun, outcome.Stage)
	assert.Empty(t, outcome.LastActivity)

	// Mark the job's package operation as completed
	// with no errors.
	job.PackageOp.Result.Start()
	job.PackageOp.Result.Finish(nil)
	outcome = job.Outcome()
	assert.True(t, outcome.JobWasRun)
	assert.True(t, outcome.JobSucceeded)
	assert.Equal(t, "Packaging succeeded", outcome.Message)
	assert.Equal(t, constants.StagePackage, outcome.Stage)
	assert.Equal(t, job.PackageOp.Result.Completed, outcome.LastActivity)

	// Package attempted but failed
	errors := map[string]string{
		"Oops!": "I did it again",
	}
	job.PackageOp.Result.Finish(errors)
	outcome = job.Outcome()
	assert.True(t, outcome.JobWasRun)
	assert.False(t, outcome.JobSucceeded)
	assert.Equal(t, "Packaging failed", outcome.Message)
	assert.Equal(t, constants.StagePackage, outcome.Stage)

	// Validation succeeded
	job.ValidationOp.Result.Start()
	job.ValidationOp.Result.Finish(nil)
	outcome = job.Outcome()
	assert.True(t, outcome.JobWasRun)
	assert.True(t, outcome.JobSucceeded)
	assert.Equal(t, "Validation succeeded", outcome.Message)
	assert.Equal(t, constants.StageValidation, outcome.Stage)
	assert.Equal(t, job.ValidationOp.Result.Completed, outcome.LastActivity)

	// Validation failed
	job.ValidationOp.Result.Finish(errors)
	outcome = job.Outcome()
	assert.True(t, outcome.JobWasRun)
	assert.False(t, outcome.JobSucceeded)
	assert.Equal(t, "Validation failed", outcome.Message)
	assert.Equal(t, constants.StageValidation, outcome.Stage)

	// Upload succeeded
	job.UploadOps = make([]*core.UploadOperation, 1)
	ss := core.NewStorageService()
	files := make([]string, 0)
	job.UploadOps[0] = core.NewUploadOperation(ss, files)
	job.UploadOps[0].Result.Start()
	job.UploadOps[0].Result.Finish(nil)
	outcome = job.Outcome()
	assert.True(t, outcome.JobWasRun)
	assert.True(t, outcome.JobSucceeded)
	assert.Equal(t, "Upload succeeded", outcome.Message)
	assert.Equal(t, constants.StageFinish, outcome.Stage)
	assert.Equal(t, job.UploadOps[0].Result.Completed, outcome.LastActivity)

	assert.Empty(t, outcome.FailedUploads)
	require.Equal(t, 1, len(outcome.SuccessfulUploads))
	assert.Equal(t, job.UploadOps[0].StorageService.Name, outcome.SuccessfulUploads[0])

	// Upload failed
	job.UploadOps[0].Result.Finish(errors)
	outcome = job.Outcome()
	assert.True(t, outcome.JobWasRun)
	assert.False(t, outcome.JobSucceeded)
	assert.Equal(t, "One or more uploads failed", outcome.Message)
	assert.Equal(t, constants.StageUpload, outcome.Stage)

	assert.Empty(t, outcome.SuccessfulUploads)
	require.Equal(t, 1, len(outcome.FailedUploads))
	assert.Equal(t, job.UploadOps[0].StorageService.Name, outcome.FailedUploads[0])

}
