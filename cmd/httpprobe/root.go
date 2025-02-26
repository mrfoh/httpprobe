package httpprobe

import "github.com/spf13/cobra"

var (
	VERSION = "0.0.1"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "httpprobe",
		Short: "CLI tool to test HTTP API endpoints",
		Long:  "CLI tool to test HTTP API endpoints using a declarative YAML configuration file.",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
		},
	}

	return cmd
}
