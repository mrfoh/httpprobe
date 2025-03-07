package tests

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONResultWriter struct {
	OutputPath string
}

func NewJSONResultWriter() *JSONResultWriter {
	return &JSONResultWriter{
		OutputPath: "test-results.json",
	}
}

// JSONResult is a serializable representation of test results
type JSONResult struct {
	TestDefinitions []JSONTestDefinition `json:"testDefinitions"`
	Summary         JSONSummary          `json:"summary"`
}

type JSONTestDefinition struct {
	Name   string           `json:"name"`
	Path   string           `json:"path"`
	Suites []JSONTestSuite  `json:"suites"`
}

type JSONTestSuite struct {
	Name  string          `json:"name"`
	Cases []JSONTestCase  `json:"cases"`
}

type JSONTestCase struct {
	Name          string   `json:"name"`
	Passed        bool     `json:"passed"`
	Timing        float64  `json:"timingMs"`
	FailureReasons []string `json:"failureReasons,omitempty"`
}

type JSONSummary struct {
	TotalTestDefinitions int `json:"totalTestDefinitions"`
	TotalSuites          int `json:"totalSuites"`
	PassedSuites         int `json:"passedSuites"`
	TotalCases           int `json:"totalCases"`
	PassedCases          int `json:"passedCases"`
	TotalTimeMs          float64 `json:"totalTimeMs"`
}

func (w *JSONResultWriter) Write(results map[string]TestDefinitionExecResult) {
	jsonResult := JSONResult{
		TestDefinitions: make([]JSONTestDefinition, 0, len(results)),
		Summary: JSONSummary{},
	}
	
	totalPassedSuites := 0
	testSuiteCount := 0
	testCaseCount := 0
	passedTestCaseCount := 0
	totalTiming := 0.0
	
	// Convert results to serializable format
	for defName, defResult := range results {
		jsonDef := JSONTestDefinition{
			Name: defName,
			Path: defResult.Path,
			Suites: make([]JSONTestSuite, 0, len(defResult.Suites)),
		}
		
		for suiteName, suiteResult := range defResult.Suites {
			testSuiteCount++
			jsonSuite := JSONTestSuite{
				Name: suiteName,
				Cases: make([]JSONTestCase, 0, len(suiteResult.Cases)),
			}
			
			allCasesPassed := true
			
			for caseName, caseResult := range suiteResult.Cases {
				testCaseCount++
				if caseResult.Passed {
					passedTestCaseCount++
				} else {
					allCasesPassed = false
				}
				
				totalTiming += caseResult.Timing
				
				jsonCase := JSONTestCase{
					Name:           caseName,
					Passed:         caseResult.Passed,
					Timing:         caseResult.Timing,
					FailureReasons: caseResult.FailureReasons,
				}
				
				jsonSuite.Cases = append(jsonSuite.Cases, jsonCase)
			}
			
			if allCasesPassed {
				totalPassedSuites++
			}
			
			jsonDef.Suites = append(jsonDef.Suites, jsonSuite)
		}
		
		jsonResult.TestDefinitions = append(jsonResult.TestDefinitions, jsonDef)
	}
	
	// Fill in summary
	jsonResult.Summary = JSONSummary{
		TotalTestDefinitions: len(results),
		TotalSuites:          testSuiteCount,
		PassedSuites:         totalPassedSuites,
		TotalCases:           testCaseCount,
		PassedCases:          passedTestCaseCount,
		TotalTimeMs:          totalTiming,
	}
	
	// Convert to JSON
	jsonData, err := json.MarshalIndent(jsonResult, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON output: %v\n", err)
		return
	}
	
	// Save to file if output path is set
	if w.OutputPath != "" {
		err = os.WriteFile(w.OutputPath, jsonData, 0644)
		if err != nil {
			fmt.Printf("Error writing JSON results to %s: %v\n", w.OutputPath, err)
		} else {
			fmt.Printf("Test results written to %s\n", w.OutputPath)
		}
	} else {
		// Print to stdout
		fmt.Println(string(jsonData))
	}
}
