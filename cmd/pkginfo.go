package cmd

import (
	"fmt"

	"gitlab.eng.cleardata.com/dash/mirach/plugin/pkginfo"

	"github.com/spf13/cobra"
)

var pkginfoCmd = &cobra.Command{
	Use:   "pkginfo",
	Short: "Run mirach's built in pkginfo plugin.",
	Long: "mirach plugins are primarily used from within mirach, but this allows " +
		"you to run this one directly. It will return a json string of the " +
		"type passed in with the -i switch or system information by default.",
	Run: func(cmd *cobra.Command, args []string) {
		if pkgInfoGroup != "all" {
			fmt.Println(pkginfo.GetInfoGroup(pkgInfoGroup))
		} else {
			pkginfo.GetInfo()
			fmt.Println(pkginfo.String())
		}
	},
}
