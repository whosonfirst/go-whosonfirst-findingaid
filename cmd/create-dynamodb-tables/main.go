package doctore

import (
	"flag"
	"github.com/aaronland/go-aws-dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

func main() {

	table_name := flag.String("table-name", "findingaid", "...")
	billing_mode := flag.String("billing-mode", "", "...")

	flag.Parse()

	var client *aws_dynamodb.DynamoDB

	tables := map[string]*aws_dynamodb.CreateTableInput{
		*table_name: &aws_dynamodb.CreateTableInput{
			AttributeDefinitions: []*aws_dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: aws.String("N"),
				},
				{
					AttributeName: aws.String("repo_name"),
					AttributeType: aws.String("S"),
				},
			},
			/*
				KeySchema: []*aws_dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("address"),
						KeyType:       aws.String("HASH"),
					},
				},
				GlobalSecondaryIndexes: []*aws_dynamodb.GlobalSecondaryIndex{
					{
						IndexName: aws.String("status"),
						KeySchema: []*aws_dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("status"),
								KeyType:       aws.String("HASH"),
							},
						},
						Projection: &aws_dynamodb.Projection{
							// maybe just address...?
							ProjectionType: aws.String("ALL"),
						},
					},
				},
			*/
			BillingMode: aws.String(*billing_mode),
			TableName:   aws.String(*table_name),
		},
	}

	opts := &dynamodb.CreateTablesOptions{
		Tables: tables,
	}

	err := dynamodb.CreateTables(client, opts)

	if err != nil {
		log.Fatalf("Failed to create tables, %v", err)
	}

}
