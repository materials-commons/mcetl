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
		for _, sample := range process.Samples {
			fmt.Printf("%s associated with sample %s\n", spaces(6), sample.Name)
			for _, pattr := range sample.ProcessAttrs {
				showAttr(10, pattr)
			}
		}
		fmt.Printf("%sSamples:\n", spaces(4))
		for _, sample := range process.Samples {
			fmt.Printf("%s%s\n", spaces(6), sample.Name)
			fmt.Printf("%sAttributes:\n", spaces(8))
			for _, sattr := range sample.Attributes {
				showAttr(10, sattr)
			}
		}
		fmt.Println("")
	}

	return nil
}

func showAttr(numberOfSpaces int, attr *model.Attribute) {
	unit := "(No units given)"
	if attr.Unit != "" {
		unit = fmt.Sprintf("(%s)", attr.Unit)
	}
	if len(attr.Value) != 0 {
		fmt.Printf("%s%s: %s %s\n", spaces(numberOfSpaces), attr.Name, attr.Value["value"], unit)
	} else {
		fmt.Printf("%s%s: %s %s\n", spaces(numberOfSpaces), attr.Name, "No value given", unit)
	}
}

func spaces(count int) string {
	return strings.Repeat(" ", count)
}
