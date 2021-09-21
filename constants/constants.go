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
	FileTypeFetchTxt         = "fetch.txt"
	FileTypeManifest         = "manifest"
	FileTypePayload          = "payload_file"
	FileTypeTag              = "tag_file"
	FileTypeTagManifest      = "tag_manifest"
	MaxS3ObjectSize          = int64(5497558138880) // 5TB
	MaxServerSideCopySize    = int64(5368709120)    // 5GB
	MaxValidationErrors      = 30
	SerializationForbidden   = "forbidden"
	SerializationOptional    = "optional"
	SerializationRequired    = "required"
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
