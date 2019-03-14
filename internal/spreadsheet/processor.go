package spreadsheet

import (
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
	"github.com/materials-commons/mcetl/internal/spreadsheet/processor"
)

type Processor interface {
	Apply(processes []*model.Process) error
}

var Display = processor.NewDisplayer()

func Create(projectID, name string) *processor.Creater {
	return processor.NewCreater(projectID, name, "")
}
