module github.com/streamingfast/dtracing

go 1.16

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.10 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.8.2
	github.com/streamingfast/logging v0.0.0-20220304183711-ddba33d79e27
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.uber.org/zap v1.14.0
)
