package runner

import (
	"fmt"
	"path/filepath"

	"github.com/mrfoh/httpprobe/internal/tests"
	"go.uber.org/zap"
)

// executeHooks runs the specified hook test definitions and returns any exported variables
func (r *Runner) executeHooks(hookPaths []string, parentVars map[string]tests.Variable) (map[string]tests.Variable, error) {
	// Result map to collect variables from all hooks
	hookVars := make(map[string]tests.Variable)

	// Execute each hook sequentially
	for _, hookPath := range hookPaths {
		// Resolve the hook path relative to the current working directory
		absPath, err := filepath.Abs(hookPath)
		if err != nil {
			r.Logger.Warn("Failed to resolve hook path", zap.String("path", hookPath), zap.Error(err))
			continue
		}

		// Check if we've already processed this hook to prevent infinite recursion
		r.hooksMutex.Lock()
		alreadyProcessed := r.processedHooks[absPath]
		r.hooksMutex.Unlock()
		
		if alreadyProcessed {
			r.Logger.Warn("Skipping already processed hook to prevent recursion", zap.String("path", absPath))
			continue
		}

		// Process the hook file
		r.Logger.Debug("Loading hook", zap.String("path", absPath))
		hookDef, err := r.processTestFile(absPath)
		if err != nil {
			return hookVars, fmt.Errorf("error loading hook %s: %v", hookPath, err)
		}

		// Merge parent variables with hook variables
		if hookDef.Variables == nil {
			hookDef.Variables = make(map[string]tests.Variable)
		}
		for k, v := range parentVars {
			// Don't override hook-specific variables
			if _, exists := hookDef.Variables[k]; !exists {
				hookDef.Variables[k] = v
			}
		}

		// Execute the hook
		r.Logger.Debug("Executing hook", zap.String("name", hookDef.Name), zap.String("path", absPath))
		hookResult, err := r.executeTestDefinition(hookDef)
		if err != nil {
			return hookVars, fmt.Errorf("error executing hook %s: %v", hookPath, err)
		}

		// Extract variables from hook execution
		for _, suiteResult := range hookResult.Suites {
			if suiteResult.Variables != nil {
				for k, v := range suiteResult.Variables {
					hookVars[k] = v
				}
			}
		}
	}

	return hookVars, nil
}