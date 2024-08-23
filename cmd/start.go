package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// a global variable to track if the CLI has started
var started bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the interactive CLI for SecretBase",
	Long: `Start the interactive CLI for SecretBase, allowing you to enter commands continuously
until you decide to exit.`,
	Run: func(cmd *cobra.Command, args []string) {
		started = true
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("sbx> ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			// Allow user to exit the loop
			if input == "exit" || input == "quit" {
				fmt.Println("Exiting SecretBase CLI...")
				break
			}

			// Execute the command entered by the user
			if input != "" {
				args := strings.Split(input, " ")
				rootCmd.SetArgs(args)
				if err := rootCmd.Execute(); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
