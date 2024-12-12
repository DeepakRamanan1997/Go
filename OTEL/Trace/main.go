package main

import (
        "context"
        "log"
        "math/rand"
        "os"
        "os/signal"
        "syscall"
        "time"

        "go.opentelemetry.io/otel"
        "go.opentelemetry.io/otel/attribute"
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
        "go.opentelemetry.io/otel/sdk/resource"
        sdktrace "go.opentelemetry.io/otel/sdk/trace"
        semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
        "go.opentelemetry.io/otel/trace"
        "google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
)

// setupTraceProvider configures and returns a new trace provider
func setupTraceProvider(jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
        // Create context
        ctx := context.Background()

        // Create gRPC connection
        conn, err := grpc.Dial(
                jaegerEndpoint,
                grpc.WithTransportCredentials(insecure.NewCredentials()),
        )
        if err != nil {
                return nil, err
        }

        // Create OTLP exporter
        exporter, err := otlptracegrpc.New(ctx,
                otlptracegrpc.WithGRPCConn(conn),
        )
        if err != nil {
                return nil, err
        }

        // Create resource
        res, err := resource.New(ctx,
                resource.WithAttributes(
                        semconv.ServiceName("trace-generator"),
                        semconv.ServiceVersion("v0.1.0"),
                ),
        )
        if err != nil {
                return nil, err
        }

        // Create trace provider
        return sdktrace.NewTracerProvider(
                sdktrace.WithBatcher(exporter),
                sdktrace.WithResource(res),
                sdktrace.WithSampler(sdktrace.AlwaysSample()),
        ), nil
}

// generateDatabaseTrace simulates a database operation trace
func generateDatabaseTrace(tracer trace.Tracer) {
        ctx, span := tracer.Start(
                context.Background(),
                "database-operation",
                trace.WithAttributes(
                        attribute.String("db.system", "postgresql"),
                        attribute.String("db.operation", "select"),
                ),
        )
        defer span.End()

        // Simulate some work
        time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

        // Create a child span for query execution
        _, querySpan := tracer.Start(ctx, "db-query-execution")
        time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
        querySpan.End()
}

// generateHttpTrace simulates an HTTP request trace
func generateHttpTrace(tracer trace.Tracer) {
        _, span := tracer.Start(
                context.Background(),
                "http-request",
                trace.WithAttributes(
                        attribute.String("http.method", "GET"),
                        attribute.String("http.url", "/api/users"),
                        attribute.Int("http.status_code", 200),
                ),
        )
        defer span.End()

        // Simulate some work
        time.Sleep(time.Duration(rand.Intn(700)) * time.Millisecond)
}

// generateCacheTrace simulates a cache operation trace
func generateCacheTrace(tracer trace.Tracer) {
        ctx, span := tracer.Start(
                context.Background(),
                "cache-operation",
                trace.WithAttributes(
                        attribute.String("cache.system", "redis"),
                        attribute.String("cache.operation", "get"),
                ),
        )
        defer span.End()

        // Simulate some work
        time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

        // Create a child span for cache lookup
        _, lookupSpan := tracer.Start(ctx, "cache-lookup")
        time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
        lookupSpan.End()
}

func main() {
        // Get Jaeger endpoint from environment variable
        jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
        if jaegerEndpoint == "" {
                log.Fatal("JAEGER_ENDPOINT environment variable is required")
        }

        // Setup trace provider
        traceProvider, err := setupTraceProvider(jaegerEndpoint)
        if err != nil {
                log.Fatalf("Failed to setup trace provider: %v", err)
        }
        defer func() {
                if err := traceProvider.Shutdown(context.Background()); err != nil {
                        log.Printf("Error shutting down trace provider: %v", err)
                }
        }()

        // Set the global trace provider
        otel.SetTracerProvider(traceProvider)

        // Create a tracer
        tracer := traceProvider.Tracer("trace-generator")

        // Trace generation goroutine
        go func() {
                for {
                        // Randomly choose and generate a trace type
                        traceType := rand.Intn(3)
                        switch traceType {
                        case 0:
                                generateDatabaseTrace(tracer)
                        case 1:
                                generateHttpTrace(tracer)
                        case 2:
                                generateCacheTrace(tracer)
                        }

                        // Wait for 5 seconds before next trace
                        time.Sleep(5 * time.Second)
                }
        }()

        // Handle graceful shutdown
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

        // Wait for shutdown signal
        <-sigChan
        log.Println("Received shutdown signal, stopping trace generator...")
}
