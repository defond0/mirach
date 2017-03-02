package cmd

import (
	"fmt"
	"os"

	"gitlab.eng.cleardata.com/dash/mirach/mirachlib"
	"gitlab.eng.cleardata.com/dash/mirach/util"

	"github.com/spf13/cobra"
)

// flag variables
var (
	infoGroup string
	level     string
	version   bool
)

// MirachCmd is the root mirach command.
var MirachCmd = &cobra.Command{
	Use:   "mirach",
	Short: "mirach collects data and send it over mqtt.",
	Long: "mirach is a lightweight tool for collecting information about a " +
		"computer then sending that information to an mqtt broker for consumption. " +
		"The command alone runs the main mirach program.",
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			util.ShowVersion()
			return
		}
		if err := mirachlib.SetLogLevel(level); err != nil {
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
	MirachCmd.PersistentFlags().StringVarP(&level, "loglevel", "l", "error",
		"log level: error, info, trace")
	MirachCmd.Flags().BoolVar(&version, "version", false, "display current mirach version")

	MirachCmd.AddCommand(compinfoCmd)
	compinfoCmd.Flags().StringVarP(&infoGroup, "infogroup", "i", "system",
		"compinfo group to check: docker, load, system")

	MirachCmd.AddCommand(pkginfoCmd)
	pkginfoCmd.Flags().StringVarP(&infoGroup, "infogroup", "i", "all",
		"pkginfo group to check: available, available_security, installed")

	MirachCmd.AddCommand(versionCmd)
}
