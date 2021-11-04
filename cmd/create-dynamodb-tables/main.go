package main

import (
	"context"
	"flag"
	"github.com/aaronland/go-aws-dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"net/url"
)

/*

$> go run -mod vendor cmd/create-dynamodb-tables/main.go -dynamodb-uri 'awsdynamodb://findinaid?region=us-west-2&endpoint=http://localhost:8000&credentials=static:local:local:local'

*/

func main() {

	dynamodb_uri := flag.String("dynamodb-uri", "", "A valid aaronland/go-aws-session DSN")
	billing_mode := flag.String("billing-mode", "PAY_PER_REQUEST", "...")

	flag.Parse()

	ctx := context.Background()

	client, err := dynamodb.NewClientWithURI(ctx, *dynamodb_uri)

	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
	}

	u, _ := url.Parse(*dynamodb_uri)
	table_name := u.Host

	tables := map[string]*aws_dynamodb.CreateTableInput{
		table_name: &aws_dynamodb.CreateTableInput{
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
			KeySchema: []*aws_dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       aws.String("HASH"),
				},
			},
			GlobalSecondaryIndexes: []*aws_dynamodb.GlobalSecondaryIndex{
				{
					IndexName: aws.String("status"),
					KeySchema: []*aws_dynamodb.KeySchemaElement{
						{
							AttributeName: aws.String("repo_name"),
							KeyType:       aws.String("HASH"),
						},
					},
					Projection: &aws_dynamodb.Projection{
						// maybe just address...?
						ProjectionType: aws.String("ALL"),
					},
				},
			},
			BillingMode: aws.String(*billing_mode),
			TableName:   aws.String(table_name),
		},
	}

	opts := &dynamodb.CreateTablesOptions{
		Tables: tables,
	}

	err = dynamodb.CreateTables(client, opts)

	if err != nil {
		log.Fatalf("Failed to create tables, %v", err)
	}

}
