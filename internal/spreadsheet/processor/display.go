package processor

import (
	"fmt"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type Displayer struct{}

func NewDisplayer() *Displayer {
	return &Displayer{}
}

func (d *Displayer) Apply(processes []*model.Process) error {
	for _, process := range processes {
		fmt.Println("Process", process.Name)
		fmt.Println("   Process Attributes:")
		for _, pattr := range process.Attributes {
			fmt.Println("     ", pattr.Name)
		}
		fmt.Println("    Samples:")
		for _, sample := range process.Samples {
			fmt.Println("        ", sample.Name)
			fmt.Println("             Attributes:")
			for _, sattr := range sample.Attributes {
				fmt.Println("               ", sattr.Name, "/", sattr.Value)
			}
		}
	}

	return nil
}
