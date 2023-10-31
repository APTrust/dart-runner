package core

import (
	"fmt"

	"github.com/APTrust/dart-runner/constants"
	"github.com/google/uuid"
)

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
	Field  string            `json:"field"`
	Errors map[string]string `json:"-"`
}

// NewExportQuestion returns a new ExportQuestion with a unique ID.
func NewExportQuestion() *ExportQuestion {
	return &ExportQuestion{
		ID:     uuid.NewString(),
		Errors: make(map[string]string),
	}
}

func (q *ExportQuestion) ToForm() *Form {
	form := NewForm(constants.TypeExportQuestion, q.ID, q.Errors)
	form.UserCanDelete = true

	// Note that we override control id because there
	// can be multiple questions on one form,
	// and the controls will have duplicate IDs if we stick
	// with the ObjType_FieldName pattern.

	// Also note that we use data-control-id so the front-end
	// JavaScript knows which controls are logically grouped
	// together. We use data-control-name to attach events to
	// select lists, and to know which select lists to update.

	// To understand how the front-end JS uses these attributes,
	// see views/settings/question_form.html.

	idField := form.AddField("ID", "ID", q.ID, true)
	idField.ID = fmt.Sprintf("id-%s", q.ID)

	promptField := form.AddField("Prompt", "Prompt", q.Prompt, true)
	promptField.ID = fmt.Sprintf("prompt-%s", q.ID)
	promptField.Help = "Enter the text of the question here."
	promptField.Attrs["data-question-id"] = q.ID
	promptField.Attrs["data-control-name"] = "prompt"

	objTypeField := form.AddField("ObjType", "Setting Type", q.ObjType, true)
	objTypeField.ID = fmt.Sprintf("objType-%s", q.ID)
	objTypeField.Help = "Copy the user's answer to this type of object."
	objTypeField.Choices = MakeChoiceList(constants.ExportableSettingTypes, "")
	objTypeField.Attrs["data-question-id"] = q.ID
	objTypeField.Attrs["data-control-name"] = "objType"

	objIDField := form.AddField("ObjID", "Setting Name", q.ObjID, true)
	objIDField.ID = fmt.Sprintf("objId-%s", q.ID)
	objIDField.Help = "Copy the user's answer to this specific object."
	objIDField.Choices = ObjChoiceList(q.ObjType, []string{q.ObjID})
	objIDField.Attrs["data-question-id"] = q.ID
	objIDField.Attrs["data-control-name"] = "objId"

	opts := NewExportOptions()

	fieldField := form.AddField("Field", "Field", q.Field, false)
	fieldField.ID = fmt.Sprintf("field-%s", q.ID)
	fieldField.Help = "Copy the user's answer to this property or tag."
	fieldField.Attrs["data-question-id"] = q.ID
	fieldField.Attrs["data-control-name"] = "field"

	switch q.ObjType {
	case constants.TypeAppSetting:
		fieldField.Choices = MakeChoiceList(opts.AppSettingFields, q.Field)
	case constants.TypeBagItProfile:
		pairs, err := TagsForProfile(q.ObjID)
		if err == nil {
			fieldField.Choices = MakeChoiceListFromPairs(pairs, q.Field)
		}
	case constants.TypeRemoteRepository:
		fieldField.Choices = MakeChoiceList(opts.RemoteRepositoryFields, q.Field)
	case constants.TypeStorageService:
		fieldField.Choices = MakeChoiceList(opts.StorageServiceFields, q.Field)
	}

	return form
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
