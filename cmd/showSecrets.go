package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

// showSecretsCmd represents the show secrets command
var showSecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Show secrets for a specific environment",
	Long:  `The show secrets command allows you to display the key/value pairs associated with a specific environment (dev, staging, prod) for a given project.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectName, _ := cmd.Flags().GetString("project")
		isDev, _ := cmd.Flags().GetBool("dev")
		isStaging, _ := cmd.Flags().GetBool("staging")
		isProd, _ := cmd.Flags().GetBool("prod")

		if projectName == "" {
			var err error
			projectName, err = helpers.GetCurrentDirName()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to determine project name: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Using current directory name as project name: %s\n", projectName)
		}

		var environmentType string
		switch {
		case isDev:
			environmentType = "development"
		case isStaging:
			environmentType = "staging"
		case isProd:
			environmentType = "production"
		default:
			fmt.Println("You must specify one of the following flags: --dev, --staging, or --prod")
			os.Exit(1)
		}

		db, err := dbpkg.ConnectToDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to the database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// first ProjectExists check to make sure that we can proceed
		projectExists, err := dbpkg.ProjectExists(db, projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking if project exists: %v\n", err)
			os.Exit(1)
		}
		if !projectExists {
			fmt.Fprintf(os.Stderr, "Project '%s' does not exist.\n", projectName)
			os.Exit(1)
		}

		rows, err := db.Query(`
			SELECT s.key, s.value
			FROM secrets s
			INNER JOIN environment_secrets es ON s.id = es.secret_id
			INNER JOIN environments e ON es.environment_id = e.id
			INNER JOIN projects p ON e.project_id = p.id
			WHERE p.name = ? AND e.environment_type = ?`, projectName, environmentType)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute query: %v\n", err)
			return
		}
		defer rows.Close()

		// Create a table to display the results
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Key", "Value"})

		for rows.Next() {
			var key string
			var value string
			if err := rows.Scan(&key, &value); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
				return
			}

			table.Append([]string{key, value})
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
	rootCmd.AddCommand(showSecretsCmd)

	// Flags for the show secrets command
	showSecretsCmd.Flags().StringP("project", "p", "", "Project name")
	showSecretsCmd.Flags().BoolP("dev", "d", false, "Show secrets for the development environment")
	showSecretsCmd.Flags().BoolP("staging", "s", false, "Show secrets for the staging environment")
	showSecretsCmd.Flags().BoolP("prod", "r", false, "Show secrets for the production environment")
}
