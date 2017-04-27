package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"
)

var envinfoCmd = &cobra.Command{
	Use:   "envinfo",
	Short: "Run one of built in envinfo plugin.",
	Long: "mirach plugins are primarily used from within mirach, but this allows " +
		"you to run this one directly directly.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(envinfo.String())
	},
}
