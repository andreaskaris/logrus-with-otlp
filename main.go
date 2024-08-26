package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
)

var (
	interval     = flag.Int("interval", 1, "Interval between log messages in seconds")
	otlpEndpoint = flag.String("otlp-endpoint", "", "URL of OTLP endpoint - leave empty for logging to STDOUT")
	otlpInsecure = flag.Bool("otlp-insecure", false, "Specify if OTLP connection shall be insecure")
)

func main() {
	flag.Parse()
	fmt.Println("Starting application with: ", *interval, *otlpEndpoint, *otlpInsecure, "gRPC")
	if *otlpEndpoint != "" {
		ctx := context.Background()
		// Create a logger provider.
		// You can pass this instance directly when creating bridges.
		loggerProvider, err := newLoggerProvider(ctx, *otlpEndpoint, *otlpInsecure)
		if err != nil {
			panic(err)
		}
		// Handle shutdown properly so nothing leaks.
		defer loggerProvider.Shutdown(ctx)
		// Register as global logger provider so that it can be accessed global.LoggerProvider.
		// Most log bridges use the global logger provider as default.
		// If the global logger provider is not set then a no-op implementation
		// is used, which fails to generate data.
		global.SetLoggerProvider(loggerProvider)
		// Instrument logrus.
		logrus.AddHook(otellogrus.NewHook(
			"main",
			otellogrus.WithLevels([]logrus.Level{
				logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel},
			),
			otellogrus.WithLoggerProvider(loggerProvider), // Redundant, see comment above, but for illustration purposes.
		))
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	for {
		time.Sleep(time.Duration(*interval) * time.Second)
		logrus.WithFields(logrus.Fields{"key": "value"}).Info("New log message")
	}
}

func newLoggerProvider(ctx context.Context, otlpEndpoint string, otlpInsecure bool) (*log.LoggerProvider, error) {
	grpcExpOpt := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(otlpEndpoint),
	}

	if otlpInsecure {
		grpcExpOpt = append(grpcExpOpt, otlploggrpc.WithInsecure())
	} else {
		return nil, fmt.Errorf("secure OTLP connection not implemented")
	}

	exporter, err := otlploggrpc.New(ctx, grpcExpOpt...)
	if err != nil {
		return nil, err
	}

	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithProcessor(processor),
	)
	return provider, nil
}
