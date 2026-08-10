// Package cdk provides high-level constructs and helpers for AWS CDK.
package cdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/jsii-runtime-go"
)

// TableBuilder helps in constructing a DynamoDB table with a fluent API.
type TableBuilder struct {
	tableName          string
	partitionKey       *awsdynamodb.Attribute
	sortKey            *awsdynamodb.Attribute
	billingMode        awsdynamodb.BillingMode
	removalPolicy      awscdk.RemovalPolicy
	replicationRegions *[]*string
}

// NewTableBuilder creates a new TableBuilder with standard default settings.
// Default settings include:
// - Partition Key: Name="PK", Type=STRING
// - Sort Key: Name="SK", Type=STRING
// - Billing Mode: PAY_PER_REQUEST
// - Removal Policy: RETAIN
// - Replication Regions: eu-south-2
func NewTableBuilder(name string) *TableBuilder {
	return &TableBuilder{
		tableName: name,
		partitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("PK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		sortKey: &awsdynamodb.Attribute{
			Name: jsii.String("SK"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		billingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		removalPolicy: awscdk.RemovalPolicy_RETAIN,
		replicationRegions: &[]*string{
			jsii.String("eu-south-2"),
		},
	}
}

// WithPartitionKey sets the partition key for the table.
func (b *TableBuilder) WithPartitionKey(name string, keyType awsdynamodb.AttributeType) *TableBuilder {
	b.partitionKey = &awsdynamodb.Attribute{
		Name: jsii.String(name),
		Type: keyType,
	}
	return b
}

// WithSortKey sets the sort key for the table.
func (b *TableBuilder) WithSortKey(name string, keyType awsdynamodb.AttributeType) *TableBuilder {
	b.sortKey = &awsdynamodb.Attribute{
		Name: jsii.String(name),
		Type: keyType,
	}
	return b
}

// WithBillingMode sets the billing mode for the table.
func (b *TableBuilder) WithBillingMode(mode awsdynamodb.BillingMode) *TableBuilder {
	b.billingMode = mode
	return b
}

// WithRemovalPolicy sets the removal policy for the table.
func (b *TableBuilder) WithRemovalPolicy(policy awscdk.RemovalPolicy) *TableBuilder {
	b.removalPolicy = policy
	return b
}

// WithReplicationRegions sets the replication regions for the table.
func (b *TableBuilder) WithReplicationRegions(regions *[]*string) *TableBuilder {
	b.replicationRegions = regions
	return b
}

// Build creates a new DynamoDB table using the properties.
func (b *TableBuilder) Build(stack awscdk.Stack) awsdynamodb.ITable {
	return awsdynamodb.NewTable(stack, jsii.String(b.tableName), &awsdynamodb.TableProps{
		TableName:          jsii.String(b.tableName),
		PartitionKey:       b.partitionKey,
		SortKey:            b.sortKey,
		BillingMode:        b.billingMode,
		RemovalPolicy:      b.removalPolicy,
		ReplicationRegions: b.replicationRegions,
	})
}

// DynamoDbTableByName imports an existing DynamoDB table by its name.
func DynamoDbTableByName(stack awscdk.Stack, name string) awsdynamodb.ITable {
	tableArn := awscdk.Stack_Of(stack).FormatArn(&awscdk.ArnComponents{
		Service:  jsii.String("dynamodb"),
		Resource: jsii.String("table/" + name),
	})

	return awsdynamodb.Table_FromTableAttributes(stack, jsii.String(name+"Imported"), &awsdynamodb.TableAttributes{
		TableArn: tableArn,
	})
}
