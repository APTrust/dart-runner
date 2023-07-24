package core

import "github.com/APTrust/dart-runner/constants"

func YesNoChoices(trueOrFalse bool) []Choice {
	return []Choice{
		{Label: "", Value: ""},
		{Label: "Yes", Value: "true", Selected: trueOrFalse},
		{Label: "No", Value: "false", Selected: !trueOrFalse},
	}
}

func MakeChoiceList(options []string, selectedValue string) []Choice {
	choices := make([]Choice, len(options)+1)
	choices[0] = Choice{Label: "", Value: ""}
	for i, item := range options {
		selected := item == selectedValue
		choices[i+1] = Choice{Label: item, Value: item, Selected: selected}
	}
	return choices
}

func MakeMultiChoiceList(options []string, selectedValues []string) []Choice {
	choices := make([]Choice, len(options)+1)
	choices[0] = Choice{Label: "", Value: ""}
	for i, item := range options {
		selected := false
		for _, selectedValue := range selectedValues {
			if item == selectedValue {
				selected = true
				break
			}
		}
		choices[i+1] = Choice{Label: item, Value: item, Selected: selected}
	}
	return choices
}

func BagItProfileChoiceList(selectedValue string) []Choice {
	result := ObjList(constants.TypeBagItProfile, "obj_name", 1000, 0)
	choices := make([]Choice, len(result.BagItProfiles)+1)
	choices[0] = Choice{Label: "", Value: ""}
	for i, profile := range result.BagItProfiles {
		selected := profile.ID == selectedValue
		choices[i+1] = Choice{Label: profile.Name, Value: profile.ID, Selected: selected}
	}
	return choices
}
