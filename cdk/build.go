// Package cdk provides utilities for AWS CDK infrastructure development.
package cdk

import (
	"fmt"
	"os/exec"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

// LocalGoBuilder implements the awscdk.ILocalBundling interface to build Go applications locally.
type LocalGoBuilder struct {
	SourceFile   string
	Architecture string
}

// TryBundle attempts to build the Go application locally.
// It returns true if the build succeeds, otherwise it returns false to fall back to Docker bundling.
func (l *LocalGoBuilder) TryBundle(outputDir *string, options *awscdk.BundlingOptions) *bool {
	if _, err := exec.LookPath("go"); err != nil {
		return jsii.Bool(false)
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		"go mod tidy && GOARCH=%s GOOS=linux go build -tags lambda.norpc -o %s/bootstrap %s.go",
		l.Architecture, *outputDir, l.SourceFile,
	))

	if err := cmd.Run(); err != nil {
		return jsii.Bool(false)
	}

	return jsii.Bool(true)
}

// GetCachedBundlingOptions returns BundlingOptions for a Go Lambda function with caching.
// It accepts the source name, target architecture (arm64 or amd64), and project name for cache path.
func GetCachedBundlingOptions(name string, arch string, projectName string) *awscdk.BundlingOptions {
	if arch != "arm64" && arch != "amd64" {
		arch = "arm64"
	}

	return &awscdk.BundlingOptions{
		Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.25-trixie")),
		Volumes: &[]*awscdk.DockerVolume{
			{
				HostPath:      jsii.String(fmt.Sprintf("/tmp/cdk-go-cache/mod/%s", projectName)),
				ContainerPath: jsii.String("/go/pkg/mod"),
			},
			{
				HostPath:      jsii.String(fmt.Sprintf("/tmp/cdk-go-cache/build/%s", projectName)),
				ContainerPath: jsii.String("/go-build"),
			},
		},

		Environment: &map[string]*string{
			"GOMODCACHE": jsii.String("/go/pkg/mod"),
			"GOCACHE":    jsii.String("/go-build"),
		},
		Command: &[]*string{
			jsii.String("bash"),
			jsii.String("-c"),
			jsii.String(fmt.Sprintf("go mod tidy && GOARCH=%s GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap %s.go", arch, name)),
		},
		Local: &LocalGoBuilder{
			SourceFile:   name,
			Architecture: arch,
		},
	}
}
