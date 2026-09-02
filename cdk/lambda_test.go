package cdk

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
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

func TestLambdaBuilder_WithPolicy(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackWithPolicy"), nil)

	policyStmt := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("firehose:PutRecord"),
		Resources: jsii.Strings("arn:aws:firehose:us-east-1:123456789012:deliverystream/test-stream"),
	})

	builder := NewLambdaBuilder(stack, "stream-service", "arm64").
		WithPolicy(policyStmt)

	assert.NotNil(t, builder)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]any{
		"PolicyDocument": map[string]any{
			"Statement": []any{
				map[string]any{
					"Action":   "firehose:PutRecord",
					"Effect":   "Allow",
					"Resource": "arn:aws:firehose:us-east-1:123456789012:deliverystream/test-stream",
				},
			},
		},
	})
}

func TestLambdaBuilder_WithPolicyStatement_And_WithPolicies(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackWithMultiplePolicies"), nil)

	stmt1 := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("s3:GetObject"),
		Resources: jsii.Strings("arn:aws:s3:::my-bucket/*"),
	})
	stmt2 := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("sqs:SendMessage"),
		Resources: jsii.Strings("arn:aws:sqs:us-east-1:123456789012:my-queue"),
	})

	builder := NewLambdaBuilder(stack, "worker-service", "arm64").
		WithPolicyStatement(stmt1).
		WithPolicies(stmt2)

	assert.NotNil(t, builder)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]any{
		"PolicyDocument": map[string]any{
			"Statement": []any{
				map[string]any{
					"Action":   "s3:GetObject",
					"Effect":   "Allow",
					"Resource": "arn:aws:s3:::my-bucket/*",
				},
				map[string]any{
					"Action":   "sqs:SendMessage",
					"Effect":   "Allow",
					"Resource": "arn:aws:sqs:us-east-1:123456789012:my-queue",
				},
			},
		},
	})
}

func TestLambdaBuilder_WithRoleAndPolicies(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackCustomRole"), nil)

	customRole := awsiam.NewRole(stack, jsii.String("CustomRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	stmt := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("dynamodb:GetItem"),
		Resources: jsii.Strings("arn:aws:dynamodb:us-east-1:123456789012:table/my-table"),
	})

	builder := NewLambdaBuilder(stack, "custom-role-service", "arm64").
		WithPolicy(stmt).
		WithRole(customRole)

	assert.NotNil(t, builder)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]any{
		"PolicyDocument": map[string]any{
			"Statement": []any{
				map[string]any{
					"Action":   "dynamodb:GetItem",
					"Effect":   "Allow",
					"Resource": "arn:aws:dynamodb:us-east-1:123456789012:table/my-table",
				},
			},
		},
	})
}
