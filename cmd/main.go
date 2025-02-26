package main

import (
	"log"

	"github.com/mrfoh/httpprobe/cmd/httpprobe"
)

func main() {
	// Root command
	rootCmd := httpprobe.NewRootCmd()
	versionCmd := httpprobe.NewVersionCmd()
	runCmd := httpprobe.NewRunCmd()

	rootCmd.AddCommand(versionCmd, runCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing root command: %v", err)
	}
}
