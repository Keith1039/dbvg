// Package validate provides the commands relating to schema validation
//
// This package contains the validation code for the CLI. It mirrors the functionality and depends on the `graph` package
package validate

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	ConnString string
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "The palette responsible for schema validation.",
	Long: `This palette is responsible for detecting and removing cycles from the database schema.
Suggestions can also be given regarding the detected cycles.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func addSubCommands() {
	ValidateCmd.AddCommand(schemaCmd)
}

func init() {
	addSubCommands()
	ValidateCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")

	if err := ValidateCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// validateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
