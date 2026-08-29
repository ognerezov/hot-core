package aws

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	smClient     *secretsmanager.Client
	smClientsMu  sync.RWMutex
	smClientsMap = make(map[string]*secretsmanager.Client)
)

// SetSmClient sets a custom Secrets Manager client. Useful for testing.
func SetSmClient(client *secretsmanager.Client) {
	smClientsMu.Lock()
	defer smClientsMu.Unlock()
	smClient = client
}

// SetSmClientInRegion sets a custom Secrets Manager client for the specified region. Useful for testing.
func SetSmClientInRegion(region string, client *secretsmanager.Client) {
	if region == "" {
		SetSmClient(client)
		return
	}
	smClientsMu.Lock()
	defer smClientsMu.Unlock()
	smClientsMap[region] = client
}

// GetSmClient returns a singleton Secrets Manager client, initializing it if necessary.
func GetSmClient() *secretsmanager.Client {
	smClientsMu.Lock()
	defer smClientsMu.Unlock()

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

func getSmClient() *secretsmanager.Client {
	return GetSmClient()
}

// GetSmClientInRegion returns a Secrets Manager client for the specified region.
func GetSmClientInRegion(region string) *secretsmanager.Client {
	if region == "" {
		return GetSmClient()
	}

	smClientsMu.Lock()
	defer smClientsMu.Unlock()

	if client, exists := smClientsMap[region]; exists {
		return client
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config for Secrets Manager in region %s: %v", region, err))
	}

	client := secretsmanager.NewFromConfig(cfg)
	smClientsMap[region] = client
	return client
}

// LoadSecretInRegion retrieves a secret string from AWS Secrets Manager by its name in the specified region.
func LoadSecretInRegion(region string, secretName string) (*string, error) {
	client := GetSmClientInRegion(region)

	result, err := client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: jsii.String(secretName),
	})
	if err != nil {
		return nil, err
	}

	return result.SecretString, nil
}

// LoadSecret retrieves a secret string from AWS Secrets Manager by its name in the default region.
func LoadSecret(secretName string) (*string, error) {
	return LoadSecretInRegion("", secretName)
}

// SecretExistsInRegion checks if a secret with the given name exists in AWS Secrets Manager in the specified region.
func SecretExistsInRegion(region string, secretName string) (bool, error) {
	_, err := LoadSecretInRegion(region, secretName)
	if err != nil {
		var nsr *types.ResourceNotFoundException
		if errors.As(err, &nsr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SecretExists checks if a secret with the given name exists in AWS Secrets Manager in the default region.
func SecretExists(secretName string) (bool, error) {
	return SecretExistsInRegion("", secretName)
}

// CreateSecretInRegion creates a new secret in AWS Secrets Manager in the specified region.
func CreateSecretInRegion(region string, secretName string, value *string) error {
	client := GetSmClientInRegion(region)
	_, err := client.CreateSecret(context.Background(), &secretsmanager.CreateSecretInput{
		Name:         jsii.String(secretName),
		SecretString: value,
	})
	return err
}

// CreateSecret creates a new secret in AWS Secrets Manager in the default region.
func CreateSecret(secretName string, value *string) error {
	return CreateSecretInRegion("", secretName, value)
}

// AnyFromSecretInRegion retrieves a secret and unmarshals it into the provided output structure from the specified region.
func AnyFromSecretInRegion[T any](region string, secretName string, out *T) error {
	secret, err := LoadSecretInRegion(region, secretName)
	if err != nil {
		return err
	}

	return tools.AnyFromString(secret, out)
}

// AnyFromSecret retrieves a secret and unmarshals it into the provided output structure from the default region.
func AnyFromSecret[T any](secretName string, out *T) error {
	return AnyFromSecretInRegion("", secretName, out)
}

// SaveSecretInRegion updates the value of an existing secret in AWS Secrets Manager in the specified region.
func SaveSecretInRegion(region string, secretName string, value *string) error {
	client := GetSmClientInRegion(region)

	_, err := client.PutSecretValue(context.Background(), &secretsmanager.PutSecretValueInput{
		SecretId:     jsii.String(secretName),
		SecretString: value,
	})
	return err
}

// SaveSecret updates the value of an existing secret in AWS Secrets Manager in the default region.
func SaveSecret(secretName string, value *string) error {
	return SaveSecretInRegion("", secretName, value)
}

// SaveFileAsSecretInRegion reads a file and saves its content as a secret in AWS Secrets Manager in the specified region.
func SaveFileAsSecretInRegion(region string, secretName string, fileName string) error {
	data, err := tools.ReadFile(fileName)
	if err != nil {
		return err
	}
	str := string(data)
	client := GetSmClientInRegion(region)
	_, err = client.PutSecretValue(context.Background(), &secretsmanager.PutSecretValueInput{
		SecretId:     jsii.String(secretName),
		SecretString: jsii.String(str),
	})
	return err
}

// SaveFileAsSecret reads a file and saves its content as a secret in AWS Secrets Manager in the default region.
func SaveFileAsSecret(secretName string, fileName string) error {
	return SaveFileAsSecretInRegion("", secretName, fileName)
}

// GetDbCredentialsInRegion retrieves database credentials from a secret by its ARN in the specified region.
func GetDbCredentialsInRegion(region string, secretArn string) (*DbSecret, error) {
	var sec DbSecret
	err := AnyFromSecretInRegion(region, secretArn, &sec)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

// GetDbCredentials retrieves database credentials from a secret by its ARN in the default region.
func GetDbCredentials(secretArn string) (*DbSecret, error) {
	return GetDbCredentialsInRegion("", secretArn)
}

// SaveSecretValueInRegion reads a file and either creates or updates a secret with the file's content in the specified region.
// The secret name is derived from the file name.
func SaveSecretValueInRegion(region string, filePath string) error {
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

	exists, err := SecretExistsInRegion(region, secretName)
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

		err = SaveSecretInRegion(region, secretName, jsii.String(secretValue))
		if err != nil {
			console.RedPrintln(fmt.Sprintf("Error updating secret: %v", err))
			return err
		}
		console.GreenPrintln(fmt.Sprintf("Secret '%s' successfully updated.", secretName))
	} else {
		err = CreateSecretInRegion(region, secretName, jsii.String(secretValue))
		if err != nil {
			console.RedPrintln(fmt.Sprintf("Error creating secret: %v", err))
			return err
		}
		console.GreenPrintln(fmt.Sprintf("Secret '%s' successfully created.", secretName))
	}
	return nil
}

// SaveSecretValue reads a file and either creates or updates a secret with the file's content in the default region.
// The secret name is derived from the file name.
func SaveSecretValue(filePath string) error {
	return SaveSecretValueInRegion("", filePath)
}
