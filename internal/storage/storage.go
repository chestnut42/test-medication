package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Config struct {
	MedicationTable string
}

type Database interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type Service struct {
	cfg      Config
	database Database
}

func NewService(cfg Config, database Database) *Service {
	return &Service{
		cfg:      cfg,
		database: database,
	}
}
