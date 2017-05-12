package cmd

import (
	"github.com/mono83/slf/util/slfcobra"
	"github.com/spf13/cobra"
)

// MainCmd is main dogrelay command
var MainCmd = slfcobra.Wrap(&cobra.Command{
	Use: "dogrelay",
})

func init() {
	MainCmd.AddCommand(
		versionCmd,
		influxCmd,
	)
}
