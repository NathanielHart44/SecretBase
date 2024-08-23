package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

// createProjectCmd represents the create project command
var createProjectCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project with associated environments",
	Long: `The create project command allows you to create a new project in the database,
along with its associated development, staging, and production environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.CheckIfStarted(started)

		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			var err error
			name, err = helpers.GetCurrentDirName()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to determine project name: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Using current directory name as project name: %s\n", name)
		}

		db, err := dbpkg.ConnectToDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to the database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		err = dbpkg.CreateProject(db, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create project: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createProjectCmd)

	// Flags for the create project command
	createProjectCmd.Flags().StringP("name", "n", "", "Name of the project")
}
