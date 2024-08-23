package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

var shareSecretsCmd = &cobra.Command{
	Use:   "share",
	Short: "Add, update, or delete secrets based on .env files for a specific environment",
	Long: `The share command allows you to add, update, or delete key/value pairs 
from .env files into the database for the specified project and environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectName, _ := cmd.Flags().GetString("project")
		isDev, _ := cmd.Flags().GetBool("dev")
		isStaging, _ := cmd.Flags().GetBool("staging")
		isProd, _ := cmd.Flags().GetBool("prod")
		secretPair, _ := cmd.Flags().GetString("secret")

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

		if secretPair != "" {
			// Handle single key/value pair passed via --secret
			err = handleSingleSecret(db, projectName, environmentType, secretPair)
		} else {
			// Handle .env files
			err = handleEnvFiles(db, projectName, environmentType)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to process secrets: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shareSecretsCmd)

	// Flags for the add secrets command
	shareSecretsCmd.Flags().StringP("project", "p", "", "Project name")
	shareSecretsCmd.Flags().BoolP("dev", "d", false, "Add secrets for the development environment")
	shareSecretsCmd.Flags().BoolP("staging", "g", false, "Add secrets for the staging environment")
	shareSecretsCmd.Flags().BoolP("prod", "r", false, "Add secrets for the production environment")
	shareSecretsCmd.Flags().StringP("secret", "s", "", "Single key=value pair to add or update as a secret")
}

func handleSingleSecret(db *sql.DB, projectName, environmentType, secretPair string) error {
	// Split the key=value pair
	parts := strings.SplitN(secretPair, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for --secret flag. Expected format: key=value")
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Determine the location as the current directory
	location := "."

	// Check if the secret already exists
	secretExists, err := dbpkg.SecretExists(db, key, projectName, environmentType)
	if err != nil {
		return fmt.Errorf("error checking if secret exists: %v", err)
	}

	if secretExists {
		// Update existing secret
		err = dbpkg.UpdateSecret(db, key, value, location, projectName, environmentType)
		if err != nil {
			return fmt.Errorf("error updating secret: %v", err)
		}
		fmt.Printf("Updated secret: %s\n", key)
	} else {
		// Insert new secret
		err = dbpkg.CreateSecret(db, key, value, location, projectName, environmentType)
		if err != nil {
			return fmt.Errorf("error creating secret: %v", err)
		}
		fmt.Printf("Created new secret: %s\n", key)
	}

	return nil
}

func handleEnvFiles(db *sql.DB, projectName, environmentType string) error {
	// Get the current working directory
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting the current working directory: %v", err)
	}

	// Track the keys found in local .env files
	localKeys := make(map[string]bool)

	// Walk through the directory recursively
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a .env extension
		if filepath.Ext(info.Name()) == ".env" {
			fmt.Printf("Processing file: %s\n", path)

			// Determine the relative path to the .env file
			relativePath, err := filepath.Rel(root, filepath.Dir(path))
			if err != nil {
				return fmt.Errorf("error calculating relative path: %v", err)
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file %s: %v", path, err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				// Skip lines that are comments or empty
				if strings.HasPrefix(line, "#") || line == "" {
					continue
				}

				// Handle inline comments by stripping everything after the first #
				if index := strings.Index(line, "#"); index != -1 {
					line = strings.TrimSpace(line[:index])
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Track this key as found locally
				localKeys[key] = true

				// Check if the secret already exists
				secretExists, err := dbpkg.SecretExists(db, key, projectName, environmentType)
				if err != nil {
					return fmt.Errorf("error checking if secret exists: %v", err)
				}

				if secretExists {
					// Update existing secret
					err = dbpkg.UpdateSecret(db, key, value, relativePath, projectName, environmentType)
					if err != nil {
						return fmt.Errorf("error updating secret: %v", err)
					}
					fmt.Printf("Updated secret: %s\n", key)
				} else {
					// Insert new secret
					err = dbpkg.CreateSecret(db, key, value, relativePath, projectName, environmentType)
					if err != nil {
						return fmt.Errorf("error creating secret: %v", err)
					}
					fmt.Printf("Created new secret: %s\n", key)
				}
			}

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading file %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Delete secrets that are in the database but not in the local .env files
	err = deleteUnusedSecrets(db, projectName, environmentType, localKeys)
	if err != nil {
		return fmt.Errorf("error deleting unused secrets: %v", err)
	}

	return nil
}

func deleteUnusedSecrets(db *sql.DB, projectName, environmentType string, localKeys map[string]bool) error {
	// Get all keys from the database for the given project and environment
	dbKeys, err := dbpkg.GetAllSecretsKeys(db, projectName, environmentType)
	if err != nil {
		return fmt.Errorf("error fetching keys from database: %v", err)
	}

	// Delete keys that are in the database but not in the local .env files
	for _, key := range dbKeys {
		if !localKeys[key] {
			err = dbpkg.DeleteSecret(db, key, projectName, environmentType)
			if err != nil {
				return fmt.Errorf("error deleting secret: %v", err)
			}
			fmt.Printf("Deleted unused secret: %s\n", key)
		}
	}

	return nil
}
