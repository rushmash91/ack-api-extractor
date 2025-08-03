package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)


const maxOperationsPerBatch = 100

// ClassifyOperations uses AWS Bedrock Inline Agent to classify operations as control plane vs data plane
func ClassifyOperations(serviceName string, operations []Operation) (*ClassificationResult, error) {
	if len(operations) == 0 {
		return &ClassificationResult{
			ControlPlane: []string{},
			DataPlane:    []string{},
		}, nil
	}

	var operationNames []string
	for _, op := range operations {
		operationNames = append(operationNames, op.Name)
	}

	return classifyInBatches(serviceName, operationNames, maxOperationsPerBatch)
}

// classifyInBatches processes large operation lists in smaller batches
func classifyInBatches(serviceName string, operationNames []string, batchSize int) (*ClassificationResult, error) {
	var allControlPlane []string
	var allDataPlane []string

	for i := 0; i < len(operationNames); i += batchSize {
		end := i + batchSize
		if end > len(operationNames) {
			end = len(operationNames)
		}

		batch := operationNames[i:end]
		fmt.Printf("Processing batch %d/%d (%d operations)\n", 
			(i/batchSize)+1, (len(operationNames)+batchSize-1)/batchSize, len(batch))

		inputText := buildClassificationInput(serviceName, batch)
		response, err := invokeInlineAgent(inputText)
		if err != nil {
			return nil, fmt.Errorf("failed to invoke inline agent for batch %d: %w", (i/batchSize)+1, err)
		}

		result, err := parseClassificationResponse(response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse classification response for batch %d: %w", (i/batchSize)+1, err)
		}

		allControlPlane = append(allControlPlane, result.ControlPlane...)
		allDataPlane = append(allDataPlane, result.DataPlane...)
	}

	return &ClassificationResult{
		ControlPlane: allControlPlane,
		DataPlane:    allDataPlane,
	}, nil
}

// buildClassificationInput creates the input text for operation classification
func buildClassificationInput(serviceName string, operations []string) string {
	operationList := strings.Join(operations, ", ")
	
	prompt := fmt.Sprintf(`You are an AWS architecture expert. Your task is to classify AWS API operations into two categories based on their primary purpose in cloud infrastructure management.

## CLASSIFICATION CATEGORIES:

**CONTROL_PLANE**: Operations that manage the AWS infrastructure itself - creating, configuring, deleting, or modifying AWS resources and their settings. These operations affect the structure, permissions, configuration, or existence of AWS resources.

**DATA_PLANE**: Operations that work with data stored within existing AWS resources. These operations read, write, query, or manipulate application data but do not change the underlying resource configuration.

## DETAILED CLASSIFICATION RULES:

### CONTROL_PLANE Operations:
- **Resource Lifecycle**: Create*, Delete*, Update* operations that manage resource existence
- **Resource Configuration**: Put*Policy, Put*Configuration, Update*Settings, Modify*Attributes
- **Resource Permissions**: Attach*, Detach*, Associate*, Disassociate* permissions/policies
- **Resource Metadata**: Tag/Untag operations, Update*Tags
- **Infrastructure Management**: Enable*, Disable*, Start*, Stop*, Restart* services
- **Access Control**: Operations that grant/revoke access to resources
- **Monitoring Setup**: Put*MetricFilter, Create*Alarm, Put*Retention

### DATA_PLANE Operations:
- **Data Access**: Get*, Describe*, List* data within resources (not resource configuration)
- **Data Manipulation**: Put*, Post*, Update*, Delete* data items/objects (not resources)
- **Data Queries**: Query*, Scan*, Search*, Select* operations
- **Data Streaming**: Read*, Write* streams, Consume*, Produce* messages
- **Data Processing**: Execute*, Invoke*, Process*, Transform* operations on data
- **Data Transfer**: Upload*, Download*, Import*, Export* data content
- **Transactional Operations**: Begin*, Commit*, Rollback* data transactions

## SERVICE-SPECIFIC EXAMPLES:

**DynamoDB**:
- CONTROL_PLANE: CreateTable, DeleteTable, UpdateTable, TagResource, PutItem (creates table structure)
- DATA_PLANE: GetItem, PutItem (inserts data), Query, Scan, UpdateItem (modifies data), DeleteItem (removes data)

**S3**:
- CONTROL_PLANE: CreateBucket, DeleteBucket, PutBucketPolicy, PutBucketEncryption, PutBucketVersioning
- DATA_PLANE: GetObject, PutObject, DeleteObject, ListObjects, CopyObject, HeadObject

**IAM**:
- CONTROL_PLANE: CreateRole, DeleteRole, AttachRolePolicy, CreateUser, CreatePolicy, TagRole
- DATA_PLANE: GetUser, GetRole, ListUsers, ListRoles, GetPolicy (reading existing configurations)

**Lambda**:
- CONTROL_PLANE: CreateFunction, DeleteFunction, UpdateFunctionCode, PutProvisionedConcurrencyConfig
- DATA_PLANE: Invoke, InvokeAsync (executing the function with data)

**EC2**:
- CONTROL_PLANE: RunInstances, TerminateInstances, CreateSecurityGroup, AuthorizeSecurityGroupIngress
- DATA_PLANE: DescribeInstances, DescribeImages, GetConsoleOutput (reading instance data)

**RDS**:
- CONTROL_PLANE: CreateDBInstance, DeleteDBInstance, ModifyDBInstance, CreateDBSnapshot
- DATA_PLANE: DescribeDBInstances, DescribeDBSnapshots (reading database metadata)

## EDGE CASES AND GUIDANCE:

1. **Describe Operations**: 
   - CONTROL_PLANE if describing resource configuration (DescribeTable schema, DescribeSecurityGroups)
   - DATA_PLANE if describing data content (DescribeStream data, DescribeLogEvents)

2. **List Operations**:
   - CONTROL_PLANE if listing resources (ListTables, ListBuckets, ListFunctions)
   - DATA_PLANE if listing data within resources (ListObjects in bucket, ListStreams data)

3. **Update Operations**:
   - CONTROL_PLANE if updating resource configuration (UpdateTable provisioning, UpdateFunctionConfiguration)
   - DATA_PLANE if updating data content (UpdateItem in table, UpdateRecord in stream)

4. **Ambiguous Cases**: When in doubt, classify as DATA_PLANE as these operations are typically more common.

## TASK:
Classify these %s service operations: %s

## OUTPUT FORMAT:
Respond with ONLY valid JSON in exactly this format:
{
  "control_plane": ["operation1", "operation2"],
  "data_plane": ["operation3", "operation4"]
}

Ensure every operation from the input list appears in exactly one category. Do not add explanations or additional text.`, serviceName, operationList)

	return prompt
}

// invokeInlineAgent creates and invokes an inline Bedrock agent for operation classification
func invokeInlineAgent(inputText string) (string, error) {
	ctx := context.Background()
	
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Bedrock Agent Runtime client
	client := bedrockagentruntime.NewFromConfig(cfg)

	// Invoke the inline agent
	result, err := client.InvokeInlineAgent(ctx, &bedrockagentruntime.InvokeInlineAgentInput{
		FoundationModel: aws.String("us.anthropic.claude-3-5-sonnet-20241022-v2:0"),
		Instruction: aws.String(`You are an AWS architecture expert specialized in classifying AWS API operations.
Your task is to classify AWS API operations into two categories:
1. CONTROL_PLANE: Operations that manage AWS infrastructure (create, configure, delete resources)  
2. DATA_PLANE: Operations that work with data within existing resources

Respond with ONLY valid JSON in this format:
{
  "control_plane": ["operation1", "operation2"],
  "data_plane": ["operation3", "operation4"] 
}

Ensure every operation from the input list appears in exactly one category.`),
		AgentName:   aws.String("OperationClassifier"),
		InputText:   aws.String(inputText),
		SessionId:   aws.String("classification-session"),
		EnableTrace: aws.Bool(false),
	})

	if err != nil {
		return "", fmt.Errorf("failed to invoke inline agent: %w", err)
	}

	// Extract text from the response stream
	var responseText strings.Builder
	for event := range result.GetStream().Events() {
		if chunk, ok := event.(*types.InlineAgentResponseStreamMemberChunk); ok {
			if chunk.Value.Bytes != nil {
				responseText.Write(chunk.Value.Bytes)
			}
		}
	}

	if err := result.GetStream().Err(); err != nil {
		return "", fmt.Errorf("error reading stream: %w", err)
	}

	return responseText.String(), nil
}

// parseClassificationResponse parses the JSON response from Bedrock
func parseClassificationResponse(response string) (*ClassificationResult, error) {
	response = strings.TrimSpace(response)
	
	start := strings.Index(response, "{")
	if start == -1 {
		return nil, fmt.Errorf("no valid JSON found in response: %s", response)
	}
	
	end := strings.LastIndex(response, "}")
	if end == -1 || end <= start {
		return nil, fmt.Errorf("incomplete JSON in response: %s", response)
	}
	
	jsonStr := response[start : end+1]
	
	var result ClassificationResult
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w, response: %s", err, jsonStr)
	}

	return &result, nil
}

// ApplyClassification applies the classification results to operations
func ApplyClassification(operations []Operation, classification *ClassificationResult) []Operation {
	controlPlaneMap := make(map[string]bool)
	dataPlaneMap := make(map[string]bool)
	
	for _, op := range classification.ControlPlane {
		controlPlaneMap[op] = true
	}
	for _, op := range classification.DataPlane {
		dataPlaneMap[op] = true
	}

	// Apply classification to operations
	for i := range operations {
		if controlPlaneMap[operations[i].Name] {
			operations[i].Type = "control_plane"
		} else if dataPlaneMap[operations[i].Name] {
			operations[i].Type = "data_plane"
		} else {
			// Default to data_plane if not found
			operations[i].Type = "data_plane"
		}
	}

	return operations
}

// CountControlPlaneOperations counts control plane operations and how many are supported
func CountControlPlaneOperations(operations []Operation) (controlPlane int, supportedControlPlane int) {
	for _, op := range operations {
		if op.Type == "control_plane" {
			controlPlane++
			// Count as supported if it has file and line info (implemented in controller)
			if op.File != "" && op.Line > 0 {
				supportedControlPlane++
			}
		}
	}
	return controlPlane, supportedControlPlane
}
