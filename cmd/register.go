package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	dbpkg "github.com/spf13/sbx/db"
	"github.com/spf13/sbx/helpers"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user in the database",
	Long: `The register command allows you to create a new user in the database.
You need to provide an email, password, and specify if the user is an admin.`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.CheckIfStarted(started)

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		admin, _ := cmd.Flags().GetBool("admin")

		if email == "" || password == "" {
			fmt.Println("Email and password are required")
			os.Exit(1)
		}

		db, err := dbpkg.ConnectToDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to the database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		err = dbpkg.CreateUser(db, email, password, admin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to register user: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)

	// Flags for the register command
	registerCmd.Flags().StringP("email", "e", "", "Email address of the user")
	registerCmd.Flags().StringP("password", "p", "", "Password for the user")
	registerCmd.Flags().BoolP("admin", "a", false, "Set user as admin")
}
