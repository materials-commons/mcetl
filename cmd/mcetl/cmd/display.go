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
var displayCmd = &cobra.Command{
	Use:   "display",
	Short: "Displays what the spreadsheet would look like. No ETL is performed.",
	Long: `The display command validates the given spreadsheets, reports any errors and textually displays workflow. It will not perform
any ETL operations on the spreadsheets.`,
	Run: cliCmdDisplay,
}

func init() {
	rootCmd.AddCommand(displayCmd)
	displayCmd.Flags().StringP("files", "f", "", "Path to the excel spreadsheet")
	displayCmd.Flags().IntP("header-row", "r", 0, "Row to start reading from")
	displayCmd.Flags().BoolP("has-parent", "t", false, "2nd column is the parent column")
}

func cliCmdDisplay(cmd *cobra.Command, args []string) {
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

	if err := spreadsheet.Display.Apply(worksheets); err != nil {
		fmt.Println("Unable to process spreadsheet:", err)
		os.Exit(1)
	}
}
