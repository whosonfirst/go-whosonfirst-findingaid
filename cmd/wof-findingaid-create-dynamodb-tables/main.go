package main

import (
	"context"
	"flag"
	"log"
	"net/url"

	"github.com/aaronland/go-aws-dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_dynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	aws_dynamodb_types "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

	client, err := dynamodb.NewClient(ctx, *dynamodb_uri)

	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
	}

	u, _ := url.Parse(*dynamodb_uri)
	table_name := u.Host

	var aws_billing_mode aws_dynamodb_types.BillingMode

	switch *billing_mode {
	case "PAY_PER_REQUEST":
		aws_billing_mode = aws_dynamodb_types.BillingModePayPerRequest
	default:
		aws_billing_mode = aws_dynamodb_types.BillingModeProvisioned
	}

	tables := map[string]*aws_dynamodb.CreateTableInput{
		table_name: &aws_dynamodb.CreateTableInput{
			AttributeDefinitions: []aws_dynamodb_types.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: "N",
				},
			},
			KeySchema: []aws_dynamodb_types.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       "HASH",
				},
			},
			BillingMode: aws_billing_mode,
			TableName:   aws.String(table_name),
		},
	}

	opts := &dynamodb.CreateTablesOptions{
		Tables:  tables,
		Refresh: *refresh,
	}

	err = dynamodb.CreateTables(ctx, client, opts)

	if err != nil {
		log.Fatalf("Failed to create tables, %v", err)
	}

}
