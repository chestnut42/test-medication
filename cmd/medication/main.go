package main

import (
	"context"
	"errors"
	"log/slog"
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
	"github.com/chestnut42/test-medication/internal/utils/logx"
	"github.com/chestnut42/test-medication/internal/utils/metrics"
	"github.com/chestnut42/test-medication/internal/utils/signalx"
)

const dynamoPingTimeout = 10 * time.Second

func main() {
	ctx := context.Background()
	cfg := MustNewConfig()

	// logger setup
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx = logx.WithLogger(ctx, logger)

	// Dependencies
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("loading aws config", slog.Any("error", err))
		panic(err)
	}

	dyn := runDynamo(cfg.DynamoEndpoint, awsCfg)
	if err := pingTable(ctx, dyn, cfg.MedicationTable, dynamoPingTimeout); err != nil {
		logx.Logger(ctx).Error("ping table error",
			slog.String("table", cfg.MedicationTable),
			slog.Any("error", ctx.Err()))
		panic(err)
	}

	// Services
	store := storage.NewService(storage.Config{
		MedicationTable: cfg.MedicationTable,
	}, dyn)
	medSvc := medication.NewService(store)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// Running HTTP server
		router := http.NewServeMux()

		// Application
		router.Handle("PUT /v1/medication/{id}", httpmedication.CreateMedication(medSvc))

		// System
		router.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		router.Handle("GET /metrics", metrics.NewHandler())

		logger.Info("running http server", slog.String("addr", cfg.Listen))
		h := httpx.WithLogging(router)
		h = httpx.WithTelemetry(h)
		return httpx.ServeContext(ctx, h, cfg.Listen)
	})
	eg.Go(func() error {
		logger.Info("listening to os signals")
		return signalx.ListenContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	})

	if err := eg.Wait(); err != nil {
		if errors.Is(err, signalx.ErrSignal) {
			logger.Info("signal received", slog.String("signal", err.Error()))
		} else {
			logger.Error("terminated with error", slog.String("error", err.Error()))
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
		_, err := dyn.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(table),
		})
		if err == nil {
			return nil
		}

		logx.Logger(ctx).Error("describe table error",
			slog.String("table", table),
			slog.Any("error", err))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}
