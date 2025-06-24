package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/chestnut42/test-medication/internal/model"
)

func TestStorage(t *testing.T) {
	// Create a logger that discards all logs (silences testcontainers-go)
	noopLogger := log.New(io.Discard, "", 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Running a container
	ctr, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:2.6.1",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("8000/tcp")),
		testcontainers.WithLogger(noopLogger)) // Remove this to debug containers
	if err != nil {
		t.Fatalf("failed to start dynamodb container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}()

	// Creating dynamo client
	endPoint, err := ctr.PortEndpoint(ctx, "8000/tcp", "http")
	if err != nil {
		t.Fatalf("failed to get port endpoint: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("localhost"))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endPoint)
		o.Credentials = credentials.NewStaticCredentialsProvider("dummy", "dummy", "")
	})

	tests := []struct {
		name string
		test func(t *testing.T, ctx context.Context, service *Service)
	}{
		{name: "testStorage_Create", test: testStorageCreate},
		{name: "testStorage_Get", test: testStorageGet},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// For each test we are creating a separate table so that tests do not interfere with each other
			tableName := test.name + "_table"
			if _, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
				AttributeDefinitions: []types.AttributeDefinition{{
					AttributeName: aws.String("PK"),
					AttributeType: types.ScalarAttributeTypeS,
				}, {
					AttributeName: aws.String("SK"),
					AttributeType: types.ScalarAttributeTypeS,
				}},
				KeySchema: []types.KeySchemaElement{{
					AttributeName: aws.String("PK"),
					KeyType:       types.KeyTypeHash,
				}, {
					AttributeName: aws.String("SK"),
					KeyType:       types.KeyTypeRange,
				}},
				TableName:   aws.String(tableName),
				BillingMode: types.BillingModePayPerRequest,
			}); err != nil {
				t.Fatalf("failed to create table: %s: %v", tableName, err)
			}

			cfg := Config{
				MedicationTable: tableName,
			}
			svc := NewService(cfg, client)
			test.test(t, ctx, svc)
		})
	}
}

func testStorageCreate(t *testing.T, ctx context.Context, service *Service) {
	t.Run("happy create", func(t *testing.T) {
		err := service.CreateMedication(ctx, model.Medication{
			Identity: model.Identity{
				Id:    "some id",
				Owner: "owner",
			},
			MedicationData: model.MedicationData{
				Name:   "my name",
				Dosage: "dosage 500mg",
				Form:   "Plasma",
			},
		})
		if err != nil {
			t.Fatalf("failed to create medication: %v", err)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		err := service.CreateMedication(ctx, model.Medication{
			Identity: model.Identity{
				Id:    "some id", // the same id
				Owner: "owner",
			},
			MedicationData: model.MedicationData{
				Name:   "my other name",
				Dosage: "other dosage 500mg",
				Form:   "Liquid",
			},
		})
		if !errors.Is(err, ErrAlreadyExists) {
			t.Fatalf("creating medication with existing id: want: %v got: %v", ErrAlreadyExists, err)
		}
	})

	// Okay to create the very same medication, but with the new id
	t.Run("okay to create duplicate with different ID", func(t *testing.T) {
		err := service.CreateMedication(ctx, model.Medication{
			Identity: model.Identity{
				Id:    "some id 2",
				Owner: "owner",
			},
			MedicationData: model.MedicationData{
				Name:   "my name",
				Dosage: "dosage 500mg",
				Form:   "Plasma",
			},
		})
		if err != nil {
			t.Fatalf("failed to create medication: %v", err)
		}
	})
}

func testStorageGet(t *testing.T, ctx context.Context, service *Service) {
	expected := model.Medication{
		Identity: model.Identity{
			Id:    "42",
			Owner: "owner",
		},
		MedicationData: model.MedicationData{
			Name:   "my other name",
			Dosage: "other dosage 500mg",
			Form:   "Liquid",
		},
		Version: "some version",
	}
	err := service.CreateMedication(ctx, expected)
	if err != nil {
		t.Fatalf("failed to create medication: %v", err)
	}

	t.Run("ok", func(t *testing.T) {
		got, err := service.GetMedication(ctx, model.Identity{Id: "42", Owner: "owner"})
		if err != nil {
			t.Fatalf("failed to get medication: %v", err)
		}
		if got != expected {
			t.Fatalf("got: %v, expected: %v", got, expected)
		}
	})

	t.Run("not found id", func(t *testing.T) {
		gotMed, err := service.GetMedication(ctx, model.Identity{Id: "43", Owner: "owner"})
		if !errors.Is(err, ErrNotFound) {
			t.Logf("got med: %v", gotMed)
			t.Fatalf("got error: %v, expected: %v", err, ErrNotFound)
		}
	})

	t.Run("not found owner", func(t *testing.T) {
		_, err = service.GetMedication(ctx, model.Identity{Id: "42", Owner: "owner2"})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("got error: %v, expected: %v", err, ErrNotFound)
		}
	})
}
