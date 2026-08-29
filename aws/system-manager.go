package aws

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ognerezov/hot-core/console"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

var (
	ssmClient     *ssm.Client
	ssmClientsMu  sync.RWMutex
	ssmClientsMap = make(map[string]*ssm.Client)
)

// SetSsmClient sets a custom SSM client. Useful for testing.
func SetSsmClient(client *ssm.Client) {
	ssmClientsMu.Lock()
	defer ssmClientsMu.Unlock()
	ssmClient = client
}

// SetSsmClientInRegion sets a custom SSM client for the specified region. Useful for testing.
func SetSsmClientInRegion(region string, client *ssm.Client) {
	if region == "" {
		SetSsmClient(client)
		return
	}
	ssmClientsMu.Lock()
	defer ssmClientsMu.Unlock()
	ssmClientsMap[region] = client
}

// GetSsmClient returns a singleton SSM client, initializing it if necessary.
func GetSsmClient() *ssm.Client {
	ssmClientsMu.Lock()
	defer ssmClientsMu.Unlock()

	if ssmClient != nil {
		return ssmClient
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	ssmClient = ssm.NewFromConfig(cfg)
	return ssmClient
}

func getSsmClient() *ssm.Client {
	return GetSsmClient()
}

// GetSsmClientInRegion returns an SSM client for the specified region.
func GetSsmClientInRegion(region string) *ssm.Client {
	if region == "" {
		return GetSsmClient()
	}

	ssmClientsMu.Lock()
	defer ssmClientsMu.Unlock()

	if client, exists := ssmClientsMap[region]; exists {
		return client
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load AWS config for region %s: %v", region, err)
	}

	client := ssm.NewFromConfig(cfg)
	ssmClientsMap[region] = client
	return client
}

// GetSecureParameter retrieves a secure parameter from AWS SSM Parameter Store in default region.
func GetSecureParameter(name string) (*string, error) {
	return GetSecureParameterInRegion("", name)
}

// GetSecureParameterInRegion retrieves a secure parameter from AWS SSM Parameter Store in the specified region.
func GetSecureParameterInRegion(region string, name string) (*string, error) {
	client := GetSsmClientInRegion(region)
	out, err := client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return out.Parameter.Value, nil
}

// PutParameter uploads a secure parameter to AWS SSM Parameter Store in default region.
func PutParameter(name string, value string) error {
	return PutParameterInRegion("", name, value)
}

// PutParameterInRegion uploads a secure parameter to AWS SSM Parameter Store in the specified region.
func PutParameterInRegion(region string, name string, value string) error {
	client := GetSsmClientInRegion(region)
	_, err := client.PutParameter(context.Background(), &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Overwrite: aws.Bool(true),
		Type:      types.ParameterTypeSecureString,
	})
	console.MagentaPrintln(fmt.Sprintf("Uploaded parameter to region %s", client.Options().Region))
	return err
}
