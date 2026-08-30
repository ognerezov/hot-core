package cdk

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

var (
	userLambdaExecutionRolesMu sync.RWMutex
	userLambdaExecutionRoles   = make(map[string]*awsiam.Role)
	UserLambdaExecutionRole    *awsiam.Role
	PoolAccessRole             *awsiam.Role
)

// getArchitecture maps string architecture names to awslambda.Architecture types.
func getArchitecture(arch string) awslambda.Architecture {
	if arch == "amd64" {
		return awslambda.Architecture_X86_64()
	}
	return awslambda.Architecture_ARM_64()
}

// LambdaBuilder facilitates the creation of AWS Lambda functions with consistent configurations.
type LambdaBuilder struct {
	stack          constructs.Construct
	projectName    string
	architecture   string
	vpc            awsec2.IVpc
	role           awsiam.IRole
	env            *map[string]*string
	timeoutSeconds int
}

// NewLambdaBuilder initializes a new LambdaBuilder with mandatory project name and architecture.
func NewLambdaBuilder(stack constructs.Construct, projectName string, arch string) *LambdaBuilder {
	return &LambdaBuilder{
		stack:          stack,
		projectName:    projectName,
		architecture:   arch,
		timeoutSeconds: 30,
		role:           *GetLambdaExecutionRole(stack, projectName),
	}
}

// NewPollLambdaBuilder initializes a new LambdaBuilder with mandatory project name and architecture.
func NewPollLambdaBuilder(stack constructs.Construct, projectName string, arch string, arn string) *LambdaBuilder {
	return &LambdaBuilder{
		stack:          stack,
		projectName:    projectName,
		architecture:   arch,
		timeoutSeconds: 30,
		role:           *GetLambdaPoolRole(stack, arn),
	}
}

// WithVpc sets the VPC for the Lambda function.
func (b *LambdaBuilder) WithVpc(vpc awsec2.IVpc) *LambdaBuilder {
	b.vpc = vpc
	return b
}

// WithRole sets a custom execution role for the Lambda function.
func (b *LambdaBuilder) WithRole(role awsiam.IRole) *LambdaBuilder {
	b.role = role
	return b
}

// WithEnv sets the environment variables for the Lambda function.
func (b *LambdaBuilder) WithEnv(env *map[string]*string) *LambdaBuilder {
	b.env = env
	return b
}

// WithTimeout sets the timeout for the Lambda function.
func (b *LambdaBuilder) WithTimeout(seconds int) *LambdaBuilder {
	b.timeoutSeconds = seconds
	return b
}

// Build creates a standard Lambda function using the builder's configuration.
func (b *LambdaBuilder) Build(functionName string, fileName string) awslambda.Function {
	props := &awslambda.FunctionProps{
		FunctionName: jsii.String(functionName),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Architecture: getArchitecture(b.architecture),
		Code: awslambda.Code_FromAsset(jsii.String("./"), &awss3assets.AssetOptions{
			Bundling: GetCachedBundlingOptions(fileName, b.architecture, b.projectName),
		}),
		Environment: b.env,
		Role:        b.role,
		Timeout:     awscdk.Duration_Seconds(jsii.Number(b.timeoutSeconds)),
	}

	if b.vpc != nil {
		props.Vpc = b.vpc
	}

	return awslambda.NewFunction(b.stack, jsii.String(functionName), props)
}

// toPascalCase converts a hyphen, underscore, or space separated string into PascalCase (UpperCamelCase).
func toPascalCase(s string) string {
	var parts []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	}) {
		if len(p) > 0 {
			r := []rune(p)
			r[0] = unicode.ToUpper(r[0])
			parts = append(parts, string(r))
		}
	}
	return strings.Join(parts, "")
}

// getLambdaExecutionRole creates the default execution role for Lambda functions.
func getLambdaExecutionRole(stack constructs.Construct, projectName string) *awsiam.Role {
	roleName := "UserLambdaExecutionRole"
	if projectName != "" {
		roleName = fmt.Sprintf("%sUserLambdaExecutionRole", toPascalCase(projectName))
	}
	lambdaRole := awsiam.NewRole(stack, jsii.String(roleName), &awsiam.RoleProps{
		RoleName:    jsii.String(roleName),
		AssumedBy:   awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		Description: jsii.String("Hot Core execution role for Lambda functions"),
	})
	lambdaRole.AddManagedPolicy(
		awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
	)
	return &lambdaRole
}

// GetLambdaExecutionRole returns the singleton UserLambdaExecutionRole (or project-specific execution role), creating it if necessary.
func GetLambdaExecutionRole(stack constructs.Construct, projectName ...string) *awsiam.Role {
	proj := ""
	if len(projectName) > 0 {
		proj = projectName[0]
	}

	userLambdaExecutionRolesMu.Lock()
	defer userLambdaExecutionRolesMu.Unlock()

	if role, exists := userLambdaExecutionRoles[proj]; exists && role != nil {
		return role
	}

	role := getLambdaExecutionRole(stack, proj)
	userLambdaExecutionRoles[proj] = role
	UserLambdaExecutionRole = role
	return role
}

// getLambdaPoolRole creates the execution role with cognito pool access for Lambda functions.
func getLambdaPoolRole(stack constructs.Construct, arn string) *awsiam.Role {
	roleName := "pool-access-lambda-role"
	lambdaRole := awsiam.NewRole(stack, jsii.String(roleName), &awsiam.RoleProps{
		RoleName:    jsii.String(roleName),
		AssumedBy:   awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		Description: jsii.String("Hot Core execution role for social app Lambda functions with Cognito pool access"),
	})
	lambdaRole.AddManagedPolicy(
		awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
	)
	userPool := "*"
	if arn != "" {
		userPool = arn
	}
	lambdaRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("cognito-idp:AdminGetUser", "cognito-idp:AdminCreateUser", "cognito-idp:AdminSetUserPassword", "cognito-idp:ListUsers"),
		Resources: jsii.Strings(userPool),
	}))
	return &lambdaRole
}

// GetLambdaPoolRole returns the singleton PoolAccessRole, creating it if necessary.
func GetLambdaPoolRole(stack constructs.Construct, arn string) *awsiam.Role {
	if PoolAccessRole != nil {
		return PoolAccessRole
	}

	PoolAccessRole = getLambdaPoolRole(stack, arn)
	return PoolAccessRole
}
