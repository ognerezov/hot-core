package cdk

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
)

func TestGetLambdaExecutionRole_WithProjectName(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackWithProject"), nil)

	role := GetLambdaExecutionRole(stack, "auth-service")
	assert.NotNil(t, role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "AuthServiceUserLambdaExecutionRole",
	})
}

func TestGetLambdaExecutionRole_Default(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackDefault"), nil)

	role := GetLambdaExecutionRole(stack)
	assert.NotNil(t, role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "UserLambdaExecutionRole",
	})
}

func TestNewLambdaBuilder_RoleName(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackBuilder"), nil)

	builder := NewLambdaBuilder(stack, "billing-service", "arm64")
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "BillingServiceUserLambdaExecutionRole",
	})
}

func TestToPascalCase(t *testing.T) {
	assert.Equal(t, "AuthService", toPascalCase("auth-service"))
	assert.Equal(t, "BillingService", toPascalCase("billing_service"))
	assert.Equal(t, "MyProject", toPascalCase("myProject"))
	assert.Equal(t, "Auth", toPascalCase("auth"))
	assert.Equal(t, "", toPascalCase(""))
}
