package constants

// Version value is injected at build time.
var Version string

const (
	AlgMd5                       = "md5"
	AlgSha1                      = "sha1"
	AlgSha256                    = "sha256"
	AlgSha512                    = "sha512"
	BaggingDirectory             = "Bagging Directory"
	BagItProfileBTR              = "btr-v1.0.json"
	BagItProfileDefault          = "aptrust-v2.2.json"
	BTRProfileIdentifier         = "https://github.com/dpscollaborative/btr_bagit_profile/releases/download/1.0/btr-bagit-profile.json"
	ControlCharactersInFileNames = "Control Characters in File Names"
	ControlCharFailValidation    = "Fail Validation"
	ControlCharIgnore            = "Ignore"
	ControlCharRefuseToBag       = "Refuse to Bag"
	ControlCharWarn              = "Warn"
	DefaultProfileIdentifier     = "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json"
	EmptyProfileID               = "73d1b307-4d6b-494b-b0c9-a8595222ae5a"
	EmptyProfileIdentifier       = "https://raw.githubusercontent.com/APTrust/dart/tree/master/profiles/empty_profile.json"
	EmptyUUID                    = "00000000-0000-0000-0000-000000000000"
	EventTypeBatchCompleted      = "batch completed"
	EventTypeDisconnect          = "disconnect"
	EventTypeError               = "error"
	EventTypeFinish              = "finish"
	EventTypeInfo                = "info"
	EventTypeInit                = "init"
	EventTypeStart               = "start"
	EventTypeWarning             = "warning"
	ExitOK                       = 0
	ExitRuntimeErr               = 1
	ExitUsageErr                 = 2
	FileTypeFetchTxt             = "fetch.txt"
	FileTypeJsonData             = "json data"
	FileTypeManifest             = "manifest"
	FileTypePayload              = "payload file"
	FileTypeTag                  = "tag file"
	FileTypeTagManifest          = "tag manifest"
	FlashCookieName              = "dart-flash-message"
	ImportSourceJson             = "json"
	ImportSourceUrl              = "url"
	ItemTypeFile                 = "file"
	ItemTypeJobResult            = "job result"
	ItemTypeManifest             = "manifest"
	ItemTypeTagFile              = "tag file"
	MaxS3ObjectSize              = int64(5497558138880) // 5TB
	MaxServerSideCopySize        = int64(5368709120)    // 5GB
	MaxValidationErrors          = 30
	ModeAptCmd                   = "apt-cmd"
	ModeDartGUI                  = "dart-gui"
	ModeDartRunner               = "dart-runner"
	PackageFormatBagIt           = "BagIt"
	PackageFormatNone            = "None" // Used when a job or workflow has no package operation.
	PluginIdAPTrustClientv3      = "c5a6b7db-5a5f-4ca5-a8f8-31b2e60c84bd"
	PluginIdLOCKSSClientv2       = "0dabdd1d-6227-4ad5-8a48-add1c699f8ab"
	PluginNameAPTrustClientv3    = "APTrust Registry Client (API Version 3)"
	PluginNameLOCKSSClientv2     = "LOCKSS Client (API Version 2)"
	ProfileIDAPTrust             = "043f1c22-c9ff-4112-86f8-8f8f1e6a2dca"
	ProfileIDBTR                 = "a4e95eae-9b93-4ebb-895e-d2ab23fd2c7c"
	ProfileIDEmpty               = "73d1b307-4d6b-494b-b0c9-a8595222ae5a"
	ProfileTypeDart              = "dart"
	ProfileTypeLOCOrdered        = "loc-ordered"
	ProfileTypeLOCUnordered      = "loc-unordered"
	ProfileTypeStandard          = "standard"
	ProfileTypeUnknown           = "unknown"
	ProtocolS3                   = "s3"
	ProtocolSFTP                 = "sftp"
	ResultTypeList               = "list"
	ResultTypeSingle             = "single"
	ResultTypeUnitialized        = "unintialized"
	SerializationForbidden       = "forbidden"
	SerializationOptional        = "optional"
	SerializationRequired        = "required"
	StageFinish                  = "finish"
	StagePackage                 = "package"
	StagePreRun                  = "pre-run"
	StageUpload                  = "upload"
	StageValidation              = "validation"
	StatusFailed                 = "failed"
	StatusRunning                = "running"
	StatusStarting               = "starting"
	StatusSuccess                = "success"
	TypeAppSetting               = "AppSetting"
	TypeBagItProfile             = "BagItProfile"
	TypeBagItProfileImport       = "BagItProfileImport"
	TypeExportQuestion           = "ExportQuestion"
	TypeExportSettings           = "ExportSettings"
	TypeInternalSetting          = "InternalSetting"
	TypeJob                      = "Job"
	TypeRemoteRepository         = "RemoteRepository"
	TypeStorageService           = "StorageService"
	TypeTagDefinition            = "TagDefinition"
	TypeUploadJob                = "UploadJob"
	TypeValidationJob            = "ValidationJob"
	TypeWorkflow                 = "Workflow"
	TypeWorkflowBatch            = "WorkflowBatch"
)

var AcceptBagItVersion = []string{
	"0.97",
	"1.0",
}

var AcceptSerialization = []string{
	"application/tar",
}

var SerializationOptions = []string{
	SerializationForbidden,
	SerializationOptional,
	SerializationRequired,
}

var PreferredAlgsInOrder = []string{
	AlgSha512,
	AlgSha256,
	AlgMd5,
	AlgSha1,
}

var AllItemTypes = []string{
	TypeAppSetting,
	TypeBagItProfile,
	TypeInternalSetting,
	TypeRemoteRepository,
	TypeStorageService,
	TypeTagDefinition,
}

var SavableItemTypes = []string{
	TypeAppSetting,
	TypeBagItProfile,
	TypeInternalSetting,
	TypeRemoteRepository,
	TypeStorageService,
}

var ExportableSettingTypes = []string{
	TypeAppSetting,
	TypeBagItProfile,
	TypeRemoteRepository,
	TypeStorageService,
}

// We have only one format at the moment, but in future we
// may add OCFL and others.
var PackageFormats = []string{
	PackageFormatBagIt,
}

// AppVersion is the version of DART Runner. This is set by
// the linker at compile time in main.Version, which then
// sets this var before it runs a job or workflow. This isn't
// really a constant, since it changes from build to build, but
// then didn't Heraclitus say "The only constant in life is change?"
var AppVersion string
