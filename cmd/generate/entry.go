/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package generate

import (
	"github.com/Keith1039/Capstone_Test/parameters"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	table         string
	template      string
	amount        int
	run           bool
	defaultConfig bool
)

// entryCmd represents the entry command
var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := InitDB()
		if err != nil {
			log.Fatal(err)
		}
		if amount <= 0 {
			log.Fatal("amount must be greater than zero")
		}
		if _, err := os.Stat(template); !os.IsNotExist(err) {
			_, err := parameters.NewQueryWriterFor(db, table)
			if err != nil {
				log.Fatal(err)
			}
			//writer.GenerateEntries(amount)

		} else {
			log.Fatal("template path is invalid")
		}
	},
}

func init() {

	entryCmd.Flags().StringVarP(&template, "template", "", "", "path to the template being used")
	err := entryCmd.MarkFlagRequired("template")
	if err != nil {
		log.Fatal(err)
	}

	entryCmd.Flags().IntVarP(&amount, "amount", "", 1, "amount of entries this will generate")
	entryCmd.Flags().BoolVarP(&run, "run", "", false, "generate an entry")
	entryCmd.Flags().BoolVarP(&defaultConfig, "default", "", false, "run using the default template")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// entryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// entryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
