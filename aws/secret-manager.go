package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/ognerezov/hot-core/console"
	"github.com/ognerezov/hot-core/tools"

	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/jsii-runtime-go"
)

// DbSecret represents the credentials for a database stored in Secrets Manager.
type DbSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	smClient *secretsmanager.Client
)

func getSmClient() *secretsmanager.Client {
	if smClient != nil {
		return smClient
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %v", err))
	}

	smClient = secretsmanager.NewFromConfig(cfg)
	return smClient
}

// LoadSecret retrieves a secret string from AWS Secrets Manager by its name.
func LoadSecret(secretName string) (*string, error) {
	client := getSmClient()

	result, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: jsii.String(secretName),
	})
	if err != nil {
		return nil, err
	}

	return result.SecretString, nil
}

// SecretExists checks if a secret with the given name exists in AWS Secrets Manager.
func SecretExists(secretName string) (bool, error) {
	_, err := LoadSecret(secretName)
	if err != nil {
		var nsr *types.ResourceNotFoundException
		if errors.As(err, &nsr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateSecret creates a new secret in AWS Secrets Manager.
func CreateSecret(secretName string, value *string) error {
	client := getSmClient()
	_, err := client.CreateSecret(context.Background(), &secretsmanager.CreateSecretInput{
		Name:         jsii.String(secretName),
		SecretString: value,
	})
	return err
}

// AnyFromSecret retrieves a secret and unmarshals it into the provided output structure.
func AnyFromSecret[T any](secretName string, out *T) error {
	secret, err := LoadSecret(secretName)
	if err != nil {
		return err
	}

	return tools.AnyFromString(secret, out)
}

// SaveSecret updates the value of an existing secret in AWS Secrets Manager.
func SaveSecret(secretName string, value *string) error {
	client := getSmClient()

	_, err := client.PutSecretValue(context.Background(), &secretsmanager.PutSecretValueInput{
		SecretId:     jsii.String(secretName),
		SecretString: value,
	})
	return err
}

// SaveFileAsSecret reads a file and saves its content as a secret in AWS Secrets Manager.
func SaveFileAsSecret(secretName string, fileName string) error {
	data, err := tools.ReadFile(fileName)
	if err != nil {
		return err
	}
	str := string(data)
	client := getSmClient()
	_, err = client.PutSecretValue(context.Background(), &secretsmanager.PutSecretValueInput{
		SecretId:     jsii.String(secretName),
		SecretString: jsii.String(str),
	})
	return err
}

// GetDbCredentials retrieves database credentials from a secret by its ARN.
func GetDbCredentials(secretArn string) (*DbSecret, error) {
	var sec DbSecret
	err := AnyFromSecret(secretArn, &sec)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

// SaveSecretValue reads a file and either creates or updates a secret with the file's content.
// The secret name is derived from the file name.
func SaveSecretValue(filePath string) error {
	ext := filepath.Ext(filePath)
	extLower := strings.ToLower(ext)
	if extLower != ".json" && extLower != ".p8" && extLower != ".p12" {
		console.RedPrintln("Error: File must have .json, .p8 or .p12 extension")
		return fmt.Errorf("invalid file extension")
	}

	fileName := filepath.Base(filePath)
	secretName := strings.TrimSuffix(fileName, ext)

	content, err := tools.ReadFile(filePath)
	if err != nil {
		console.RedPrintln(fmt.Sprintf("Error reading file: %v", err))
		return err
	}
	secretValue := string(content)

	exists, err := SecretExists(secretName)
	if err != nil {
		console.RedPrintln(fmt.Sprintf("Error checking secret existence: %v", err))
		return err
	}

	if exists {
		console.YellowPrintln(fmt.Sprintf("Secret '%s' already exists.", secretName))
		confirm := console.ReadStr("Do you want to overwrite it? (yes/no):")
		if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
			console.BluePrintln("Operation cancelled.")
			return fmt.Errorf("operation cancelled")
		}

		err = SaveSecret(secretName, jsii.String(secretValue))
		if err != nil {
			console.RedPrintln(fmt.Sprintf("Error updating secret: %v", err))
			return err
		}
		console.GreenPrintln(fmt.Sprintf("Secret '%s' successfully updated.", secretName))
	} else {
		err = CreateSecret(secretName, jsii.String(secretValue))
		if err != nil {
			console.RedPrintln(fmt.Sprintf("Error creating secret: %v", err))
			return err
		}
		console.GreenPrintln(fmt.Sprintf("Secret '%s' successfully created.", secretName))
	}
	return nil
}
