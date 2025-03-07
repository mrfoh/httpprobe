package tests

import (
	"fmt"

	"github.com/fatih/color"
)

type TextResultWriter struct{}

func NewTextResultWriter() *TextResultWriter {
	return &TextResultWriter{}
}

func (w *TextResultWriter) Write(results map[string]TestDefinitionExecResult) {
	totalPassedSuites := 0
	testSuiteCount := 0
	testCaseCount := 0
	passedTestCaseCount := 0
	totalTiming := 0.0

	for defName, defResult := range results {
		color.Cyan("%s: %s\n", defName, defResult.Path)

		for suiteName, suiteResult := range defResult.Suites {
			testSuiteCount++
			testCaseCount += len(suiteResult.Cases)

			if suiteResult.Passed() {
				totalPassedSuites++
			}

			fmt.Printf("  Suite: %s\n", suiteName)

			for caseName, caseResult := range suiteResult.Cases {
				status := color.RedString("FAIL")

				if caseResult.Passed {
					status = color.GreenString("PASS")
					passedTestCaseCount++
				}

				totalTiming += caseResult.Timing

				fmt.Printf("    %s (%.2f ms): %s\n", caseName, caseResult.Timing, status)
				
				// If the test failed and we have failure reasons, display them
				if !caseResult.Passed && len(caseResult.FailureReasons) > 0 {
					fmt.Println("      Failures:")
					for _, reason := range caseResult.FailureReasons {
						fmt.Printf("        - %s\n", reason)
					}
				}
			}
		}
	}

	color.White("\nTest Suites: %s, %d total\n", color.GreenString("%d passed", totalPassedSuites), testSuiteCount)
	color.White("Test Cases: %s, %d total\n", color.GreenString("%d passed", passedTestCaseCount), testCaseCount)
	color.White("Total time: %.2f ms\n", totalTiming)
}
