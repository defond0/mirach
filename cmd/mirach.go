package cmd

import (
	"fmt"
	"os"

	"github.com/cleardataeng/mirach/lib"
	"github.com/cleardataeng/mirach/lib/util"

	"github.com/spf13/cobra"
)

var (
	fullText bool
	licGroup string
	version  bool
)

// MirachCmd is the root mirach command.
var MirachCmd = &cobra.Command{
	Use:   "mirach",
	Short: "mirach collects and sends data.",
	Long: "mirach is a lightweight tool for collecting information about a " +
		"computer or attached devices. It can then send the data to a clearing house.",
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			util.ShowVersion()
			return
		}
		if err := lib.SetLogLevel(level); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := lib.Start(); err != nil {
			util.CustomOut(nil, err)
			os.Exit(1)
		}
	},
}

func init() {
	licenseCmd.Flags().BoolVar(&fullText, "full-text", false, "display full text for each license")
	licenseCmd.Flags().StringVarP(&licGroup, "group", "g", "mirach",
		`which licenses to display: "all", "mirach", or "other" for libraries used in mirach`)
	MirachCmd.Flags().BoolVar(&version, "version", false, "display current mirach version")
	MirachCmd.AddCommand(licenseCmd, pluginCmd, versionCmd)
	pluginCmd.Flags().StringVarP(&pluginName, "plugin", "p", "all", "run a given plugin")
}
