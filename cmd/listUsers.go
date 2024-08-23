package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

// listUsersCmd represents the list users command
var listUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users",
	Long:  `List all users in the system, displaying their email addresses and admin status.`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.CheckIfStarted(started)

		db, err := dbpkg.ConnectToDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to the database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		rows, err := db.Query("SELECT email, admin FROM users")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute query: %v\n", err)
			return
		}
		defer rows.Close()

		// Create a table to display the results
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Email", "Admin"})

		for rows.Next() {
			var email string
			var admin bool
			if err := rows.Scan(&email, &admin); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
				return
			}

			adminStr := "No"
			if admin {
				adminStr = "Yes"
			}

			table.Append([]string{email, adminStr})
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
	rootCmd.AddCommand(listUsersCmd)
}
