package cmd

import (
	"fmt"
	"os"

	"gitlab.eng.cleardata.com/dash/mirach/mirachlib"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"
	"gitlab.eng.cleardata.com/dash/mirach/util"

	"github.com/spf13/cobra"
)

// flag variables
var (
	compInfoGroup string
	incText       bool
	level         string
	licenseGroup  string
	pkgInfoGroup  string
	version       bool
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
			util.CustomOut(nil, err)
			os.Exit(1)
		}
	},
}

func init() {
	if envinfo.Env == nil {
		envinfo.Env = new(envinfo.EnvInfoGroup)
		envinfo.Env.GetInfo()
	}
	MirachCmd.PersistentFlags().StringVarP(&level, "loglevel", "l", "error",
		"log level: error, info, trace")
	MirachCmd.Flags().BoolVar(&version, "version", false, "display current mirach version")

	MirachCmd.AddCommand(compinfoCmd)
	compinfoCmd.Flags().StringVarP(&compInfoGroup, "infogroup", "i", "system",
		"compinfo group to check: docker, load, system")

	MirachCmd.AddCommand(pkginfoCmd)
	pkginfoCmd.Flags().StringVarP(&pkgInfoGroup, "infogroup", "i", "all",
		"pkginfo group to check: available, available_security, installed")
	MirachCmd.AddCommand(envinfoCmd)
	MirachCmd.AddCommand(ebsinfoCmd)
	MirachCmd.AddCommand(licenseCmd)
	licenseCmd.Flags().BoolVarP(&incText, "include-text", "t", false,
		"display full text for each license")
	licenseCmd.Flags().StringVarP(&licenseGroup, "group", "g", "mirach",
		`which licenses to display: "all", "mirach", or "other" for libraries used in mirach`)
	MirachCmd.AddCommand(versionCmd)
}
