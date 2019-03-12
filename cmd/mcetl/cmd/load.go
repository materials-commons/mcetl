package cmd

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/materials-commons/mcetl/internal/spreadsheet"
	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Loads the given spreadsheet(s) and performs ETL.",
	Long: `The load command will read and process the given spreadsheets. The spreadsheets are processes
in the order given.`,
	Run: cliCmdLoad,
}

func init() {
	rootCmd.AddCommand(loadCmd)
	loadCmd.Flags().StringP("file", "f", "", "Path to the excel spreadsheet")
}

func cliCmdLoad(cmd *cobra.Command, args []string) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		fmt.Print("error", err)
	}

	xlsx, err := excelize.OpenFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	var processes []*spreadsheet.Process

	for index, name := range xlsx.GetSheetMap() {
		p := &spreadsheet.Process{
			Name:  name,
			Index: index,
		}
		processes = append(processes, p)
		fmt.Println(index, name)
		loadWorksheet(xlsx, p)
	}
}

func loadWorksheet(xlsx *excelize.File, p *spreadsheet.Process) {
	rows, err := xlsx.Rows(p.Name)
	if err != nil {
		fmt.Println("Rows returned error", err)
		return
	}

	for rows.Next() {
		for _, colCell := range rows.Columns() {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}

}
