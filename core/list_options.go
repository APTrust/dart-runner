package core

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

func MakeChoiceListFromPairs(options []NameIDPair, selectedValue string) []Choice {
	choices := make([]Choice, len(options)+1)
	choices[0] = Choice{Label: "", Value: ""}
	for i, item := range options {
		selected := item.ID == selectedValue
		choices[i+1] = Choice{Label: item.Name, Value: item.ID, Selected: selected}
	}
	return choices
}
