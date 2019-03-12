// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
