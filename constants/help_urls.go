package constants

const BaseHelpUrl = "https://aptrust.github.io/dart-docs/"

// HelpUrlFor maps DART http handlers to context-specific help
// pages. There should be one entry here for each handler. If
// an entry is an empty string, we'll use the BaseHelpUrl.
var HelpUrlFor = map[string]string{
	"AboutShow":                      "",
	"AppSettingDelete":               "users/settings/app_settings/",
	"AppSettingEdit":                 "users/settings/app_settings/",
	"AppSettingIndex":                "users/settings/app_settings/",
	"AppSettingNew":                  "users/settings/app_settings/",
	"AppSettingSave":                 "users/settings/app_settings/",
	"BagItProfileCreate":             "users/bagit/creating/",
	"BagItProfileCreateTagFile":      "users/bagit/customizing/#adding-a-new-tag-file",
	"BagItProfileDelete":             "users/bagit/",
	"BagItProfileDeleteTag":          "users/bagit/customizing/#deleting-a-tag",
	"BagItProfileDeleteTagFile":      "users/bagit/customizing/#deleting-a-tag-file",
	"BagItProfileEdit":               "users/bagit/customizing/",
	"BagItProfileEditTag":            "users/bagit/customizing/#editing-a-tag",
	"BagItProfileExport":             "users/bagit/exporting/",
	"BagItProfileImport":             "users/bagit/importing/",
	"BagItProfileImportStart":        "users/bagit/importing/",
	"BagItProfileIndex":              "users/bagit/",
	"BagItProfileNew":                "users/bagit/creating/",
	"BagItProfileNewTag":             "users/bagit/customizing/#adding-a-new-tag-file",
	"BagItProfileNewTagFile":         "users/bagit/customizing/#adding-a-new-tag-file",
	"BagItProfileSave":               "users/bagit/customizing/",
	"BagItProfileSaveTag":            "users/bagit/customizing/#adding-a-tag",
	"DashboardGetReport":             "users/dashboard/",
	"DashboardShow":                  "users/dashboard/",
	"ExportSettingsIndex":            "users/settings/export/",
	"InternalSettingIndex":           "users/settings/internal_settings/",
	"JobAddFile":                     "users/jobs/files/",
	"JobAddTag":                      "users/jobs/metadata/#adding-custom-tags",
	"JobArtifactShow":                "", // page does not exist yet because artifacts are new
	"JobDelete":                      "users/jobs/delete/",
	"JobDeleteFile":                  "users/jobs/files/#removing-files",
	"JobDeleteTag":                   "users/jobs/metadata/#adding-custom-tags", // we need an actual delete section on this page
	"JobIndex":                       "users/jobs/list/",
	"JobNew":                         "users/jobs/", // needs section on how to create new job
	"JobRunExecute":                  "users/jobs/run/",
	"JobRunShow":                     "users/jobs/run/",
	"JobSaveMetadata":                "users/jobs/metadata/",
	"JobSavePackaging":               "users/jobs/packaging/",
	"JobSaveTag":                     "users/jobs/metadata/#adding-custom-tags",
	"JobSaveUpload":                  "users/jobs/upload/",
	"JobShowFiles":                   "users/jobs/files/",
	"JobShowMetadata":                "users/jobs/metadata/",
	"JobShowPackaging":               "users/jobs/packaging/",
	"JobShowUpload":                  "users/jobs/upload/",
	"OpenDataFolder":                 "",            // No help for this. Don't need it.
	"OpenExternalUrl":                "",            // No help for this. Don't need it.
	"OpenLog":                        "users/logs/", // Page needs update
	"OpenLogFolder":                  "users/logs/", // Page needs update
	"RemoteRepositoryDelete":         "users/settings/remote_repositories/",
	"RemoteRepositoryEdit":           "users/settings/remote_repositories/#editing-remote-repositories",
	"RemoteRepositoryIndex":          "users/settings/remote_repositories/",
	"RemoteRepositoryNew":            "users/settings/remote_repositories/",
	"RemoteRepositorySave":           "users/settings/remote_repositories/",
	"RemoteRepositoryTestConnection": "users/settings/remote_repositories/", // We need to add info to the page for this
	"SettingsExportDelete":           "users/settings/export/",
	"SettingsExportDeleteQuestion":   "users/settings/export/#export-questions",
	"SettingsExportEdit":             "users/settings/export/",
	"SettingsExportEditQuestion":     "users/settings/export/#export-questions",
	"SettingsExportNew":              "users/settings/export/",
	"SettingsExportNewQuestion":      "users/settings/export/#export-questions",
	"SettingsExportSave":             "users/settings/export/",
	"SettingsExportSaveQuestion":     "users/settings/export/#export-questions",
	"SettingsExportShowJson":         "users/settings/export/",
	"SettingsImportAnswers":          "users/settings/import/",
	"SettingsImportRun":              "users/settings/import/",
	"SettingsImportShow":             "users/settings/import/",
	"SettingsProfileTagList":         "", // Delete this endpoint? It doesn't seem to be used.
	"StorageServiceDelete":           "users/settings/storage_services/",
	"StorageServiceEdit":             "users/settings/storage_services/#editing-storage-services",
	"StorageServiceIndex":            "users/settings/storage_services/",
	"StorageServiceNew":              "users/settings/storage_services/",
	"StorageServiceSave":             "users/settings/storage_services/#editing-storage-services",
	"StorageServiceTestConnection":   "users/settings/storage_services/", // need to add documentation for this
	"UploadJobAddFile":               "",                                 // We need to add this page
	"UploadJobDeleteFile":            "",                                 // We need to add this page
	"UploadJobNew":                   "",                                 // We need to add this page
	"UploadJobReview":                "",                                 // We need to add this page
	"UploadJobRun":                   "",                                 // We need to add this page
	"UploadJobSaveTarget":            "",                                 // We need to add this page
	"UploadJobShowFiles":             "",                                 // We need to add this page
	"UploadJobShowTargets":           "",                                 // We need to add this page
	"ValidationJobAddFile":           "",                                 // We need to add this page
	"ValidationJobDeleteFile":        "",                                 // We need to add this page
	"ValidationJobNew":               "",                                 // We need to add this page
	"ValidationJobReview":            "",                                 // We need to add this page
	"ValidationJobRun":               "",                                 // We need to add this page
	"ValidationJobSaveProfile":       "",                                 // We need to add this page
	"ValidationJobShowFiles":         "",                                 // We need to add this page
	"ValidationJobShowProfiles":      "",                                 // We need to add this page
	"WorkflowBatchValidate":          "users/workflows/batch_jobs/",
	"WorkflowCreateFromJob":          "users/workflows/#creating-a-workflow-from-a-job",
	"WorkflowDelete":                 "users/workflows/",
	"WorkflowEdit":                   "users/workflows/#creating-a-workflow-from-scratch", // is this still possible in the new dart?
	"WorkflowExport":                 "",                                                  // is this still possible in the new dart? It should be.
	"WorkflowIndex":                  "users/workflows/",
	"WorkflowNew":                    "users/workflows/#creating-a-workflow-from-a-job",
	"WorkflowRun":                    "users/workflows/batch_jobs/", // is this still accurate? Probably not.
	"WorkflowRunBatch":               "users/workflows/batch_jobs/",
	"WorkflowSave":                   "users/workflows/", // is this still accurate?
	"WorkflowShowBatchForm":          "users/workflows/batch_jobs/",
}
