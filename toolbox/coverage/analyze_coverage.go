package coverage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AnalyzeCoverage runs tests with coverage and returns the coverage percentage
func AnalyzeCoverage() (float64, error) {
	// Run tests with coverage
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = filepath.Join("h:", "Projects", "toolbox-project")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to run tests: %v\nOutput: %s", err, output)
	}

	// Parse the coverage output
	coverageStr := string(output)
	coverageLines := strings.Split(coverageStr, "\n")
	var coveragePercent float64
	
	for _, line := range coverageLines {
		if strings.Contains(line, "coverage:") {
			fmt.Sscanf(line, "coverage: %f", &coveragePercent)
			break
		}
	}

	// Generate HTML report
	cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o=coverage.html")
	cmd.Dir = filepath.Join("h:", "Projects", "toolbox-project")
	err = cmd.Run()
	if err != nil {
		return coveragePercent, fmt.Errorf("failed to generate HTML report: %v", err)
	}

	return coveragePercent, nil
}

// IdentifyUncoveredFunctions analyzes the coverage report to find uncovered functions
func IdentifyUncoveredFunctions() ([]string, error) {
	// Read the coverage file
	coverageData, err := os.ReadFile(filepath.Join("h:", "Projects", "toolbox-project", "coverage.out"))
	if err != nil {
		return nil, fmt.Errorf("failed to read coverage file: %v", err)
	}

	// Parse the coverage data
	lines := strings.Split(string(coverageData), "\n")
	var uncoveredFunctions []string

	for _, line := range lines {
		if strings.HasPrefix(line, "h:/Projects/toolbox-project/toolbox") && strings.Contains(line, " 0.0%") {
			// Extract function name from the line
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				funcName := strings.TrimSpace(parts[1])
				uncoveredFunctions = append(uncoveredFunctions, funcName)
			}
		}
	}

	return uncoveredFunctions, nil
}

// GenerateTestRecommendations suggests tests for uncovered functions
func GenerateTestRecommendations(uncoveredFunctions []string) []string {
	var recommendations []string

	for _, funcName := range uncoveredFunctions {
		recommendation := fmt.Sprintf("Consider adding tests for function: %s", funcName)
		recommendations = append(recommendations, recommendation)
	}

	return recommendations
}