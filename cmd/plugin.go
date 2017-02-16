package cmd

import (
	"fmt"

	"cleardata.com/dash/mirach/plugins/compinfo"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Run one of mirach's build in plugins.",
	Long: "mirach plugins are primarily used from within mirach, but this allows " +
		"you to run them directly.",
	Run: func(cmd *cobra.Command, args []string) {
		switch InfoGroup {
		case "load":
			fmt.Println(compinfo.GetLoadString())
		case "docker":
			fmt.Println(compinfo.GetDockerString())
		case "sys", "system":
			fmt.Println(compinfo.GetSysString())
		default:
			fmt.Printf("choose infogroup from %s", "docker, load, system (default)")
		}
	},
}
