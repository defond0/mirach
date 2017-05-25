package cmd

import (
	"fmt"

	"github.com/cleardataeng/mirach/util"
	"github.com/spf13/cobra"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display license information about mirach.",
	Long:  "Display license information about mirach.",
	Run: func(cmd *cobra.Command, args []string) {
		if licenseGroup == "mirach" || licenseGroup == "all" {
			util.ShowMirachLicense()
		}
		if licenseGroup == "all" {
			fmt.Print("\n\n\n")
		}
		if licenseGroup == "other" || licenseGroup == "all" {
			util.ShowOtherLicenses(incText)
		}
	},
}
