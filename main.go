package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	extractor "github.com/aws-controllers-k8s/ack-api-extractor/pkg"
)

func main() {
	servicesFlag := flag.String("service", "", "AWS service name(s), comma-separated (e.g., acm,dynamodb,lambda)")
	outputFlag := flag.String("output", "", "Output directory for files (creates <service>-operations.json)")
	classifyFlag := flag.Bool("classify", false, "Enable AWS Bedrock inline agent classification of operations as control plane vs data plane")
	generatePoliciesFlag := flag.Bool("generate-policies", false, "Generate recommended IAM policies for supported operations")
	flag.Parse()

	if *servicesFlag == "" || *outputFlag == "" {
		fmt.Println("Usage: go run main.go --service=<service1>[,service2,service3...] --output=<directory> [--classify] [--generate-policies]")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go --service=dynamodb --output=./results --classify --generate-policies")
		os.Exit(1)
	}


	// Parse comma-separated services
	services := strings.Split(*servicesFlag, ",")
	for i, service := range services {
		services[i] = strings.TrimSpace(service)
	}
	var features []string
	if *classifyFlag {
		features = append(features, "Bedrock classification")
	}
	if *generatePoliciesFlag {
		features = append(features, "IAM policy generation")
	}
	
	if len(features) > 0 {
		fmt.Printf("Generating files with %s for %d service(s)\n\n", strings.Join(features, " and "), len(services))
	} else {
		fmt.Printf("Generating files for %d service(s)\n\n", len(services))
	}
	
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputFlag, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	
	totalOperations := 0
	successfulServices := 0

	for _, serviceName := range services {
		serviceOps, err := extractor.ExtractDetailedOperationsFromService(serviceName, *classifyFlag)
		if err != nil {
			fmt.Printf("Error extracting operations for %s: %v\n", serviceName, err)
			continue
		}

		if len(serviceOps.Operations) == 0 {
			fmt.Printf("No operations found for %s\n", serviceName)
			continue
		}

		outputFile := fmt.Sprintf("%s/%s-operations.json", *outputFlag, serviceName)
		if writeErr := extractor.WriteServiceOperationsJSON(serviceOps, outputFile); writeErr != nil {
			fmt.Printf("Error writing JSON file for %s: %v\n", serviceName, writeErr)
			continue
		}

		fmt.Printf("%s: %d operations → %s\n", serviceName, len(serviceOps.Operations), outputFile)

		if *generatePoliciesFlag {
			policy, policyErr := extractor.GenerateSinglePolicy(serviceName, serviceOps.Operations)
			if policyErr != nil {
				fmt.Printf("Error generating policy for %s: %v\n", serviceName, policyErr)
			} else {
				if validateErr := extractor.ValidatePolicyJSON(*policy); validateErr != nil {
					fmt.Printf("Warning: Policy validation failed for %s: %v\n", serviceName, validateErr)
				}
				
				policyFile := fmt.Sprintf("%s/%s-policy.json", *outputFlag, serviceName)
				if writePolicyErr := extractor.WritePolicyJSON(policy, policyFile); writePolicyErr != nil {
					fmt.Printf("Error writing policy file for %s: %v\n", serviceName, writePolicyErr)
				} else {
					fmt.Printf("%s: policy → %s\n", serviceName, policyFile)
				}
			}
		}
		totalOperations += len(serviceOps.Operations)
		successfulServices++
	}

	fmt.Printf("\nSuccessfully generated JSON files for %d/%d services\n", successfulServices, len(services))
	fmt.Printf("Total operations extracted: %d\n", totalOperations)
}