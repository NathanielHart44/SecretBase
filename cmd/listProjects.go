package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

// listProjectsCmd represents the list projects command
var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	Long:  `List all projects in the system, displaying their names and active status.`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.CheckIfStarted(started)

		db, err := dbpkg.ConnectToDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to the database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		rows, err := db.Query("SELECT name, active FROM projects")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute query: %v\n", err)
			return
		}
		defer rows.Close()

		// Create a table to display the results
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Project Name", "Active"})

		for rows.Next() {
			var name string
			var active bool
			if err := rows.Scan(&name, &active); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
				return
			}

			activeStr := "No"
			if active {
				activeStr = "Yes"
			}

			table.Append([]string{name, activeStr})
		}

		if err := rows.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error during rows iteration: %v\n", err)
			return
		}

		// Render the table to stdout
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(listProjectsCmd)
}
