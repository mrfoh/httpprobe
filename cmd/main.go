package main

import (
	"log"

	"github.com/mrfoh/httpprobe/cmd/httpprobe"
)

func main() {
	// Root command
	rootCmd := httpprobe.NewRootCmd()
	runCmd := httpprobe.NewRunCmd()

	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing root command: %v", err)
	}
}
