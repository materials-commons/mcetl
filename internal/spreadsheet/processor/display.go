package processor

import (
	"fmt"
	"strings"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"
)

type Displayer struct{}

func NewDisplayer() *Displayer {
	return &Displayer{}
}

func (d *Displayer) Apply(processes []*model.Process) error {
	for _, process := range processes {
		fmt.Println("Process", process.Name)
		fmt.Printf("%sProcess Attributes:\n", spaces(4))
		for _, pattr := range process.Attributes {
			fmt.Printf("%s%s\n", spaces(6), pattr.Name)
		}
		fmt.Printf("%sSamples:\n", spaces(4))
		for _, sample := range process.Samples {
			fmt.Printf("%s%s\n", spaces(6), sample.Name)
			fmt.Printf("%sAttributes:\n", spaces(8))
			for _, sattr := range sample.Attributes {
				fmt.Printf("%s%s: %s(%s)\n", spaces(10), sattr.Name, sattr.Value, sattr.Unit)
			}
		}
		fmt.Println("")
	}

	return nil
}

func spaces(count int) string {
	return strings.Repeat(" ", count)
}
