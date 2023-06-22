package core

func YesNoChoices(trueOrFalse bool) []Choice {
	return []Choice{
		{Label: "", Value: ""},
		{Label: "Yes", Value: "true", Selected: trueOrFalse},
		{Label: "No", Value: "false", Selected: !trueOrFalse},
	}
}
