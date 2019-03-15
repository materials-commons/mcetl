package cmd

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/materials-commons/mcetl/internal/spreadsheet"

	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Loads the given spreadsheet(s) and performs ETL.",
	Long:  `The load command will read and process the given spreadsheets. The spreadsheets are processed in the order given.`,
	Run:   cliCmdLoad,
}

func init() {
	rootCmd.AddCommand(loadCmd)
	loadCmd.Flags().StringP("file", "f", "", "Path to the excel spreadsheet to create experiment from")
	loadCmd.Flags().StringP("project-id", "p", "", "Project to create experiment in")
	loadCmd.Flags().StringP("experiment-name", "n", "", "Name of experiment to create")
}

func cliCmdLoad(cmd *cobra.Command, args []string) {
	var (
		file      string
		projectId string
		name      string
		err       error
	)
	if file, err = cmd.Flags().GetString("file"); err != nil {
		fmt.Println("error", err)
		return
	}

	if projectId, err = cmd.Flags().GetString("project-id"); err != nil {
		fmt.Println("error", err)
	}

	if name, err = cmd.Flags().GetString("experiment-name"); err != nil {
		fmt.Println("error", err)
	}

	processes, err := spreadsheet.Load(file)
	if err != nil {
		fmt.Println("Loading spreadsheet failed")
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				fmt.Println(" ", e)
			}
		}

		return
	}

	if err := spreadsheet.Create(projectId, name).Apply(processes); err != nil {
		fmt.Println("Unable to process spreadsheet:", err)
	}
}
