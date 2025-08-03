package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GenerateSinglePolicy creates a single IAM policy for supported operations only
func GenerateSinglePolicy(serviceName string, operations []Operation) (*IAMPolicy, error) {
	var supportedActions []string
	for _, op := range operations {
		if op.File != "" && op.Line > 0 {
			action := mapOperationToIAMAction(serviceName, op.Name)
			supportedActions = append(supportedActions, action)
		}
	}

	if len(supportedActions) == 0 {
		return nil, fmt.Errorf("no supported operations found for service %s", serviceName)
	}

	resourcePattern := generateSimpleResourcePattern(serviceName)
	policy := createPolicy(supportedActions, resourcePattern)

	return &policy, nil
}

// mapOperationToIAMAction converts an AWS operation to IAM action format
func mapOperationToIAMAction(serviceName, operationName string) string {
	modelName, err := getModelNameFromController(serviceName)
	if err != nil {
		modelName = serviceName
	}
	
	servicePrefix := strings.ToLower(modelName)
	return fmt.Sprintf("%s:%s", servicePrefix, operationName)
}

// generateSimpleResourcePattern creates a simple wildcard resource ARN pattern for the service
func generateSimpleResourcePattern(serviceName string) string {
	modelName, err := getModelNameFromController(serviceName)
	if err != nil {
		modelName = serviceName
	}
	
	serviceForARN := strings.ToLower(modelName)

	// TODO -> this is a hack
	switch serviceForARN {
	case "s3":
		// S3 has global ARNs
		return "*"
	case "iam":
		// IAM is global service (no region)
		return "arn:aws:iam::*:*"
	default:
		return fmt.Sprintf("arn:aws:%s:*:*:*", serviceForARN)
	}
}

// createPolicy creates an IAM policy with the given actions and resources
func createPolicy(actions []string, resource string) IAMPolicy {
	if len(actions) == 0 {
		return IAMPolicy{
			Version:   "2012-10-17",
			Statement: []PolicyStatement{},
		}
	}

	return IAMPolicy{
		Version: "2012-10-17",
		Statement: []PolicyStatement{
			{
				Effect:   "Allow",
				Action:   actions,
				Resource: resource,
			},
		},
	}
}

// ValidatePolicyJSON validates that the generated policy is valid JSON
func ValidatePolicyJSON(policy IAMPolicy) error {
	_, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("invalid policy JSON: %w", err)
	}
	
	// Basic validation checks
	if policy.Version == "" {
		return fmt.Errorf("policy Version is required")
	}
	
	if len(policy.Statement) == 0 {
		return fmt.Errorf("policy must have at least one statement")
	}
	
	for i, stmt := range policy.Statement {
		if stmt.Effect != "Allow" && stmt.Effect != "Deny" {
			return fmt.Errorf("statement %d: Effect must be 'Allow' or 'Deny'", i)
		}
		
		if len(stmt.Action) == 0 {
			return fmt.Errorf("statement %d: Action is required", i)
		}
		
		if stmt.Resource == nil {
			return fmt.Errorf("statement %d: Resource is required", i)
		}
	}
	
	return nil
}

// WritePolicyJSON writes a policy to a JSON file
func WritePolicyJSON(policy *IAMPolicy, outputPath string) error {
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policy JSON: %w", err)
	}
	
	return os.WriteFile(outputPath, data, 0644)
}