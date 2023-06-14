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

$> go run -mod vendor cmd/create-dynamodb-tables/main.go -dynamodb-uri 'awsdynamodb://findingaid?region=us-west-2&endpoint=http://localhost:8000&credentials=static:local:local:local'

$> ./bin/create-dynamodb-tables -dynamodb-uri 'awsdynamodb://findingaid?region=us-west-2&credentials=session'
$> ./bin/populate -producer-uri 'awsdynamodb://findingaid?region=us-west-2&credentials=session&partition_key=id' /usr/local/data/sfomuseum-data-maps/

$> cd /usr/local/whosonfirst/go-reader-findingaid
$> ./bin/read -reader-uri 'findingaid://awsdynamodb/findingaid?region=us-west-2&credentials=session&partition_key=id&template=https://raw.githubusercontent.com/sfomuseum-data/{repo}/main/data/' 1360391327 | jq '.["properties"]["wof:name"]'
"SFO (1988)"

*/

func main() {

	dynamodb_uri := flag.String("dynamodb-uri", "", "A valid aaronland/go-aws-session DSN")
	billing_mode := flag.String("billing-mode", "PAY_PER_REQUEST", "...")
	refresh := flag.Bool("refresh", false, "...")

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
			},
			KeySchema: []*aws_dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       aws.String("HASH"),
				},
			},
			BillingMode: aws.String(*billing_mode),
			TableName:   aws.String(table_name),
		},
	}

	opts := &dynamodb.CreateTablesOptions{
		Tables:  tables,
		Refresh: *refresh,
	}

	err = dynamodb.CreateTables(client, opts)

	if err != nil {
		log.Fatalf("Failed to create tables, %v", err)
	}

}
