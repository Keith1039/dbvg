// Package format is all about formating and updating templates
//
// The format package provides the utility to create and update existing templates
package format

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var ConnString string

// FormatCmd represents the format command
var FormatCmd = &cobra.Command{
	Use:   "format",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("format called")
	},
}

func init() {
	FormatCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")

	if err := FormatCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// formatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// formatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
