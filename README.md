# hot-core

[![Go Version](https://img.shields.io/github/go-mod/go-version/ognerezov/hot-core)](https://go.dev/)
[![License](https://img.shields.io/github/license/ognerezov/hot-core)](LICENSE)

**hot-core** is a lightweight, high-level Go toolkit for AWS CDK and AWS SDK, designed to streamline infrastructure-as-code development and cloud resource management. It provides powerful helpers for Lambda bundling, configuration fetching, and common AWS service interactions.

## Installation

```bash
go get github.com/ognerezov/hot-core
```

## Quick Start

The following example demonstrates how to use `LambdaBuilder` to create an AWS Lambda function with cached Go bundling:

```go
package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/ognerezov/hot-core/cdk"
	"github.com/aws/constructs-go/constructs/v10"
)

func createMyLambda(scope constructs.Construct) {
	// Create a builder with project-wide settings
	builder := cdk.NewLambdaBuilder("my-project", "arm64")
	
	// Build a specific function
	myFunction := builder.Build(scope, "MyFunction", "cmd/handler/main.go")
	
	// Add environment variables if needed
	builder.WithEnv(map[string]*string{
		"STAGE": awscdk.JsiiString("prod"),
	}).Build(scope, "AnotherFunction", "cmd/another/main.go")
}
```

## Features

- **CDK Helpers**: 
    - `LambdaBuilder`: Fluent API for creating Lambda functions.
    - Cached Bundling: Optimized Go build process with Docker and local caching support.
    - Multi-architecture support (`arm64` and `amd64`).
- **AWS Utilities**:
    - **System Manager (SSM)**: Easy parameter retrieval and pre-processing.
    - **Secrets Manager**: Simple secret fetching.
    - **S3 & DynamoDB**: High-level wrappers for common operations.
- **Process & Tools**:
    - **Pre-process**: Dynamic loading of configuration from AWS SSM.
    - **File Detector**: MIME type detection for various file formats (JPEG, PNG, HEIC, MP4, etc.).
    - **Console**: Structured CLI input/output helpers.

## Prerequisites

- **Go**: 1.25.0 or later.
- **AWS CDK**: v2.263.0 or later.
- **Docker**: Required for Lambda bundling (if building via Docker).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
