package cdk

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"
	"github.com/aws/jsii-runtime-go"
)

func CreateOrUpdateSsmParameter(stack awscdk.Stack, paramName string, paramValue string) {

	paramArn := awscdk.Stack_Of(stack).FormatArn(&awscdk.ArnComponents{
		Service:      jsii.String("ssm"),
		Resource:     jsii.String("parameter"),
		ResourceName: jsii.String(paramName),
	})

	resourceId := fmt.Sprintf("update-%s-param", paramName)

	customresources.NewAwsCustomResource(stack, jsii.String(resourceId), &customresources.AwsCustomResourceProps{
		OnUpdate: &customresources.AwsSdkCall{
			Service: jsii.String("SSM"),
			Action:  jsii.String("putParameter"),
			Parameters: map[string]interface{}{
				"Name":      paramName,
				"Value":     paramValue,
				"Type":      "String",
				"Overwrite": true,
			},
			PhysicalResourceId: customresources.PhysicalResourceId_Of(jsii.String(paramName)),
		},

		Policy: customresources.AwsCustomResourcePolicy_FromStatements(&[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect:    awsiam.Effect_ALLOW,
				Actions:   jsii.Strings("ssm:PutParameter"),
				Resources: &[]*string{paramArn},
			}),
		}),
	})
}
