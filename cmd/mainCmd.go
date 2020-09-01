package cmd

import (
	"github.com/spf13/cobra"
)

// MainCmd is main dogrelay command
var MainCmd = &cobra.Command{
	Use: "dogrelay",
}

func init() {
	MainCmd.AddCommand(
		versionCmd,
		blackholeCmd,
		influxCmd,
		logstashToSentryCmd,
	)
}
