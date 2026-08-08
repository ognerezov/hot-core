package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/ognerezov/hot-core/console"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

var (
	ssmClient *ssm.Client
)

// SetSsmClient sets a custom SSM client. Useful for testing.
func SetSsmClient(client *ssm.Client) {
	ssmClient = client
}

func getSsmClient() *ssm.Client {
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

// GetSecureParameter retrieves a secure parameter from AWS SSM Parameter Store.
func GetSecureParameter(name string) (*string, error) {
	client := getSsmClient()
	out, err := client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return out.Parameter.Value, nil
}

// PutParameter uploads a secure parameter to AWS SSM Parameter Store.
func PutParameter(name string, value string) error {
	client := getSsmClient()
	_, err := client.PutParameter(context.Background(), &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Overwrite: aws.Bool(true),
		Type:      types.ParameterTypeSecureString,
	})
	console.MagentaPrintln(fmt.Sprintf("Uploaded parameter to region %s", client.Options().Region))
	return err
}
