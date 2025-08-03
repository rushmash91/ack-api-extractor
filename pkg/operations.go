package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)



// ExtractDetailedOperationsFromService extracts operations with metadata structure
func ExtractDetailedOperationsFromService(serviceName string) (*ServiceOperations, error) {
	jsonFile, err := findServiceModelJSONFile(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find JSON file for service %s: %w", serviceName, err)
	}

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file %s: %w", jsonFile, err)
	}

	var model AWSServiceModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse JSON file %s: %w", jsonFile, err)
	}

	// Find the service shape and extract operations
	var operations []Operation
	
	for _, shape := range model.Shapes {
		if shape.Type == "service" && len(shape.Operations) > 0 {
			for _, opTarget := range shape.Operations {
				operationName := extractOperationName(opTarget.Target)
				if operationName != "" {
					file, line := findOperationInController(serviceName, operationName)
					operations = append(operations, Operation{
						Name: operationName,
						Type: "",
						File: file,
						Line: line,
					})
				}
			}
			break
		}
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("no operations found for service %s", serviceName)
	}

	return &ServiceOperations{
		ServiceName: serviceName,
		Operations:  operations,
	}, nil
}

// findServiceJSONFile locates the JSON file for a given service in the api-models-aws directory
func findServiceModelJSONFile(serviceName string) (string, error) {
	modelsPath := filepath.Join("..", "api-models-aws", "models", serviceName, "service")
	
	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("service directory not found: %s", modelsPath)
	}

	var jsonFile string
	err := filepath.Walk(modelsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".json") {
			jsonFile = path
			return filepath.SkipDir // Stop after finding the first JSON file
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error searching for JSON file: %w", err)
	}

	if jsonFile == "" {
		return "", fmt.Errorf("no JSON file found for service %s", serviceName)
	}

	return jsonFile, nil
}

// extractOperationName extracts the operation name from a target string
// Example: "com.amazonaws.acm#DeleteCertificate" -> "DeleteCertificate"
func extractOperationName(target string) string {
	parts := strings.Split(target, "#")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
