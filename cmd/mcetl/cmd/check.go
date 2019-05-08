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
	"strings"

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
	checkCmd.Flags().StringP("files", "f", "", "Path to the excel spreadsheet")
	checkCmd.Flags().IntP("header-row", "r", 0, "Row to start reading from")
	checkCmd.Flags().BoolP("has-parent", "t", false, "2nd column is the parent column")
	checkCmd.Flags().StringP("project-id", "p", "", "Project to create experiment in")
	checkCmd.Flags().StringP("mcurl", "u", "http://localhost:5016/api", "URL for the API service")
	checkCmd.Flags().StringP("apikey", "k", "", "apikey to pass in REST API calls")
}

func cliCmdCheck(cmd *cobra.Command, args []string) {
	files, err := cmd.Flags().GetString("files")
	if err != nil {
		fmt.Print("error", err)
		os.Exit(1)
	}

	headerRow, err := cmd.Flags().GetInt("header-row")
	if err != nil {
		fmt.Print("error", err)
		os.Exit(1)
	}

	hasParent, err := cmd.Flags().GetBool("has-parent")
	if err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}

	loader := spreadsheet.NewLoader(hasParent, headerRow, strings.Split(files, ","))

	worksheets, err := loader.Load()
	if err != nil {
		fmt.Println("Loading spreadsheet failed")
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				fmt.Println(" ", e)
			}
		}
		os.Exit(1)
	}

	client, err := createAPIClient(checkCmd)
	if err != nil {
		// No API Client params were set
		return
	}

	var projectID string
	if projectID, err = checkCmd.Flags().GetString("project-id"); err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}

	if client != nil && projectID != "" {
		if err := loader.ValidateFilesExistInProject(worksheets, projectID, client); err != nil {
			if merr, ok := err.(*multierror.Error); ok {
				for _, e := range merr.Errors {
					fmt.Println(" ", e)
				}
			}
			os.Exit(1)
		}
	}
}
