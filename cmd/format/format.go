// Package format is all about formating and updating templates
//
// The format package provides the utility to create and update existing templates
package format

import (
	"github.com/spf13/cobra"
	"log"
)

var ConnString string
var path string

// FormatCmd represents the format command
var FormatCmd = &cobra.Command{
	Use:   "format",
	Short: "The palette responsible for formatting and updating templates",
	Long: `This palette is responsible for formatting and updating the JSON templates that are used
by other portions of the code.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func addSubCommands() {
	FormatCmd.AddCommand(insertTemplateCmd)
}

func init() {
	FormatCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")
	FormatCmd.PersistentFlags().StringVarP(&path, "path", "p", "", "specifies which file to use or output data to")

	if err := FormatCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	if err := FormatCmd.MarkPersistentFlagRequired("path"); err != nil {
		log.Fatal(err)
	}

	addSubCommands()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// formatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// formatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
