package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"golang.org/x/sync/errgroup"

	"github.com/chestnut42/test-medication/internal/medication"
	"github.com/chestnut42/test-medication/internal/storage"
	httpmedication "github.com/chestnut42/test-medication/internal/transport/http/medication"
	"github.com/chestnut42/test-medication/internal/utils/httpx"
	"github.com/chestnut42/test-medication/internal/utils/signalx"
)

const dynamoPingTimeout = 10 * time.Second

func main() {
	ctx := context.Background()
	cfg := MustNewConfig()

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	dyn := runDynamo(cfg.DynamoEndpoint, awsCfg)
	if err := pingTable(ctx, dyn, cfg.MedicationTable, dynamoPingTimeout); err != nil {
		panic(err)
	}

	store := storage.NewService(storage.Config{
		MedicationTable: cfg.MedicationTable,
	}, dyn)

	medSvc := medication.NewService(store)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// Running HTTP server
		router := http.NewServeMux()
		router.Handle("PUT /v1/medication/{id}", httpmedication.CreateMedication(medSvc))

		return httpx.ServeContext(ctx, router, cfg.Listen)
	})
	eg.Go(func() error {
		return signalx.ListenContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	})

	if err := eg.Wait(); err != nil {
		if errors.Is(err, signalx.ErrSignal) {
			// TODO: TODOLOG
		} else {
			// TODO: TODOLOG
			os.Exit(1)
		}
	}
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

func pingTable(ctx context.Context, dyn *dynamodb.Client, table string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := dyn.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(table),
		}); err != nil {
			// TODO: TODOLOG add log
		}
	}
}
