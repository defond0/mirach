package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.eng.cleardata.com/dash/mirach/util"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display license information about mirach.",
	Long:  "Display license information about mirach.",
	Run: func(cmd *cobra.Command, args []string) {
		util.ShowLicenseInfo()
	},
}
