package tests

import (
	"fmt"
)

type TableResultWriter struct{}

func NewTableResultWriter() *TableResultWriter {
	return &TableResultWriter{}
}

func (w *TableResultWriter) Write(results map[string]TestDefinitionExecResult) {
	// Implement table-based output with failure details
	fmt.Println("+-----------------+-----------------+--------+------+-------------------+")
	fmt.Println("| Test Definition | Test Suite      | Test Case           | Result | Failures            |")
	fmt.Println("+-----------------+-----------------+--------+------+-------------------+")
	
	for defName, defResult := range results {
		isFirstDef := true
		
		for suiteName, suiteResult := range defResult.Suites {
			isFirstSuite := true
			
			for caseName, caseResult := range suiteResult.Cases {
				result := "PASS"
				if !caseResult.Passed {
					result = "FAIL"
				}
				
				// For the first row of a test definition, print the definition name
				defCell := ""
				if isFirstDef {
					defCell = defName
					isFirstDef = false
				}
				
				// For the first row of a suite, print the suite name
				suiteCell := ""
				if isFirstSuite {
					suiteCell = suiteName
					isFirstSuite = false
				}
				
				// Format failures for the table
				failuresCell := ""
				if !caseResult.Passed && len(caseResult.FailureReasons) > 0 {
					// Show first failure, truncate if needed
					failuresCell = caseResult.FailureReasons[0]
					if len(failuresCell) > 20 {
						failuresCell = failuresCell[:17] + "..."
					}
					
					// Indicate if there are more failures
					if len(caseResult.FailureReasons) > 1 {
						failuresCell += fmt.Sprintf(" (+%d more)", len(caseResult.FailureReasons)-1)
					}
				}
				
				fmt.Printf("| %-15s | %-15s | %-20s | %-6s | %-20s |\n", 
					defCell, suiteCell, caseName, result, failuresCell)
			}
		}
		
		fmt.Println("+-----------------+-----------------+--------+------+-------------------+")
	}
}
