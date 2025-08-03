package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v3"
)

// ExtractDetailedOperationsFromService extracts operations with metadata structure
func ExtractDetailedOperationsFromService(serviceName string, enableClassification bool) (*ServiceOperations, error) {
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
	supportedCount := 0
	
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
					if file != "" && line > 0 {
						supportedCount++
					}
				}
			}
			break
		}
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("no operations found for service %s", serviceName)
	}

	// Apply classification if enabled
	controlPlaneCount := 0
	supportedControlPlaneCount := 0
	
	if enableClassification {
		classification, err := ClassifyOperations(serviceName, operations)
		if err != nil {
			fmt.Printf("Warning: Failed to classify operations for %s: %v\n", serviceName, err)
		} else {
			operations = ApplyClassification(operations, classification)
			controlPlaneCount, supportedControlPlaneCount = CountControlPlaneOperations(operations)
		}
	}

	return &ServiceOperations{
		ServiceName:              serviceName,
		TotalOperations:          len(operations),
		SupportedOperations:      supportedCount,
		ControlPlaneOps:          controlPlaneCount,
		SupportedControlPlaneOps: supportedControlPlaneCount,
		Operations:               operations,
	}, nil
}

// getModelNameFromController reads the generator.yaml file from a controller and extracts the model_name
func getModelNameFromController(serviceName string) (string, error) {
	controllerPath := findControllerForService(serviceName)
	if controllerPath == "" {
		return "", fmt.Errorf("controller directory not found for service %s", serviceName)
	}
	
	generatorFile := filepath.Join(controllerPath, "generator.yaml")
	if _, err := os.Stat(generatorFile); os.IsNotExist(err) {
		return "", fmt.Errorf("generator.yaml not found in controller directory: %s", generatorFile)
	}
	
	data, err := os.ReadFile(generatorFile)
	if err != nil {
		return "", fmt.Errorf("failed to read generator.yaml file %s: %w", generatorFile, err)
	}
	
	var config GeneratorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse generator.yaml file %s: %w", generatorFile, err)
	}
	
	if config.SDKNames.ModelName == "" {
		return "", fmt.Errorf("model_name not found in generator.yaml file %s", generatorFile)
	}
	
	return config.SDKNames.ModelName, nil
}

// findServiceJSONFile locates the JSON file for a given service in the api-models-aws directory
func findServiceModelJSONFile(serviceName string) (string, error) {
	modelsPath := filepath.Join("..", "api-models-aws", "models", serviceName, "service")
	
	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		// Fallback: try to get the model name from the controller's generator.yaml file
		modelName, fallbackErr := getModelNameFromController(serviceName)
		if fallbackErr != nil {
			return "", fmt.Errorf("service directory not found: %s, and fallback failed: %w", modelsPath, fallbackErr)
		}
		
		// Try with the model name from generator.yaml
		modelsPath = filepath.Join("..", "api-models-aws", "models", modelName, "service")
		if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
			return "", fmt.Errorf("service directory not found for both service name (%s) and model name (%s)", serviceName, modelName)
		}
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
