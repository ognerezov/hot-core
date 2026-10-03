package cdk

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	UniversalLambdaRoleName = "universal-lambda-execution-role"
)

func CreateUniversalRole(stack constructs.Construct) awsiam.IRole {
	return awsiam.NewRole(stack, jsii.String("UniversalLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(UniversalLambdaRoleName),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			// Логи в CloudWatch + работа внутри VPC
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
			// Трейсинг AWS X-Ray
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AWSXRayDaemonWriteAccess")),
		},
	})
}

func CreateNewRole(stack constructs.Construct, forService string) awsiam.IRole {
	return awsiam.NewRole(stack, jsii.String(fmt.Sprintf("%s-role", forService)), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AWSXRayDaemonWriteAccess")),
		},
	})
}

func GetUniversalRoleArn(stack constructs.Construct) *string {
	return awscdk.Stack_Of(stack).FormatArn(&awscdk.ArnComponents{
		Service:      jsii.String("iam"),
		Region:       jsii.String(""),
		Resource:     jsii.String("role"),
		ResourceName: jsii.String(UniversalLambdaRoleName),
	})
}

func ImportUniversalRole(stack constructs.Construct) awsiam.IRole {
	if existing := stack.Node().TryFindChild(jsii.String("ImportedUniversalLambdaRole")); existing != nil {
		if role, ok := existing.(awsiam.IRole); ok {
			return role
		}
	}
	return awsiam.Role_FromRoleArn(
		stack,
		jsii.String("ImportedUniversalLambdaRole"),
		GetUniversalRoleArn(stack),
		&awsiam.FromRoleArnOptions{
			Mutable: jsii.Bool(false),
		},
	)
}
