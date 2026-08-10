package cdk

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/jsii-runtime-go"
)

func TestTableBuilder_Build(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStack"), nil)

	NewTableBuilder("MyTable").
		WithPartitionKey("CustomPK", awsdynamodb.AttributeType_BINARY).
		WithSortKey("CustomSK", awsdynamodb.AttributeType_NUMBER).
		WithBillingMode(awsdynamodb.BillingMode_PAY_PER_REQUEST).
		WithRemovalPolicy(awscdk.RemovalPolicy_DESTROY).
		Build(stack)

	template := assertions.Template_FromStack(stack, nil)

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]any{
		"TableName": "MyTable",
		"KeySchema": []any{
			map[string]any{
				"AttributeName": "CustomPK",
				"KeyType":       "HASH",
			},
			map[string]any{
				"AttributeName": "CustomSK",
				"KeyType":       "RANGE",
			},
		},
		"AttributeDefinitions": []any{
			map[string]any{
				"AttributeName": "CustomPK",
				"AttributeType": "B",
			},
			map[string]any{
				"AttributeName": "CustomSK",
				"AttributeType": "N",
			},
		},
		"BillingMode": "PAY_PER_REQUEST",
	})
}

func TestNewTableBuilder(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStack"), nil)

	NewTableBuilder("TestTable").Build(stack)

	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]any{
		"TableName": "TestTable",
		"KeySchema": []any{
			map[string]any{
				"AttributeName": "PK",
				"KeyType":       "HASH",
			},
			map[string]any{
				"AttributeName": "SK",
				"KeyType":       "RANGE",
			},
		},
		"BillingMode": "PAY_PER_REQUEST",
	})
}

func TestDynamoDbTableByName(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := awscdk.NewStack(app, jsii.String("TestStack"), nil)

	table := DynamoDbTableByName(stack, "ExistingTable")
	if table == nil {
		t.Error("expected table to be non-nil")
	}
}
