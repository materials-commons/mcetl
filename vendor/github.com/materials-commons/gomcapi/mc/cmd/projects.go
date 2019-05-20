// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"time"

	"github.com/materials-commons/gomcapi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:     "projects",
	Short:   "Lists all projects a user has access to.",
	Long:    ``,
	Aliases: []string{"proj", "p"},
	Run: func(cmd *cobra.Command, args []string) {
		projects, err := mcapi.GetAllProjects()
		if err != nil {
			fmt.Println("Unable to retrieve projects:", err)
			os.Exit(1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Description", "Owner", "Id", "MTime"})
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		for _, proj := range projects {
			t := time.Time(proj.MTime)
			dt := t.Format(time.RFC1123)
			table.Append([]string{proj.Name, proj.Description, proj.Owner, proj.ID, dt})
		}

		table.Render()

	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
