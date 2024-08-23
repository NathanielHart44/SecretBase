/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/sbx/helpers"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Sets up the environment by creating and updating .env and .env.example files.",
	Long: `The setup command scans the current directory for .env files,
extracts the keys, and creates or updates corresponding .env.example files.
Existing values in .env.example files are preserved where applicable,
and any keys not present in the .env file are removed.`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.CheckIfStarted(started)

		processEnvFiles()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// Extract keys and preserve the entire line including comments on the same line as key-value pairs
func extractKeys(content string) map[string]string {
	keys := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(trimmedLine, "=") {
			// Handle key-value pairs
			key := strings.SplitN(trimmedLine, "=", 2)[0]
			keys[key] = trimmedLine
		}
	}
	return keys
}

// Create a new line for the .env.example file
func createNewLine(key, line string) string {
	// Handle key-value pairs with comments on the same line
	keyValue := strings.SplitN(line, "=", 2)
	value := keyValue[1]
	commentIndex := strings.Index(value, "#")
	if commentIndex != -1 {
		// Preserve comments and set the value to ''
		comment := value[commentIndex:]
		return key + "='' " + comment
	} else {
		// Set the value to ''
		return key + "=''"
	}
}

// Write the .env.example file with the given content
func writeEnvExampleFile(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(content), 0644)
}

// Update the .env.example file based on the .env file
func updateEnvExampleFile(envKeys map[string]string, exampleFilePath string) error {
	var updatedExampleLines []string

	// Sort keys alphabetically
	sortedKeys := make([]string, 0, len(envKeys))
	for key := range envKeys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	// Check if .env.example file already exists
	if _, err := os.Stat(exampleFilePath); os.IsNotExist(err) {
		// If it doesn't exist, create the .env.example file with empty values
		for _, key := range sortedKeys {
			updatedExampleLines = append(updatedExampleLines, createNewLine(key, envKeys[key]))
		}
	} else {
		// If it exists, read its content
		exampleContentBytes, err := os.ReadFile(exampleFilePath)
		if err != nil {
			fmt.Println("Error reading .env.example file:", err)
			return err
		}

		// Extract existing keys and lines from the .env.example file
		exampleKeys := extractKeys(string(exampleContentBytes))

		// Add all keys from the .env file in order, and preserve existing values if they exist
		for _, key := range sortedKeys {
			if existingLine, exists := exampleKeys[key]; exists {
				// Preserve the existing value from the .env.example file if it exists
				updatedExampleLines = append(updatedExampleLines, existingLine)
			} else {
				// Add the key with an empty value and preserve any comments
				updatedExampleLines = append(updatedExampleLines, createNewLine(key, envKeys[key]))
			}
		}
	}

	// Write the updated content back to .env.example
	return writeEnvExampleFile(exampleFilePath, updatedExampleLines)
}

func processEnvFiles() {
	// Get the current working directory
	root, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting the current working directory:", err)
		return
	}

	// Print the name of the current directory
	currentDir := filepath.Base(root)
	fmt.Println("Operating in directory:", currentDir)

	// Walk through the directory recursively
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a .env extension
		if filepath.Ext(info.Name()) == ".env" {
			// Calculate the relative path
			relativePath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			// Print the relative path with a leading slash
			fmt.Println("/" + relativePath)

			// Read the contents of the .env file
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Extract keys and lines from the .env file
			envKeys := extractKeys(string(content))

			// Create the .env.example file path
			exampleFilePath := strings.TrimSuffix(path, ".env") + ".env.example"

			// Update the .env.example file
			err = updateEnvExampleFile(envKeys, exampleFilePath)
			if err != nil {
				fmt.Println("Error updating .env.example file:", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking the directory:", err)
	}
}
