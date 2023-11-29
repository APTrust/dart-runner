package core

import (
	"fmt"
	"strings"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/google/uuid"
)

type WorkflowBatch struct {
	ID            string
	Workflow      *Workflow
	PathToCSVFile string
	Errors        map[string]string
}

func NewWorkflowBatch(workflow *Workflow, pathToCSVFile string) *WorkflowBatch {
	return &WorkflowBatch{
		ID:            uuid.NewString(),
		Workflow:      workflow,
		PathToCSVFile: pathToCSVFile,
		Errors:        make(map[string]string),
	}
}

func (wb *WorkflowBatch) Validate() bool {
	wb.Errors = make(map[string]string)

	// Validate Workflow
	if wb.Workflow == nil {
		wb.Errors["WorkflowID"] = "Please choose a workflow."
	} else if !wb.Workflow.Validate() {
		for key, value := range wb.Workflow.Errors {
			wb.Errors["Workflow_"+key] = value
		}
	}

	// Make sure user selected a CSV file and that the file exists.
	if wb.PathToCSVFile == "" {
		wb.Errors["PathToCSVFile"] = "Please chose a CSV file."
	} else if !util.FileExists(wb.PathToCSVFile) {
		wb.Errors["PathToCSVFile"] = "CSV file does not exist."
	}

	if len(wb.Errors) > 0 {
		for key, value := range wb.Errors {
			Dart.Log.Errorf("%s: %s", key, value)
		}
		return false
	}

	// Now validate the contents of the file.
	return wb.validateCSVFile()
}

func (wb *WorkflowBatch) validateCSVFile() bool {
	// First, make the CSV file parses without errors.
	_, records, err := util.ParseCSV(wb.PathToCSVFile)
	if err != nil {
		wb.Errors["CSVFile"] = fmt.Sprintf("%s. Be sure this is a valid CSV file.", err.Error())
		return false
	}

	// Now validate the records, line by line...
	for i, record := range records {
		lineNumber := i + 1
		// Make sure we know which file or directory to bag for this line,
		// and make sure that file/dir actually exists.
		dirToBag, found := record.FirstMatching("Root-Directory")
		if found {
			if !util.FileExists(dirToBag.Value) {
				wb.Errors[dirToBag.Value] = fmt.Sprintf("Line %d: file or directory does not exist: '%s'.", lineNumber, dirToBag.Value)
			}
		} else {
			key := fmt.Sprintf("Line %d", lineNumber)
			wb.Errors[key] = fmt.Sprintf("Line %d: This entry is missing the 'Root-Directory' value, so DART does not know what to bag.", lineNumber)
		}
		// Lastly, make sure this line of the CSV file contains
		// valid values for all of the workflow's required tags.
		wb.checkRequiredTags(record, lineNumber)
	}

	for key, value := range wb.Errors {
		Dart.Log.Errorf("%s: %s", key, value)
	}

	return len(wb.Errors) == 0
}

func (wb *WorkflowBatch) checkRequiredTags(record *util.NameValuePairList, lineNumber int) bool {
	bagName, _ := record.FirstMatching("Bag-Name")
	if strings.TrimSpace(bagName.Value) == "" {
		errKey := fmt.Sprintf("%d-Bag-Name", lineNumber)
		wb.Errors[errKey] = fmt.Sprintf("Bag-Name is missing from line %d", lineNumber)
	}
	for _, tagDef := range wb.Workflow.BagItProfile.Tags {
		// We don't need to validate workflow tags in bagit.txt
		// because DART fills these in automatically.
		if tagDef.TagFile == "bagit.txt" {
			continue
		}
		// System-set tags like Bag-Date or Payload-Oxum may
		// be required by the profile, but we can't expect the
		// user to supply these values. The bagger calculates
		// them at runtime.
		if tagDef.SystemMustSet() {
			continue
		}
		// TODO: Do bag-info.txt tags ever not have a 'tagfile/' prefix?
		fullTagName := fmt.Sprintf("%s/%s", tagDef.TagFile, tagDef.TagName)
		errKey := fmt.Sprintf("%d-%s", lineNumber, fullTagName)
		tag, _ := record.FirstMatching(fullTagName)

		// 1. Make sure required tags have values.
		// 2. If tagDef has a non-empty .Values list, make sure the value
		//    we got from the CSV file is actually in that list.
		if strings.TrimSpace(tag.Value) == "" && tagDef.Required {
			wb.Errors[errKey] = fmt.Sprintf("Required tag %s on line %d is missing or empty.", fullTagName, lineNumber)
		} else if len(tagDef.Values) > 0 && !util.StringListContains(tagDef.Values, tag.Value) {
			wb.Errors[errKey] = fmt.Sprintf("Value %s for tag %s on line %d is not in the list of allowed values.", tag.Value, fullTagName, lineNumber)
		}
	}
	return true
}

func (wb *WorkflowBatch) ToForm() *Form {
	form := NewForm("WorkflowBatch", "ID not applicable to this type", wb.Errors)

	// Note that the form contains an upload field for the CSV file
	// instead of a string/text field for PathToCSVFile.
	csvUploadField := form.AddField("CsvUpload", "CSV Batch File", "", true)
	csvUploadField.Attrs["accept"] = ".csv"
	workflowID := ""
	if wb.Workflow != nil {
		workflowID = wb.Workflow.ID
	}
	workflowField := form.AddField("WorkflowID", "Choose a Workflow", workflowID, true)
	workflowField.Choices = ObjChoiceList(constants.TypeWorkflow, []string{workflowID})
	return form
}

// ---- PersistentObject Interface ----

// ObjID returns this items's object id (uuid).
func (wb *WorkflowBatch) ObjID() string {
	return wb.ID
}

// ObjName returns this object's name, so names will be
// searchable and sortable in the DB.
func (wb *WorkflowBatch) ObjName() string {
	return fmt.Sprintf("%s => %s", wb.Workflow.Name, wb.PathToCSVFile)

}

// ObjType returns this object's type name.
func (wb *WorkflowBatch) ObjType() string {
	return constants.TypeWorkflowBatch
}

func (wb *WorkflowBatch) String() string {
	return fmt.Sprintf("WorkflowBatch: %s => %s", wb.Workflow.Name, wb.PathToCSVFile)
}

func (wb *WorkflowBatch) GetErrors() map[string]string {
	return wb.Errors
}

func (wb *WorkflowBatch) IsDeletable() bool {
	return true
}
