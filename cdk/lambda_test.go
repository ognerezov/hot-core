package cdk

import (
	"os"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
)

func TestGetLambdaExecutionRole_WithProjectName(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackWithProject"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

	role := GetLambdaExecutionRole(stack, "auth-service")
	assert.NotNil(t, role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "AuthServiceUserLambdaExecutionRoleUsEast1",
	})
}

func TestGetLambdaExecutionRole_Default(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackDefault"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

	role := GetLambdaExecutionRole(stack)
	assert.NotNil(t, role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "UserLambdaExecutionRoleUsEast1",
	})
}

func TestNewLambdaBuilder_RoleName(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackBuilder"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

	builder := NewLambdaBuilder(stack, "billing-service", "arm64")
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "BillingServiceUserLambdaExecutionRoleUsEast1",
	})
}

func TestGetLambdaPoolRole(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackPoolRole"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

	role := GetLambdaPoolRole(stack, "arn:aws:cognito-idp:us-east-1:123456789012:userpool/us-east-1_abc123", "social-app")
	assert.NotNil(t, role)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "SocialAppPoolAccessLambdaRoleUsEast1",
	})
}

func TestToPascalCase(t *testing.T) {
	assert.Equal(t, "AuthService", toPascalCase("auth-service"))
	assert.Equal(t, "BillingService", toPascalCase("billing_service"))
	assert.Equal(t, "MyProject", toPascalCase("myProject"))
	assert.Equal(t, "Auth", toPascalCase("auth"))
	assert.Equal(t, "UsEast1", toPascalCase("us-east-1"))
	assert.Equal(t, "EuWest1", toPascalCase("eu-west-1"))
	assert.Equal(t, "ApNortheast2", toPascalCase("ap-northeast-2"))
	assert.Equal(t, "", toPascalCase(""))
}

func TestLambdaBuilder_Build(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackLambdaBuild"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

	dummyFile := "dummy_test_handler"
	err := os.WriteFile(dummyFile+".go", []byte("package main\n\nfunc main() {}\n"), 0644)
	assert.NoError(t, err)
	defer os.Remove(dummyFile + ".go")

	builder := NewLambdaBuilder(stack, "auth-service", "arm64")
	fn := builder.Build("login-handler", dummyFile)
	assert.NotNil(t, fn)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), map[string]any{
		"FunctionName": "AuthServiceLoginHandlerUsEast1",
	})
}

func TestLambdaBuilder_WithPolicy(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackWithPolicy"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

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
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackWithMultiplePolicies"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

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
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackCustomRole"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

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

func TestProjectName_MismatchPanic(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStackMismatch1"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})

	GetLambdaExecutionRole(stack, "first-project")

	assert.Panics(t, func() {
		GetLambdaExecutionRole(stack, "second-project")
	})
}

func TestGetLambdaExecutionRole_MultipleStacks(t *testing.T) {
	ResetGlobals()
	defer ResetGlobals()

	app := awscdk.NewApp(nil)
	stack1 := awscdk.NewStack(app, jsii.String("StackOne"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-east-1"),
		},
	})
	stack2 := awscdk.NewStack(app, jsii.String("StackTwo"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("us-west-2"),
		},
	})

	role1 := GetLambdaExecutionRole(stack1, "my-app")
	role2 := GetLambdaExecutionRole(stack2, "my-app")

	assert.NotNil(t, role1)
	assert.NotNil(t, role2)
	assert.NotSame(t, role1, role2)

	template1 := assertions.Template_FromStack(stack1, nil)
	template1.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "MyAppUserLambdaExecutionRoleUsEast1",
	})

	template2 := assertions.Template_FromStack(stack2, nil)
	template2.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]any{
		"RoleName": "MyAppUserLambdaExecutionRoleUsWest2",
	})
}
