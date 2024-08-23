package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

func CheckIfStarted(started bool) {
	if !started {
		fmt.Println("You must start sbx with the 'start' command before using any other commands.")
		os.Exit(1)
	}
}

// GetCurrentDirName returns the name of the current directory
func GetCurrentDirName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}
	return filepath.Base(dir), nil
}
