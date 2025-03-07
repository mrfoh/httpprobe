package httpprobe

import "github.com/spf13/cobra"

var (
	defaultTestFileExtensions = []string{".test.yaml", ".test.json"}
	defaultSearchPath         = "./"
)

var (
	// EnvFile is the path to a file containing environment variables to be used in the tests
	EnvFile string
	// SearchPath is the path to search for test files
	SearchPath string
	// ConcurrentSuites is the number of concurrent tests suites to run
	ConcurrentSuites int
	// Verbose is a flag to enable verbose output
	Verbose bool
	// FileExtensions is a list of file extensions that have test definitions
	FileExtensions []string
	// OutputFormat is the format to output the results in; text, json, table
	OutputFormat string
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

	cmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().StringVarP(&EnvFile, "envfile", "e", ".env", "Environment file to load environment variables from")
	cmd.PersistentFlags().StringVarP(&SearchPath, "searchpath", "p", defaultSearchPath, "Path to search for test files")
	cmd.PersistentFlags().StringSliceVarP(&FileExtensions, "include", "i", defaultTestFileExtensions, "Include tests with the specified extensions")
	cmd.PersistentFlags().IntVarP(&ConcurrentSuites, "concurrency", "c", 2, "Number of concurrent tests defintions to execute")
	cmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", "text", "Output format to use; text, json, table")

	return cmd
}
