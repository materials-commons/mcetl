package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"

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
	loadCmd.Flags().StringP("project-base-dir", "d", "", "project base dir on server to look for files")
}

func cliCmdLoad(cmd *cobra.Command, args []string) {
	var (
		baseDir   string
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

	if baseDir, err = cmd.Flags().GetString("project-base-dir"); err != nil {
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

	worksheets, err := spreadsheet.Load(file)
	if err != nil {
		fmt.Println("Loading spreadsheet failed:")
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				fmt.Println(" ", e)
			}
		}
		os.Exit(1)
	}

	// add baseDir to all file entries in the worksheets
	addBaseDirToFilePaths(baseDir, worksheets)

	client := mcapi.NewClient(mcurl)
	client.APIKey = apikey

	if err := spreadsheet.Create(projectId, name, client).Apply(worksheets); err != nil {
		fmt.Println("Unable to process spreadsheet:", err)
		os.Exit(1)
	}
}

// addBaseDirToFilePaths goes through all the worksheets and their associated
// samples, for each sample it goes through the list of files and appends the
// baseDir to those entries. File entries in a spreadsheet are relative to the
// location of the spreadsheet. The baseDir represents this path within the
// context of the project on the server.
func addBaseDirToFilePaths(baseDir string, worksheets []*model.Worksheet) {
	for _, worksheet := range worksheets {
		for _, sample := range worksheet.Samples {
			for _, file := range sample.Files {
				file.Path = filepath.Join(baseDir, file.Path)
			}
		}
	}
}
