package cmd

import (
	"fmt"

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
		return
	}

	processes, err := spreadsheet.Load(file)

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
}
