package core

type CSVBatchParser struct {
	PathToCSVFile string
	Workflow      *Workflow
}

func NewCSVBatchParser(pathToCSVFile string, workflow *Workflow) *CSVBatchParser {
	return &CSVBatchParser{
		PathToCSVFile: pathToCSVFile,
		Workflow:      workflow,
	}
}

func (p *CSVBatchParser) ParseAll() ([]*JobParams, error) {
	jobParamsList := make([]*JobParams, 0)
	// headers, records, err := util.ParseCSV(p.PathToCSVFile)
	// if err != nil {
	// 	return nil, err
	// }

	return jobParamsList, nil
}

func (p *CSVBatchParser) parseTags() {

}
