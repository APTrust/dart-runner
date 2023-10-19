package core

import "github.com/google/uuid"

// ExportQuestion is a question included in export settings
// to help a user set field values on settings objects.
// The user's response to an ExportQuestion will be copied
// to the specified field on the object with the specified ID.
//
// For example, the following question would copy the user's
// email address into the Login field of the StorageService
// with ID eb95dd3c-65c7-44ff-a84f-0231ebd36afc:
//
// {
//    Prompt: "What is your email address?",
//    ObjType: constants.TypeStorageService,
//    ObjID: "eb95dd3c-65c7-44ff-a84f-0231ebd36afc",
//    Field: "Login",
// }
//
// This would copy the user's response into the
// Source-Organization tag of the BagIt profile with ID
// 11111111-1111-1111-1111-111111111111, where the
// Source-Organization tag has ID 99999999-9999-9999-9999-999999999999:
//
// {
//    Prompt: "What is the name of your organization?",
//    ObjType: constants.TypeBagItProfile,
//    ObjID: "11111111-1111-1111-1111-111111111111",
//    Field: "99999999-9999-9999-9999-999999999999",
// }

type ExportQuestion struct {
	// ID is the question's unique uuid.
	ID string `json:"id"`
	// Prompt is the text of the question.
	Prompt string `json:"prompt"`
	// ObjType is the type of object to which we should
	// copy this question's response.
	ObjType string `json:"objType"`
	// ObjID is the id of the object to which we should
	// copy this question's response.
	ObjID string `json:"objId"`
	// Field is the name of the field on the object to
	// which we should copy this question's response.
	// For BagIt profiles, Field will contain the UUID
	// of the tag to which we should copy the response.
	Field string `json:"field"`
}

// NewExportQuestion returns a new ExportQuestion with a unique ID.
func NewExportQuestion() *ExportQuestion {
	return &ExportQuestion{
		ID: uuid.NewString(),
	}
}

// GetExistingValue returns the value currently stored
// in the object property to which this question refers.
// It will return a string, even if the value is of some
// other type. We use this to pre-populate answers in
// ExportQuestions.
func (q *ExportQuestion) GetExistingValue() string {

	return ""
}

// CopyResponseToField copies the user's response to field
// Field on the object specified by ExportQuestion.ObjID.
// If the object is a BagIt profile, this copies the response
// the tag whose UUID matches the value in ExportQuestion.Field.
func (q *ExportQuestion) CopyResponseToField(response string) error {

	return nil
}

func (q *ExportQuestion) getAppSetting() (*AppSetting, error) {

	return nil, nil
}

func (q *ExportQuestion) getBagItProfile() (*BagItProfile, error) {

	return nil, nil
}

func (q *ExportQuestion) getRemoteRepo() (*RemoteRepository, error) {

	return nil, nil
}

func (q *ExportQuestion) getStorageService() (*StorageService, error) {

	return nil, nil
}
