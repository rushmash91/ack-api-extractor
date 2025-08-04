# ACK API Extractor

A tool for extracting and analyzing AWS API operations from service models, designed to support the AWS Controllers for Kubernetes (ACK) project.

## Features

- Parses AWS service API models to extract all available operations
- Checks ACK controller codebases to identify which operations are implemented
- Uses AWS Bedrock to classify operations as control plane vs data plane (optional)
- Generates IAM policies for supported operations (optional)
- Process multiple AWS services in a single run
- Outputs detailed metadata in JSON format for further analysis

## Prerequisites

- Go 1.22 or later
- AWS credentials configured (required for Bedrock classification)
- Access to AWS service model files (expects `../api-models-aws/models/` directory)
- Access to ACK controller directories (expects `../<service>-controller/` directories)

## Usage

### Multiple Services

Process multiple services at once:

```bash
go run main.go --service=dynamodb,lambda,s3 --output=./results
```

### With Classification

Enable Bedrock-powered operation classification:

```bash
go run main.go --service=dynamodb --output=./results --classify
```

### With IAM Policy Generation

Generate recommended IAM policies for supported operations:

```bash
go run main.go --service=dynamodb --output=./results --generate-policies
```

### Combined Features

Use classification and policy generation together:

```bash
go run main.go --service=dynamodb --output=./results --classify --generate-policies
```

### Command Line Options

- `--service`: AWS service name(s), comma-separated (required)
- `--output`: Output directory for JSON files (required)  
- `--classify`: Enable AWS Bedrock classification of operations (optional)
- `--generate-policies`: Generate recommended IAM policies for supported operations (optional)

## Output Format

### Operations JSON

The tool generates operations JSON files (`<service>-operations.json`) with the following structure:

```json
{
  "service_name": "dynamodb",
  "total_operations": 42,
  "supported_operations": 28,
  "control_plane_operations": 15,
  "supported_control_plane_operations": 12,
  "operations": [
    {
      "name": "CreateTable",
      "type": "control_plane",
      "file": "pkg/resource/table/hooks.go",
      "line": 145
    },
    {
      "name": "GetItem",
      "type": "data_plane",
      "file": "",
      "line": 0
    }
  ]
}
```

#### Field Descriptions

- `service_name`: AWS service identifier
- `total_operations`: Total number of operations found in API model
- `supported_operations`: Number of operations implemented in ACK controller
- `control_plane_operations`: Number of control plane operations (when classification enabled)
- `supported_control_plane_operations`: Number of implemented control plane operations
- `operations`: Array of operation details with implementation status

### IAM Policy JSON

When `--generate-policies` is enabled, the tool also generates IAM policy JSON files (`<service>-policy.json`) with the following structure:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:CreateTable",
        "dynamodb:DeleteTable",
        "dynamodb:DescribeTable",
        "dynamodb:UpdateTable"
      ],
      "Resource": "arn:aws:dynamodb:*:*:*"
    }
  ]
}
```

#### IAM Policy Features

- Contains only permissions for **supported operations** (those implemented in the controller)
- Generates standard AWS IAM policy JSON format for direct use

## Operation Classification

The tool uses a two-tier classification approach:

### Automatic Classification
- **Supported Operations**: Operations found in existing controller code are automatically marked as **Control Plane**
- This assumes that implemented operations are inherently control plane by nature

### AWS Bedrock Classification  
When `--classify` is enabled, only **unsupported operations** are sent to AWS Bedrock's Claude model for classification:

- **Control Plane**: Operations that manage AWS infrastructure (create, configure, delete resources)  
- **Data Plane**: Operations that work with data within existing resources

