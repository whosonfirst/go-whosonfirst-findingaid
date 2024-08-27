package dynamodb

import (
	"context"
	"fmt"

	"github.com/aaronland/go-aws-auth"
	aws_dynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewClientWithURI(ctx context.Context, uri string) (*aws_dynamodb.Client, error) {
	return NewClient(ctx, uri)
}

func NewClient(ctx context.Context, uri string) (*aws_dynamodb.Client, error) {
	
	cfg, err := auth.NewConfig(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create config, %w", err)
	}

	client := aws_dynamodb.NewFromConfig(cfg)
	return client, nil
}
