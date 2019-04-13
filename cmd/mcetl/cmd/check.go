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
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/materials-commons/mcetl/internal/spreadsheet"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the given spreadsheet(s) for errors and reports the errors. No ETL is performed.",
	Long: `The check command validates the given spreadsheets and reports any errors. It will not perform
any ETL operations on the spreadsheets.`,
	Run: cliCmdCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringP("file", "f", "", "Path to the excel spreadsheet")
	checkCmd.Flags().IntP("header-row", "r", 0, "Row to start reading from")
}

func cliCmdCheck(cmd *cobra.Command, args []string) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		fmt.Print("error", err)
		os.Exit(1)
	}

	headerRow, err := cmd.Flags().GetInt("header-row")
	if err != nil {
		fmt.Print("error", err)
		os.Exit(1)
	}

	processes, err := spreadsheet.Load(file, headerRow)
	if err != nil {
		fmt.Println("Loading spreadsheet failed")
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				fmt.Println(" ", e)
			}
		}
		os.Exit(1)
	}

	if err := spreadsheet.Display.Apply(processes); err != nil {
		fmt.Println("Unable to process spreadsheet:", err)
		os.Exit(1)
	}
}
