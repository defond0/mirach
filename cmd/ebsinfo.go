package cmd

import (
	"fmt"

	"gitlab.eng.cleardata.com/dash/mirach/plugin/ebsinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"

	"github.com/spf13/cobra"
)

var ebsinfoCmd = &cobra.Command{
	Use:   "ebsinfo",
	Short: "Run mirach's built in ebsinfo plugin.",
	Long: "mirach plugins are primarily used from within mirach, but this allows " +
		"you to run this one directly directly. " +
		"(will not have anything to report outside of aws)",
	Run: func(cmd *cobra.Command, args []string) {
		if envinfo.Env.CloudProvider != "aws" {
			fmt.Println("Must be on an aws instance for this to be useful")
		}
		fmt.Println(ebsinfo.String())
	},
}
