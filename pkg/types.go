package extractor

// Operation represents a detailed AWS API operation with metadata
type Operation struct {
	Name string `json:"name"`
	Type string `json:"type"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// ServiceOperations represents all operations for a service
type ServiceOperations struct {
	ServiceName                    string      `json:"service_name"`
	TotalOperations                int         `json:"total_operations"`
	SupportedOperations            int         `json:"supported_operations"`
	ControlPlaneOps                int         `json:"control_plane_operations"`
	SupportedControlPlaneOps       int         `json:"supported_control_plane_operations"`
	Operations                     []Operation `json:"operations"`
}

// AWSServiceModel represents the top-level structure of AWS API model JSON files
type AWSServiceModel struct {
	Shapes map[string]ServiceShape `json:"shapes"`
}

// ServiceShape represents a shape in the AWS API model
type ServiceShape struct {
	Type       string            `json:"type"`
	Operations []OperationTarget `json:"operations,omitempty"`
}

// OperationTarget represents an operation reference in the service
type OperationTarget struct {
	Target string `json:"target"`
}

// ClassificationResult represents the result of operation classification
type ClassificationResult struct {
	ControlPlane []string `json:"control_plane"`
	DataPlane    []string `json:"data_plane"`
}

// InlineAgentConfig represents the configuration for an inline agent
type InlineAgentConfig struct {
	FoundationModel string                `json:"foundation_model"`
	Instruction     string                `json:"instruction"`
	AgentName       string                `json:"agent_name"`
	ActionGroups    []InlineActionGroup   `json:"action_groups"`
}

// InlineActionGroup represents an action group for inline agent
type InlineActionGroup struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AgentResponse represents the response from the inline agent
type AgentResponse struct {
	SessionId string `json:"session_id"`
	Trace     string `json:"trace"`
	Output    string `json:"output"`
}

// GeneratorConfig represents the structure of generator.yaml files
type GeneratorConfig struct {
	SDKNames SDKNames `yaml:"sdk_names"`
}

// SDKNames represents the SDK names configuration
type SDKNames struct {
	ModelName string `yaml:"model_name"`
}

// IAMPolicy represents an AWS IAM policy document
type IAMPolicy struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
}

// PolicyStatement represents a single IAM policy statement
type PolicyStatement struct {
	Effect    string      `json:"Effect"`
	Action    []string    `json:"Action"`
	Resource  interface{} `json:"Resource"`
	Condition interface{} `json:"Condition,omitempty"`
}
