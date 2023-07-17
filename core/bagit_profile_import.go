package core

import (
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

func NewBagItProfileImport(importSource, sourceUrl string, jsonData []byte) *BagItProfileImport {
	return &BagItProfileImport{
		ImportSource: importSource,
		URL:          sourceUrl,
		JsonData:     jsonData,
		Errors:       make(map[string]string),
	}
}

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
	return len(bpi.Errors) > 0
}

func (bpi *BagItProfileImport) ToForm() *Form {
	form := NewForm(constants.TypeBagItProfileImport, constants.EmptyUUID, bpi.Errors)

	sourceField := form.AddField("Import Source", "Source", bpi.ImportSource, true)
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
	jsonData := make([]byte, 0)
	defer response.Body.Close()
	_, err = response.Body.Read(jsonData)
	return jsonData, err
}
