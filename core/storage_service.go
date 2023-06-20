package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

type StorageService struct {
	ID             string            `json:"id"`
	AllowsDownload bool              `json:"allowsDownload"`
	AllowsUpload   bool              `json:"allowsUpload"`
	Bucket         string            `json:"bucket"`
	Description    string            `json:"description"`
	Errors         map[string]string `json:"-"`
	Host           string            `json:"host"`
	Login          string            `json:"login"`
	LoginExtra     string            `json:"loginExtra"`
	Name           string            `json:"name"`
	Password       string            `json:"password"`
	Port           int               `json:"port"`
	Protocol       string            `json:"protocol"`
}

func NewStorageService() *StorageService {
	return &StorageService{
		Errors: make(map[string]string),
	}
}

// URL returns the URL to which the file will be uploaded.
func (ss *StorageService) URL(filename string) string {
	return fmt.Sprintf("%s://%s/%s/%s", ss.Protocol, ss.HostAndPort(), ss.Bucket, filename)
}

// HostAndPort returns the host and port for connecting to a remote service.
// Use this when creating a connection to a remote S3 service.
func (ss *StorageService) HostAndPort() string {
	port := ""
	if ss.Port > 0 {
		port = fmt.Sprintf(":%d", ss.Port)
	}
	return fmt.Sprintf("%s%s", ss.Host, port)
}

func (ss *StorageService) Validate() bool {
	ss.Errors = make(map[string]string)
	if !util.LooksLikeUUID(ss.ID) {
		ss.Errors["StorageService.ID"] = "StorageService requires a valid ID."
	}
	if strings.TrimSpace(ss.Protocol) == "" {
		ss.Errors["StorageService.Protocol"] = "StorageService requires a protocol (s3, sftp, etc)."
	}
	if strings.TrimSpace(ss.Host) == "" {
		ss.Errors["StorageService.Host"] = "StorageService requires a hostname or IP address."
	}
	if strings.TrimSpace(ss.Bucket) == "" {
		ss.Errors["StorageService.Bucket"] = "StorageService requires a bucket or folder name."
	}
	if strings.TrimSpace(ss.Login) == "" {
		ss.Errors["StorageService.Login"] = "StorageService requires a login name or access key id."
	}
	if strings.TrimSpace(ss.Password) == "" {
		ss.Errors["StorageService.Password"] = "StorageService requires a password or secret access key."
	}
	return len(ss.Errors) == 0
}

// GetLogin returns the login name or AccessKeyID to connect to this
// storage service. Per the DART docts, if the login begins with "env:",
// we fetch it from the environment. For example, "env:MY_SS_LOGIN"
// causes us to fetch the env var "MY_SS_LOGIN". This allows us to
// copy Workflow info across the wire without exposing sensitive credentials.
//
// If the login does not begin with "env:", this returns it verbatim.
func (ss *StorageService) GetLogin() string {
	if strings.HasPrefix(ss.Login, "env:") {
		return ss.getEnv(ss.Login)
	}
	return ss.Login
}

// GetPassword returns this storage service's password from the
// StorageService record or from the environment as necessary. See the
// documentation for StorageService.GetLogin() for more info.
func (ss *StorageService) GetPassword() string {
	if strings.HasPrefix(ss.Password, "env:") {
		return ss.getEnv(ss.Password)
	}
	return ss.Password
}

// getEnv returns the value of an environment variable, minus the
// "ENV:" prefix.
func (ss *StorageService) getEnv(varname string) string {
	parts := strings.SplitN(varname, ":", 2)
	return os.Getenv(parts[1])
}

// Copy returns a pointer to a new StorageService whose values
// are the same as this service. The copy will have the same
// ID as the original, so if you want to change it, you'll have
// to do that yourself.
func (ss *StorageService) Copy() *StorageService {
	return &StorageService{
		ID:             ss.ID,
		AllowsDownload: ss.AllowsDownload,
		AllowsUpload:   ss.AllowsUpload,
		Bucket:         ss.Bucket,
		Description:    ss.Description,
		Errors:         ss.Errors,
		Host:           ss.Host,
		Login:          ss.Login,
		LoginExtra:     ss.LoginExtra,
		Name:           ss.Name,
		Password:       ss.Password,
		Port:           ss.Port,
		Protocol:       ss.Protocol,
	}
}

// StorageServiceFind returns the StorageService with the specified UUID,
// or sql.ErrNoRows if no matching record exists.
func StorageServiceFind(uuid string) (*StorageService, error) {
	result, err := ObjFind(uuid)
	if err != nil {
		return nil, err
	}
	return result.StorageService, err
}

// StorageServiceList returns a list of StorageServices with the specified
// order, offset and limit.
func StorageServiceList(orderBy string, limit, offset int) ([]*StorageService, error) {
	result, err := ObjList(constants.TypeStorageService, orderBy, limit, offset)
	if err != nil {
		return nil, err
	}
	return result.StorageServices, err
}

// ObjID returns this remote ss's UUID.
func (ss *StorageService) ObjID() string {
	return ss.ID
}

// ObjName returns the name of this remote ss.
func (ss *StorageService) ObjName() string {
	return ss.Name
}

// ObjType returns this object's type.
func (ss *StorageService) ObjType() string {
	return constants.TypeStorageService
}

func (ss *StorageService) String() string {
	return fmt.Sprintf("StorageService: '%s'", ss.Name)
}

// Save saves this ss, if it determines the ss is valid.
// It returns constants.ErrObjecValidation if the ss is invalid.
// Check ss.Errors if you get a validation error.
func (ss *StorageService) Save() error {
	if !ss.Validate() {
		return constants.ErrObjecValidation
	}
	return ObjSave(ss)
}

// Delete deletes this ss config.
func (ss *StorageService) Delete() error {
	return ObjDelete(ss.ID)
}

func (ss *StorageService) ToForm() *Form {
	form := NewForm(constants.TypeStorageService, ss.ID, ss.Errors)

	// TODO: Set protocol choices. Set choices for AllowsUpload/Download.

	form.AddField("ID", "ID", ss.ID, true)
	form.AddField("Name", "Name", ss.Name, true)
	form.AddField("Host", "Host", ss.Host, true)
	form.AddField("Port", "Port", strconv.Itoa(ss.Port), true)
	form.AddField("Protocol", "Protocol", ss.Protocol, true)
	form.AddField("Bucket", "Bucket", ss.LoginExtra, false)
	form.AddField("Description", "Description", ss.Description, false)
	form.AddField("Login", "Login", ss.Login, false)
	form.AddField("Password", "Password", ss.Password, false)
	form.AddField("LoginExtra", "Login Extra", ss.LoginExtra, false)
	form.AddField("AllowsUpload", "Allows Upload", strconv.FormatBool(ss.AllowsUpload), true)
	form.AddField("AllowsDownload", "Allows Download", strconv.FormatBool(ss.AllowsDownload), true)

	return form
}

func (ss *StorageService) GetErrors() map[string]string {
	return ss.Errors
}
