package cmd

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"

	"github.com/materials-commons/config"
	mcapi "github.com/materials-commons/gomcapi"
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
	loadCmd.Flags().StringP("mcurl", "u", "http://localhost:5016/api", "URL for the API service")
	loadCmd.Flags().StringP("apikey", "k", "", "apikey to pass in REST API calls")
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
		os.Exit(1)
	}

	if projectId, err = cmd.Flags().GetString("project-id"); err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}

	if name, err = cmd.Flags().GetString("experiment-name"); err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}

	mcurl := config.GetString("mcurl")
	if mcurl == "" {
		if mcurl, err = cmd.Flags().GetString("mcurl"); err != nil {
			fmt.Println("error", err)
			os.Exit(1)
		}
	}
	fmt.Println("Using mcurl:", mcurl)

	apikey := config.GetString("apikey")
	if apikey == "" {
		if apikey, err = cmd.Flags().GetString("apikey"); err != nil {
			fmt.Println("error", err)
			os.Exit(1)
		}
	}
	fmt.Println("Using apikey:", apikey)

	processes, err := spreadsheet.Load(file)
	if err != nil {
		fmt.Println("Loading spreadsheet failed:")
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				fmt.Println(" ", e)
			}
		}
		os.Exit(1)
	}

	client := mcapi.NewClient(mcurl)
	client.APIKey = apikey

	if err := spreadsheet.Create(projectId, name, client).Apply(processes); err != nil {
		fmt.Println("Unable to process spreadsheet:", err)
		os.Exit(1)
	}
}
