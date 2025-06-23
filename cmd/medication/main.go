package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/chestnut42/test-medication/internal/storage"
)

func main() {
	ctx := context.Background()
	cfg := MustNewConfig()

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	dyn := runDynamo(cfg.DynamoEndpoint, awsCfg)
	store := storage.NewService(storage.Config{
		MedicationTable: cfg.MedicationTable,
	}, dyn)

}

func runDynamo(endpoint string, cfg aws.Config) *dynamodb.Client {
	if endpoint == "" {
		return dynamodb.NewFromConfig(cfg)
	}

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Credentials = credentials.NewStaticCredentialsProvider("dummy", "dummy", "")
	})
}
