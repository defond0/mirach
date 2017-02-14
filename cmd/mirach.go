package cmd

import (
	"cleardata.com/dash/mirach/mirachlib"

	"github.com/spf13/cobra"
)

// MirachCmd is the root mirach command.
var MirachCmd = &cobra.Command{
	Use:   "mirach",
	Short: "mirach collects data and send it over mqtt.",
	Long: `mirach is a lightweight tool for collecting information
                about a computer then sending that information to an
                mqtt broker for consumption.`,
	Run: func(cmd *cobra.Command, args []string) {
		mirachlib.Start()
	},
}
