package cmd

import (
	"fmt"

	"cleardata.com/dash/mirach/plugins/compinfo"
	"cleardata.com/dash/mirach/plugins/pkginfo"
	"github.com/spf13/cobra"
)

var compinfoCmd = &cobra.Command{
	Use:   "compinfo",
	Short: "Run one of mirach's built in compinfo plugin.",
	Long: "mirach plugins are primarily used from within mirach, but this allows " +
		"you to run this one directly directly. It will return a json string of " +
		"the type passed in with the -i switch or system information by default.",
	Run: func(cmd *cobra.Command, args []string) {
		switch InfoGroup {
		case "load":
			fmt.Println(compinfo.GetLoadString())
		case "docker":
			fmt.Println(compinfo.GetDockerString())
		case "sys", "system":
			fmt.Println(compinfo.GetSysString())
		default:
			fmt.Printf("choose infogroup from %s", "docker, load, system")
		}
	},
}

var pkginfoCmd = &cobra.Command{
	Use:   "pkginfo",
	Short: "Run one of mirach's built in pkginfo plugin.",
	Long: "mirach plugins are primarily used from within mirach, but this allows " +
		"you to run this one directly directly. It will return a json string of " +
		"the type passed in with the -i switch or system information by default.",
	Run: func(cmd *cobra.Command, args []string) {
		if InfoGroup != "all" {
			fmt.Println(pkginfo.GetInfoGroup(InfoGroup))
		} else {
			pkginfo.GetInfo()
			fmt.Println(pkginfo.String())
		}
	},
}
