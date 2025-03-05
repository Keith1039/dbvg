/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package update

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	ConnString string
)

// UpdateCmd represents the update command
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
	},
}

func addSubCommands() {
	UpdateCmd.AddCommand(templateCmd)
}

func init() {
	UpdateCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")

	if err := UpdateCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	addSubCommands()

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
