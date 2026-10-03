package cdk

import (
	"fmt"
	"sync"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/ognerezov/hot-core/tools"
)

var (
	userLambdaExecutionRolesMu sync.RWMutex
	userLambdaExecutionRoles   = make(map[string]*awsiam.Role)
	UserLambdaExecutionRole    *awsiam.Role
	userLambdaPoolRolesMu      sync.RWMutex
	userLambdaPoolRoles        = make(map[string]*awsiam.Role)
	PoolAccessRole             *awsiam.Role
	projectNameMu              sync.RWMutex
	globalProjectName          string
	ProjectName                string
)

// ResetGlobals resets the cached roles and global project name (useful for testing).
func ResetGlobals() {
	projectNameMu.Lock()
	globalProjectName = ""
	ProjectName = ""
	projectNameMu.Unlock()

	userLambdaExecutionRolesMu.Lock()
	userLambdaExecutionRoles = make(map[string]*awsiam.Role)
	UserLambdaExecutionRole = nil
	userLambdaExecutionRolesMu.Unlock()

	userLambdaPoolRolesMu.Lock()
	userLambdaPoolRoles = make(map[string]*awsiam.Role)
	PoolAccessRole = nil
	userLambdaPoolRolesMu.Unlock()
}

func validateOrSetProjectName(proj string) string {
	projectNameMu.Lock()
	defer projectNameMu.Unlock()

	current := globalProjectName
	if current == "" && ProjectName != "" {
		current = ProjectName
		globalProjectName = ProjectName
	}

	if proj == "" {
		return current
	}

	if current == "" {
		globalProjectName = proj
		ProjectName = proj
		return proj
	}

	if current != proj {
		panic(fmt.Sprintf("projectName mismatch: expected %s, got %s", current, proj))
	}
	return current
}

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
	policies       []awsiam.PolicyStatement
}

// NewLambdaBuilder initializes a new LambdaBuilder with mandatory project name and architecture.
func NewLambdaBuilder(stack constructs.Construct, projectName string, arch string) *LambdaBuilder {
	proj := validateOrSetProjectName(projectName)
	return &LambdaBuilder{
		stack:          stack,
		projectName:    proj,
		architecture:   arch,
		timeoutSeconds: 30,
		role:           ImportUniversalRole(stack),
	}
}

// NewLambdaBuilderWithNewRole initializes a new LambdaBuilder with mandatory project name and architecture.
func NewLambdaBuilderWithNewRole(stack constructs.Construct, projectName string, arch string) *LambdaBuilder {
	proj := validateOrSetProjectName(projectName)
	return &LambdaBuilder{
		stack:          stack,
		projectName:    proj,
		architecture:   arch,
		timeoutSeconds: 30,
		role:           CreateNewRole(stack, fmt.Sprintf("%s-lambda", proj)),
	}
}

// NewPollLambdaBuilder initializes a new LambdaBuilder with mandatory project name and architecture.
func NewPollLambdaBuilder(stack constructs.Construct, projectName string, arch string, arn string) *LambdaBuilder {
	proj := validateOrSetProjectName(projectName)
	return &LambdaBuilder{
		stack:          stack,
		projectName:    proj,
		architecture:   arch,
		timeoutSeconds: 30,
		role:           *GetLambdaPoolRole(stack, arn, proj),
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
	if b.role != nil {
		for _, stmt := range b.policies {
			b.role.AddToPrincipalPolicy(stmt)
		}
	}
	return b
}

// WithNewRole sets a custom execution role for the Lambda function.
func (b *LambdaBuilder) WithNewRole() *LambdaBuilder {
	b.role = CreateNewRole(b.stack, fmt.Sprintf("%s-lambda", b.projectName))
	if b.role != nil {
		for _, stmt := range b.policies {
			b.role.AddToPrincipalPolicy(stmt)
		}
	}
	return b
}

// WithPolicy adds IAM policy statements to the Lambda function's execution role.
func (b *LambdaBuilder) WithPolicy(statements ...awsiam.PolicyStatement) *LambdaBuilder {
	for _, stmt := range statements {
		if stmt != nil {
			if b.role != nil {
				b.role.AddToPrincipalPolicy(stmt)
			}
			b.policies = append(b.policies, stmt)
		}
	}
	return b
}

// WithPolicies adds multiple IAM policy statements to the Lambda function's execution role.
func (b *LambdaBuilder) WithPolicies(statements ...awsiam.PolicyStatement) *LambdaBuilder {
	return b.WithPolicy(statements...)
}

// WithPolicyStatement adds IAM policy statements to the Lambda function's execution role.
func (b *LambdaBuilder) WithPolicyStatement(statements ...awsiam.PolicyStatement) *LambdaBuilder {
	return b.WithPolicy(statements...)
}

// AddToRolePolicy adds a single IAM policy statement to the Lambda function's execution role.
func (b *LambdaBuilder) AddToRolePolicy(statement awsiam.PolicyStatement) *LambdaBuilder {
	return b.WithPolicy(statement)
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
	return tools.ToPascalCase(s)
}

// getLambdaExecutionRole creates the default execution role for Lambda functions.
func getLambdaExecutionRole(stack constructs.Construct, projectName string) *awsiam.Role {
	roleName := fmt.Sprintf("UserLambdaExecutionRole%s", tools.ToPascalCase(*awscdk.Stack_Of(stack).Region()))
	if projectName != "" {
		roleName = fmt.Sprintf("%sUserLambdaExecutionRole%s", tools.ToPascalCase(projectName), tools.ToPascalCase(*awscdk.Stack_Of(stack).Region()))
	}
	lambdaRole := awsiam.NewRole(stack, jsii.String(roleName), &awsiam.RoleProps{
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
	proj = validateOrSetProjectName(proj)

	stackName := *awscdk.Stack_Of(stack).StackName()
	key := stackName

	userLambdaExecutionRolesMu.Lock()
	defer userLambdaExecutionRolesMu.Unlock()

	if role, exists := userLambdaExecutionRoles[key]; exists && role != nil {
		return role
	}

	role := getLambdaExecutionRole(stack, proj)
	userLambdaExecutionRoles[key] = role
	UserLambdaExecutionRole = role
	return role
}

// getLambdaPoolRole creates the execution role with cognito pool access for Lambda functions.
func getLambdaPoolRole(stack constructs.Construct, arn string, projectName string) *awsiam.Role {
	roleName := fmt.Sprintf("PoolAccessLambdaRole%s", tools.ToPascalCase(*awscdk.Stack_Of(stack).Region()))
	if projectName != "" {
		roleName = fmt.Sprintf("%sPoolAccessLambdaRole%s", tools.ToPascalCase(projectName), tools.ToPascalCase(*awscdk.Stack_Of(stack).Region()))
	}
	lambdaRole := awsiam.NewRole(stack, jsii.String(roleName), &awsiam.RoleProps{
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
func GetLambdaPoolRole(stack constructs.Construct, arn string, projectName string) *awsiam.Role {
	proj := validateOrSetProjectName(projectName)

	stackName := *awscdk.Stack_Of(stack).StackName()
	key := fmt.Sprintf("%s:%s", stackName, arn)

	userLambdaPoolRolesMu.Lock()
	defer userLambdaPoolRolesMu.Unlock()

	if role, exists := userLambdaPoolRoles[key]; exists && role != nil {
		return role
	}

	role := getLambdaPoolRole(stack, arn, proj)
	userLambdaPoolRoles[key] = role
	PoolAccessRole = role
	return role
}
