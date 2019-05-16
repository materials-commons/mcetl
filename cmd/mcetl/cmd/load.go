package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/materials-commons/mcetl/internal/spreadsheet/model"

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
	loadCmd.Flags().StringP("files", "f", "", "Path(s) to the excel spreadsheet(s) to create experiment from")
	loadCmd.Flags().StringP("project-id", "p", "", "Project to create experiment in")
	loadCmd.Flags().StringP("project-name", "m", "", "Project name to create experiment in")
	loadCmd.Flags().StringP("experiment-name", "n", "", "Name of experiment to create")
	loadCmd.Flags().StringP("mcurl", "u", "http://localhost:5016/api", "URL for the API service")
	loadCmd.Flags().StringP("apikey", "k", "", "apikey to pass in REST API calls")
	loadCmd.Flags().StringP("project-base-dir", "d", "", "project base dir on server to look for files")
	loadCmd.Flags().IntP("header-row", "r", 0, "Row to start reading from")
	loadCmd.Flags().BoolP("has-parent", "t", false, "2nd column is the parent column")
}

func cliCmdLoad(cmd *cobra.Command, args []string) {
	worksheets, err := loadSpreadsheet(cmd)
	if err != nil {
		os.Exit(1)
	}

	// add baseDir to all file entries in the worksheets
	if err := addBaseDirToFilePaths(cmd, worksheets); err != nil {
		os.Exit(1)
	}

	client, err := createAPIClient(cmd)
	if err != nil {
		os.Exit(1)
	}

	if err := createWorkflowFromWorksheets(cmd, client, worksheets); err != nil {
		os.Exit(1)
	}
}

// loadSpreadsheet loads the excel spreadsheet file given in the file flag and
// transforms it into the internal representation of worksheets.
func loadSpreadsheet(cmd *cobra.Command) ([]*model.Worksheet, error) {
	var (
		files     string
		headerRow int
		hasParent bool
		err       error
	)

	if files, err = cmd.Flags().GetString("files"); err != nil {
		fmt.Println("error", err)
		return nil, err
	}

	if headerRow, err = cmd.Flags().GetInt("header-row"); err != nil {
		fmt.Println("error", err)
		return nil, err
	}

	if hasParent, err = cmd.Flags().GetBool("has-parent"); err != nil {
		fmt.Println("error", err)
		return nil, err
	}

	loader := spreadsheet.NewLoader(hasParent, headerRow, strings.Split(files, ","))

	worksheets, err := loader.Load()
	if err != nil {
		printLoadSpreadsheetErrors(err)
		return nil, errors.Errorf("failed loading file")
	}

	return worksheets, nil
}

func printLoadSpreadsheetErrors(err error) {
	fmt.Println("Loading spreadsheet failed:")
	if merr, ok := err.(*multierror.Error); ok {
		for _, e := range merr.Errors {
			fmt.Println(" ", e)
		}
	}
}

// addBaseDirToFilePaths goes through all the worksheets and their associated
// samples, for each sample it goes through the list of files and appends the
// baseDir to those entries. File entries in a spreadsheet are relative to the
// location of the spreadsheet. The baseDir represents this path within the
// context of the project on the server.
func addBaseDirToFilePaths(cmd *cobra.Command, worksheets []*model.Worksheet) error {
	if baseDir, err := cmd.Flags().GetString("project-base-dir"); err != nil {
		fmt.Println("error", err)
		return err
	} else {
		for _, worksheet := range worksheets {
			for _, sample := range worksheet.Samples {
				for _, file := range sample.Files {
					file.Path = filepath.Join(baseDir, file.Path)
				}
			}
		}
	}

	return nil
}

// createAPIClient creates a mcapi.Client setting the url and apikey
// from the mcurl and apikey environment variables or command line parameters.
func createAPIClient(cmd *cobra.Command) (*mcapi.Client, error) {
	var (
		mcurl  string
		apikey string
		err    error
	)

	if mcurl, err = cmd.Flags().GetString("mcurl"); err != nil || mcurl == "" {
		mcurl = config.GetString("mcurl")
	}

	if mcurl == "" {
		err = errors.New("mcurl not set")
		fmt.Println("error", err)
		return nil, err
	}

	if apikey, err = cmd.Flags().GetString("apikey"); err != nil || apikey == "" {
		apikey = config.GetString("apikey")
	}

	if apikey == "" {
		err = errors.New("apikey not set")
		fmt.Println("error", err)
		return nil, err
	}

	client := mcapi.NewClient(mcurl)
	client.APIKey = apikey

	return client, nil
}

// createWorkflowFromWorkWorksheets creates the server side workflow from the worksheets.
func createWorkflowFromWorksheets(cmd *cobra.Command, client *mcapi.Client, worksheets []*model.Worksheet) error {
	var (
		projectId      string
		experimentName string
		projectName    string
		hasParent      bool
		err            error
	)

	if projectName, err = cmd.Flags().GetString("project-name"); err != nil || projectName == "" {
		if projectId, err = cmd.Flags().GetString("project-id"); err != nil {
			fmt.Println("error", err)
			return err
		}
	} else {
		project, err := client.CreateProject(projectName, "")
		if err != nil {
			fmt.Println("error", err)
			return err
		}

		projectId = project.ID
	}

	if hasParent, err = cmd.Flags().GetBool("has-parent"); err != nil {
		fmt.Println("error", err)
		return err
	}

	if experimentName, err = cmd.Flags().GetString("experiment-name"); err != nil {
		fmt.Println("error", err)
		return err
	}

	// Create the server side representation of the workflow from the worksheets
	if err := spreadsheet.Create(projectId, experimentName, hasParent, client).Apply(worksheets); err != nil {
		fmt.Println("Unable to process spreadsheet:", err)
		return err
	}

	return nil
}
