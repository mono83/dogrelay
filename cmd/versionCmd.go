package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "1.4.0"

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dogrelay version is", version)
	},
}
