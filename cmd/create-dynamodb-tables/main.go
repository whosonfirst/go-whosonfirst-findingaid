package doctore

import (
	"flag"
	"github.com/aaronland/go-aws-dynamodb"
	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

func main() {

	aws_dsn := flag.String("aws-dsn", "", "A valid aaronland/go-aws-session DSN")
	table_name := flag.String("table-name", "findingaid", "...")
	billing_mode := flag.String("billing-mode", "PAY_PER_REQUEST", "...")

	flag.Parse()

	sess, err := session.NewSessionWithDSN(*aws_dsn)

	if err != nil {
		log.Fatalf("Failed to create AWS session, %v", err)
	}

	client := aws_dynamodb.New(sess)

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

	err = dynamodb.CreateTables(client, opts)

	if err != nil {
		log.Fatalf("Failed to create tables, %v", err)
	}

}
