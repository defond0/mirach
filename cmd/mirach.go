package cmd

import (
	"fmt"
	"os"

	"cleardata.com/dash/mirach/mirachlib"

	"github.com/spf13/cobra"
)

// Level is a flag for setting the log level.
var Level string

// InfoGroup is a flag used to specify which compinfo group to retrieve.
var InfoGroup string

// MirachCmd is the root mirach command.
var MirachCmd = &cobra.Command{
	Use:   "mirach",
	Short: "mirach collects data and send it over mqtt.",
	Long: "mirach is a lightweight tool for collecting information about a " +
		"computer then sending that information to an mqtt broker for consumption. " +
		" The command alone runs the main mirach program.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := mirachlib.SetLogLevel(Level); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := mirachlib.Start(); err != nil {
			mirachlib.CustomOut(nil, err)
			os.Exit(1)
		}
	},
}

func init() {
	MirachCmd.AddCommand(compinfoCmd)
	MirachCmd.AddCommand(pkginfoCmd)
	MirachCmd.PersistentFlags().StringVarP(&Level, "loglevel", "l", "error",
		"log level: error (default), info, trace")
	compinfoCmd.Flags().StringVarP(&InfoGroup, "infogroup", "i", "system",
		"compinfo group to check: docker, load, system")
	pkginfoCmd.Flags().StringVarP(&InfoGroup, "infogroup", "i", "avail_sec",
		"pkginfo group to check: available, available_security, installed")
}
