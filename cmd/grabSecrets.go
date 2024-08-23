package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

// grabSecretsCmd represents the grab secrets command
var grabSecretsCmd = &cobra.Command{
	Use:   "grab",
	Short: "Retrieve secrets from the database and populate .env files",
	Long: `The grab command retrieves secrets for a specified project and environment 
from the database and updates or creates .env files in the appropriate locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.CheckIfStarted(started)

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

		dbConn, err := dbpkg.ConnectToDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to the database: %v\n", err)
			os.Exit(1)
		}
		defer dbConn.Close()

		// first ProjectExists check to make sure that we can proceed
		projectExists, err := dbpkg.ProjectExists(dbConn, projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking if project exists: %v\n", err)
			os.Exit(1)
		}
		if !projectExists {
			fmt.Fprintf(os.Stderr, "Project '%s' does not exist.\n", projectName)
			os.Exit(1)
		}

		err = processSecrets(dbConn, projectName, environmentType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process secrets: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(grabSecretsCmd)

	// Flags for the grab secrets command
	grabSecretsCmd.Flags().StringP("project", "p", "", "Project name")
	grabSecretsCmd.Flags().BoolP("dev", "d", false, "Grab secrets for the development environment")
	grabSecretsCmd.Flags().BoolP("staging", "s", false, "Grab secrets for the staging environment")
	grabSecretsCmd.Flags().BoolP("prod", "r", false, "Grab secrets for the production environment")
}

func processSecrets(db *sql.DB, projectName, environmentType string) error {
	// Fetch all secrets for the given project and environment
	secrets, err := dbpkg.GetSecrets(db, projectName, environmentType)
	if err != nil {
		return fmt.Errorf("error fetching secrets: %v", err)
	}

	// Group secrets by location
	secretsByLocation := make(map[string]map[string]string)
	for _, secret := range secrets {
		if _, exists := secretsByLocation[secret.Location]; !exists {
			secretsByLocation[secret.Location] = make(map[string]string)
		}
		secretsByLocation[secret.Location][secret.Key] = secret.Value
	}

	// Process each location and create/update .env files
	for location, kvPairs := range secretsByLocation {
		// Resolve the full path
		var fullPath string
		if location == "." {
			fullPath = "./.env"
		} else {
			fullPath = filepath.Join(location, ".env")
		}

		// Create the directory if it doesn't exist
		err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %v", filepath.Dir(fullPath), err)
		}

		// Open or create the .env file
		file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error opening file %s: %v", fullPath, err)
		}
		defer file.Close()

		// Write the key/value pairs to the file
		for key, value := range kvPairs {
			_, err := file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
			if err != nil {
				return fmt.Errorf("error writing to file %s: %v", fullPath, err)
			}
		}

		fmt.Printf("Updated file: %s\n", fullPath)
	}

	return nil
}
