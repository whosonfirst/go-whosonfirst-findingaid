package main

import (
	"flag"
	"github.com/aaronland/go-aws-dynamodb"
	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"net/url"
)

/*

> go run -mod vendor cmd/create-dynamodb-tables/main.go -dynamodb-uri 'awsdynamodb://findinaid?region=us-west-2&endpoint=http://localhost:8000&credentials=static:local:local:local'
2021/11/04 11:23:03 Failed to create tables, Failed to create table 'findinaid', InvalidParameter: 1 validation error(s) found.
- missing required field, CreateTableInput.KeySchema.
exit status 1

*/

func main() {

	dynamodb_uri := flag.String("dynamodb-uri", "", "A valid aaronland/go-aws-session DSN")
	billing_mode := flag.String("billing-mode", "PAY_PER_REQUEST", "...")

	flag.Parse()

	// START OF put me in aaronland/go-aws-dynamodb

	u, err := url.Parse(*dynamodb_uri)

	if err != nil {
		log.Fatalf("Failed to parse URI, %v", err)
	}

	table_name := u.Host

	q := u.Query()

	// partition_key := q.Get("partition_key")
	region := q.Get("region")
	endpoint := q.Get("endpoint")

	credentials := q.Get("credentials")

	cfg, err := session.NewConfigWithCredentialsAndRegion(credentials, region)

	if err != nil {
		log.Fatalf("Failed to create new session for credentials '%s', %w", credentials, err)
	}

	if endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}

	sess, err := aws_session.NewSession(cfg)

	if err != nil {
		log.Fatalf("Failed to create AWS session, %w", err)
	}

	client := aws_dynamodb.New(sess)

	// END OF put me in aaronland/go-aws-dynamodb

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
