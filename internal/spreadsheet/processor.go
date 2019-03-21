package spreadsheet

import (
	mcapi "github.com/materials-commons/gomcapi"
	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
	"github.com/materials-commons/mcetl/internal/spreadsheet/processor"
)

type Processor interface {
	Apply(processes []*model.Worksheet) error
}

var Display = processor.NewDisplayer()

func Create(projectID, name string, client *mcapi.Client) *processor.Creater {
	return processor.NewCreater(projectID, name, "", client)
}
