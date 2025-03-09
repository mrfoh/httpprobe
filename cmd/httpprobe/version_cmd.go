package httpprobe

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	VERSION = "1.0.0"
)

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
