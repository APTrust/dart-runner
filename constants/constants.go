package constants

const (
	AlgMd5                   = "md5"
	AlgSha1                  = "sha1"
	AlgSha256                = "sha256"
	AlgSha512                = "sha512"
	BagItProfileBTR          = "btr-v1.0.json"
	BagItProfileDefault      = "aptrust-v2.2.json"
	BTRProfileIdentifier     = "https://github.com/dpscollaborative/btr_bagit_profile/releases/download/1.0/btr-bagit-profile.json"
	DefaultProfileIdentifier = "https://raw.githubusercontent.com/APTrust/preservation-services/master/profiles/aptrust-v2.2.json"
	EmptyUUID                = "00000000-0000-0000-0000-000000000000"
	ExitOK                   = 0
	ExitRuntimeErr           = 1
	ExitUsageErr             = 2
	FileTypeFetchTxt         = "fetch.txt"
	FileTypeManifest         = "manifest"
	FileTypePayload          = "payload file"
	FileTypeTag              = "tag file"
	FileTypeTagManifest      = "tag manifest"
	MaxS3ObjectSize          = int64(5497558138880) // 5TB
	MaxServerSideCopySize    = int64(5368709120)    // 5GB
	MaxValidationErrors      = 30
	PackageFormatBagIt       = "BagIt"
	ProtocolS3               = "s3"
	ProtocolSFTP             = "sftp"
	ResultTypeList           = "list"
	ResultTypeSingle         = "single"
	ResultTypeUnitialized    = "unintialized"
	SerializationForbidden   = "forbidden"
	SerializationOptional    = "optional"
	SerializationRequired    = "required"
	TypeAppSetting           = "AppSetting"
	TypeInternalSetting      = "InternalSetting"
	TypeRemoteRepository     = "RemoteRepository"
	TypeStorageService       = "StorageService"
)

var AcceptBagItVersion = []string{
	"0.97",
	"1.0",
}

var AcceptSerialization = []string{
	".tar",
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

// AppVersion is the version of DART Runner. This is set by
// the linker at compile time in main.Version, which then
// sets this var before it runs a job or workflow. This isn't
// really a constant, since it changes from build to build, but
// then didn't Heraclitus say "The only constant in life is change?"
var AppVersion string
