package extractor

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteServiceOperationsJSON writes service operations to a JSON file
func WriteServiceOperationsJSON(serviceOps *ServiceOperations, outputPath string) error {
	data, err := json.MarshalIndent(serviceOps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return os.WriteFile(outputPath, data, 0644)
}