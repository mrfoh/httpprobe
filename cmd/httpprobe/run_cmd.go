package httpprobe

import (
	"github.com/alitto/pond/v2"
	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/runner"
	"github.com/mrfoh/httpprobe/internal/tests"
	"github.com/spf13/cobra"
)

var (
	defaultTestFileExtensions = []string{".test.yaml", ".test.json"}
	defaultSearchPath         = "./"
)

var (
	EnvFile        string
	SearchPath     string
	FileExtensions []string
	PoolSize       int
	Verbose        bool
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute tests in the specified path or file",
		Long:  "Execute tests in the specified path or file",
		Run: func(cmd *cobra.Command, args []string) {
			testFilesSearchPath, _ := cmd.Flags().GetString("searchpath")
			testFileExtensions, _ := cmd.Flags().GetStringSlice("include")
			verbose, _ := cmd.Flags().GetBool("verbose")

			var loggingLevel string

			if verbose {
				loggingLevel = "debug"
			} else {
				loggingLevel = "info"
			}

			logger, err := logging.NewLogger(&logging.LoggerOptions{
				LogLevel: loggingLevel,
			})
			if err != nil {
				cmd.PrintErrln(err)
				return
			}

			parser := tests.NewTestDefinitionParser()
			pool := pond.NewResultPool[interface{}](PoolSize)

			testrunner := runner.NewRunner(parser, logger, pool)

			definitions, err := testrunner.GetTestDefinitions(&runner.GetTestDefinitionsParams{
				SearchPath:     testFilesSearchPath,
				FileExtensions: testFileExtensions,
			})
			if err != nil {
				cmd.PrintErrln(err)
				return
			}

			_, err = testrunner.Execute(definitions)
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&EnvFile, "envfile", "e", ".env", "Environment file to load environment variables from")
	cmd.PersistentFlags().StringVarP(&SearchPath, "searchpath", "p", defaultSearchPath, "Path to search for test files")
	cmd.PersistentFlags().IntVarP(&PoolSize, "pool", "s", 10, "Number of concurrent tests defintions to execute")
	cmd.PersistentFlags().StringSliceVarP(&FileExtensions, "include", "i", defaultTestFileExtensions, "Include tests with the specified extensions")
	cmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
