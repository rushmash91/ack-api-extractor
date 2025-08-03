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
	flag.Parse()

	if *servicesFlag == "" || *outputFlag == "" {
		fmt.Println("Usage: go run main.go --service=<service1>[,service2,service3...] --output=<directory>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go --service=acm --output=./results")
		fmt.Println("  go run main.go --service=acm,dynamodb,lambda --output=./results")
		os.Exit(1)
	}


	// Parse comma-separated services
	services := strings.Split(*servicesFlag, ",")
	for i, service := range services {
		services[i] = strings.TrimSpace(service)
	}
	
	fmt.Printf("Generating JSON files for %d service(s)\n\n", len(services))
	
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputFlag, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	
	totalOperations := 0
	successfulServices := 0

	for _, serviceName := range services {
		serviceOps, err := extractor.ExtractDetailedOperationsFromService(serviceName)
		if err != nil {
			fmt.Printf("Error extracting operations for %s: %v\n", serviceName, err)
			continue
		}

		if len(serviceOps.Operations) == 0 {
			fmt.Printf("No operations found for %s\n", serviceName)
			continue
		}

		// Write JSON file
		outputFile := fmt.Sprintf("%s/%s-operations.json", *outputFlag, serviceName)
		if writeErr := extractor.WriteServiceOperationsJSON(serviceOps, outputFile); writeErr != nil {
			fmt.Printf("Error writing JSON file for %s: %v\n", serviceName, writeErr)
			continue
		}

		fmt.Printf("%s: %d operations → %s\n", serviceName, len(serviceOps.Operations), outputFile)
		totalOperations += len(serviceOps.Operations)
		successfulServices++
	}

	fmt.Printf("\nSuccessfully generated JSON files for %d/%d services\n", successfulServices, len(services))
	fmt.Printf("Total operations extracted: %d\n", totalOperations)
}