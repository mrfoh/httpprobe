package httpprobe

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set during build using -ldflags
var version = "dev"

// VERSION is exported for backward compatibility
var VERSION = version

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of httpprobe",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("httpprobe version %s\n", VERSION)
		},
	}

	return cmd
}
