package httpprobe

import (
	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/runner"
	"github.com/mrfoh/httpprobe/internal/tests"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute tests in the specified path or file",
		Long:  "Execute tests in the specified path or file",
		Run: func(cmd *cobra.Command, args []string) {
			testFilesSearchPath, _ := cmd.Flags().GetString("searchpath")
			testFileExtensions, _ := cmd.Flags().GetStringSlice("include")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			verbose, _ := cmd.Flags().GetBool("verbose")
			output, _ := cmd.Flags().GetString("output")
			envFile, _ := cmd.Flags().GetString("envfile")

			// Load environment variables from file
			if err := tests.LoadEnvFile(envFile); err != nil {
				cmd.PrintErrf("Error loading environment variables: %v\n", err)
				return
			}

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

			if envFile != "" {
				logger.Debug("Loaded environment variables from file", zap.String("file", envFile))
			}

			httpClientOptions := easyreq.NewOptions().
				WithLogger(logger)

			httpClient := easyreq.New(httpClientOptions)

			parser := tests.NewTestDefinitionParser()

			writer := tests.NewResultWriter(output)

			runnerOptions := runner.NewOptions().
				SetLogger(logger).
				SetParser(parser).
				SetConcurrency(concurrency).
				SetHttpClient(httpClient).
				SetResultWriter(writer)

			testrunner := runner.NewRunner(runnerOptions)

			definitions, err := testrunner.GetTestDefinitions(&runner.GetTestDefinitionsParams{
				SearchPath:     testFilesSearchPath,
				FileExtensions: testFileExtensions,
			})
			if err != nil {
				cmd.PrintErrln(err)
				return
			}

			results, err := testrunner.Execute(definitions)
			if err != nil {
				cmd.PrintErrln(err)
				return
			}

			testrunner.Write(results)
		},
	}

	return cmd
}
