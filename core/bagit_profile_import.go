package core

import (
	"io"
	"net/http"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
)

// BagItProfileImport represents a URL or JSON blob to be imported
// as a DART BagIt profile.
type BagItProfileImport struct {
	ImportSource string
	URL          string
	JsonData     []byte
	Errors       map[string]string
}

// NewBagItProfileImport creates a new BagItProfileImport object.
// Param importSource should be either constants.ImportSourceUrl or
// constants.ImportSourceJson. If the source is a url, we'll fetch
// the JSON from the sourceUrl. Otherwise, we'll use jsonData.
// You only need to supply one or the other of the last two params.
func NewBagItProfileImport(importSource, sourceUrl string, jsonData []byte) *BagItProfileImport {
	return &BagItProfileImport{
		ImportSource: importSource,
		URL:          sourceUrl,
		JsonData:     jsonData,
		Errors:       make(map[string]string),
	}
}

// Convert converts the JSON data provided by the user to a DART-style
// BagIt profile.
func (bpi *BagItProfileImport) Convert() (*BagItProfile, error) {
	if !bpi.Validate() {
		return nil, constants.ErrObjecValidation
	}
	if bpi.ImportSource == constants.ImportSourceUrl {
		jsonData, err := bpi.getUrl()
		if err != nil {
			return nil, err
		}
		bpi.JsonData = jsonData
	}
	return ConvertProfile(bpi.JsonData, bpi.URL)
}

// Validate returns true if the object is valid, false if not.
// If this returns false, specific errors will be recorded in
// BagItProfileImport.Errors.
func (bpi *BagItProfileImport) Validate() bool {
	bpi.Errors = make(map[string]string)
	if bpi.ImportSource != constants.ImportSourceUrl && bpi.ImportSource != constants.ImportSourceJson {
		bpi.Errors["ImportSource"] = "Please specify either URL or JSON as the import source."
	}
	if bpi.ImportSource == constants.ImportSourceUrl && !util.LooksLikeURL(bpi.URL) {
		bpi.Errors["URL"] = "Please specify a valid URL."
	}
	if bpi.ImportSource == constants.ImportSourceJson && len(bpi.JsonData) < 2 {
		bpi.Errors["JsonData"] = "Please enter JSON to be imported."
	}
	return len(bpi.Errors) == 0
}

// ToForm returns a form for creating a BagItProfile import job.
func (bpi *BagItProfileImport) ToForm() *Form {
	form := NewForm(constants.TypeBagItProfileImport, constants.EmptyUUID, bpi.Errors)

	sourceField := form.AddField("ImportSource", "Source", bpi.ImportSource, true)
	sourceField.Choices = []Choice{
		{Label: "", Value: ""},
		{Label: "URL", Value: "URL"},
		{Label: "JSON", Value: "JsonData"},
	}

	urlField := form.AddField("URL", "URL", bpi.URL, false)
	urlField.Help = "The URL from which to import the BagIt profile."

	jsonDataField := form.AddField("JsonData", "Profile JSON", string(bpi.JsonData), false)
	jsonDataField.Help = "Paste the BagIt profile JSON you want to convert here."

	return form
}

func (bpi *BagItProfileImport) getUrl() ([]byte, error) {
	response, err := http.Get(bpi.URL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
