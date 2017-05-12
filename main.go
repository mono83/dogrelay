package main

import (
	"fmt"
	"github.com/mono83/dogrelay/cmd"
	"os"
)

func main() {
	if err := cmd.MainCmd.Execute(); err != nil {
		fmt.Println("Execution error occured:", err)
		os.Exit(1)
	}
}
