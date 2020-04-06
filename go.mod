module github.com/dfuse-io/dtracing

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.12.6
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/dfuse-io/logging v0.0.0-20200406213449-45fc25dc6a8d
	github.com/openzipkin/zipkin-go v0.1.6
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.1
	go.uber.org/zap v1.14.0
)

go 1.13
