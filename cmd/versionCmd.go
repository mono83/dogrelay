package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "2.0.3"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dogrelay version is", version)
	},
}
