package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.eng.cleardata.com/dash/mirach/util"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the current mirach version.",
	Long: "This will be a semantic version with the version of source " +
		"from which this application was built.",
	Run: func(cmd *cobra.Command, args []string) {
		util.ShowVersion()
	},
}
