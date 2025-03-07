package httpprobe

import (
	"fmt"
	"os"

	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/runner"
	"github.com/mrfoh/httpprobe/internal/tests"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// displayResults outputs the test results in a formatted table grouped by test suite
func displayResults(results map[string]tests.TestDefinitionExecResult) {
	var totalTests, passedTests int

	// First, organize data by definition and suite
	for defName, defResult := range results {
		fmt.Printf("\n[1m%s[0m\n", defName)

		for suiteName, suiteResult := range defResult.Suites {
			// Create a new table for each suite
			fmt.Printf("\n  [1mSuite: %s[0m\n", suiteName)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Test Case", "Status", "Time (s)"})
			table.SetBorder(false)
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
				tablewriter.Colors{tablewriter.Bold},
			)

			// Track suite statistics
			var suiteTotal, suitePassed int

			// Add the test cases for this suite
			for caseName, caseResult := range suiteResult.Cases {
				totalTests++
				suiteTotal++

				status := "FAIL"
				color := tablewriter.FgRedColor

				if caseResult.Passed {
					status = "PASS"
					color = tablewriter.FgGreenColor
					passedTests++
					suitePassed++
				}

				// Format the time with 2 decimal places
				timeStr := fmt.Sprintf("%.2f", caseResult.Timing)

				table.Rich(
					[]string{caseName, status, timeStr},
					[]tablewriter.Colors{
						{},
						{color},
						{},
					},
				)
			}

			table.Render()

			// Print suite summary
			suitePassRate := 0.0
			if suiteTotal > 0 {
				suitePassRate = (float64(suitePassed) / float64(suiteTotal)) * 100
			}

			fmt.Printf("  Suite Summary: %d/%d tests passed (%.1f%%)\n",
				suitePassed, suiteTotal, suitePassRate)
		}
	}

	// Print overall summary
	overallPassRate := 0.0
	if totalTests > 0 {
		overallPassRate = (float64(passedTests) / float64(totalTests)) * 100
	}

	fmt.Printf("\n[1mOverall Summary:[0m %d/%d tests passed (%.1f%%)\n",
		passedTests, totalTests, overallPassRate)
}

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
