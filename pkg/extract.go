package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"bufio"
)

// findControllerForService returns the path to the controller directory for a given service
func findControllerForService(serviceName string) string {
	controllerPath := filepath.Join("..", serviceName+"-controller")
	if _, err := os.Stat(controllerPath); err == nil {
		return controllerPath
	}
	return ""
}

// findOperationInController searches for an operation in the controller's pkg directory
func findOperationInController(serviceName, operationName string) (string, int) {
	controllerPath := findControllerForService(serviceName)
	if controllerPath == "" {
		return "", 0
	}

	pkgPath := filepath.Join(controllerPath, "pkg")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return "", 0
	}

	// Compile regex patterns for operation detection
	patterns := []*regexp.Regexp{
		// RecordAPICall pattern
		regexp.MustCompile(`RecordAPICall\([^,]+,\s*"` + operationName + `"`),
		// rm.sdkapi pattern
		regexp.MustCompile(`rm\.sdkapi\.` + operationName + `\(`),
		// client pattern (for tags, etc.)
		regexp.MustCompile(`client\.` + operationName + `\(`),
	}

	var foundFile string
	var foundLine int

	// Walk through all Go files in pkg directory
	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Open and scan the file
		file, err := os.Open(path)
		if err != nil {
			return nil // Skip files we can't open
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Check each pattern
			for _, pattern := range patterns {
				if pattern.MatchString(line) {
					relPath, _ := filepath.Rel(controllerPath, path)
					foundFile = relPath
					foundLine = lineNum
					return filepath.SkipAll // Stop searching immediately
				}
			}
		}
		return nil
	})

	if err != nil {
		return "", 0
	}

	return foundFile, foundLine
}

