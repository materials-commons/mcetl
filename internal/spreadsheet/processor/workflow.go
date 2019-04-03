package processor

import "github.com/materials-commons/mcetl/internal/spreadsheet/model"

type Workflow struct{}

func NewWorkflow() *Workflow {
	return &Workflow{}
}

func (w *Workflow) Apply(worksheets []*model.Worksheet) error {
	return nil
}
