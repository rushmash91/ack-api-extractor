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
	ServiceName string      `json:"service_name"`
	Operations  []Operation `json:"operations"`
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