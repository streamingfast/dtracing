// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dtracing

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"

	"go.uber.org/zap"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

type TraceAttributes map[string]interface{}

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// GetTraceID try to find from the context the correct TraceID associated
// with it. When none is found, returns an randomly generated one.
func GetTraceID(ctx context.Context) (out trace.TraceID) {
	span := trace.FromContext(ctx)
	if span == nil {
		return config.Load().(*defaultIDGenerator).NewTraceID()
	}

	out = span.SpanContext().TraceID
	return
}

// NewRandomTraceID returns a random trace ID using OpenCensus default config IDGenerator.
func NewRandomTraceID() trace.TraceID {
	return config.Load().(*defaultIDGenerator).NewTraceID()
}

// NewZeroedTraceID returns a mocked, fixed trace ID containing only 0s.
func NewZeroedTraceID() trace.TraceID {
	return NewFixedTraceID("00000000000000000000000000000000")
}

// NewFixedTraceID returns a mocked, fixed trace ID from an hexadecimal string.
// The string in question must be a valid hexadecimal string containing exactly
// 32 characters (16 bytes). Any invalid input results in a panic.
func NewFixedTraceID(hexTraceID string) (out trace.TraceID) {
	if len(hexTraceID) != 32 {
		panic(fmt.Errorf("trace id hexadecimal value should have 32 characters, received %d for %q", len(hexTraceID), hexTraceID))
	}

	bytes, err := hex.DecodeString(hexTraceID)
	if err != nil {
		panic(fmt.Errorf("unable to decode hex trace id %q: %s", hexTraceID, err))
	}

	for i := 0; i < 16; i++ {
		out[i] = bytes[i]
	}

	return
}

// NewZeroedTraceIDInContext is similar to NewZeroedTraceID but will actually
// insert the span straight into a context that can later be used
// to ensure the trace id is controlled.
//
// This should be use only in testing to provide a fixed trace ID
// instead of generating a new one each time.
func NewZeroedTraceIDInContext(ctx context.Context) context.Context {
	ctx, _ = trace.StartSpanWithRemoteParent(ctx, "zeroed", trace.SpanContext{
		TraceID: NewZeroedTraceID(),
		SpanID:  config.Load().(*defaultIDGenerator).NewSpanID(),
	})

	return ctx
}

// NewFixedTraceIDInContext is similar to NewFixedTraceID but will actually
// insert the span straight into a context that can later be used
// to ensure the trace id is controlled.
//
// This should be use only in testing to provide a fixed trace ID
// instead of generating a new one each time.
func NewFixedTraceIDInContext(ctx context.Context, hexTraceID string) context.Context {
	ctx, _ = trace.StartSpanWithRemoteParent(ctx, "fixed", trace.SpanContext{
		TraceID: NewFixedTraceID(hexTraceID),
		SpanID:  config.Load().(*defaultIDGenerator).NewSpanID(),
	})

	return ctx
}

// SetupTracing make sensible decision to setup tracing exporters based
// on the environment.
//
// If in production, registers the `StackDriver` exporter. It defines two
// pre-defined attribute for all traces. It defines `serviceName`
// attribute (receiver in parameter here) and it defiens `pod`
// which corresponds to `hostname` resolution.
//
// In development, registers exporters based on environment variables
// "TRACING_ZAP_EXPORTER" (zap exporter) and `TRACING_ZIPKIN_EXPORTER=zipkinURL`
// for Zipkin exporter.
//
// Options:
// - A `trace.Sampler` instance: sets `trace` default config `DefaultSampler` value to this value (defaults `1/4.0`)
// - A `dtracing.TraceAttributes` instance: sets additional StackDriver default attributes (defaults to `nil`)
func SetupTracing(serviceName string, options ...interface{}) error {
	sampler := samplerOptionOrDefault(options, trace.ProbabilitySampler(1/4.0))
	defaultAttributes := traceAttributesOptionOrDefault(options, nil)

	if IsProductionEnvironment() {
		zlog.Info("registering StackDriver exporter")
		return RegisterStackDriverExporter(serviceName, sampler, stackdriver.Options{
			DefaultTraceAttributes: defaultAttributes,
		})
	}

	zlog.Info("registering development exporters from environment variables")
	return RegisterDevelopmentExportersFromEnv(serviceName, sampler)
}

// RegisterStackDriverExporter registers the production `StackDriver` exporter
// for all traces. Uses the `sampler` as the default sampler for all traces.
// The service name is also added a a label to all traces created.
func RegisterStackDriverExporter(serviceName string, sampler trace.Sampler, options stackdriver.Options) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: sampler})

	if options.DefaultTraceAttributes == nil {
		options.DefaultTraceAttributes = map[string]interface{}{}
	}

	options.DefaultTraceAttributes["serviceName"] = serviceName
	options.DefaultTraceAttributes["pod"] = hostname

	zlog.Info("creating StackDriver exporter", zap.Any("default_attributes", options.DefaultTraceAttributes))

	exporter, err := stackdriver.NewExporter(options)
	if err != nil {
		return fmt.Errorf("failed to create StackDriver exporter: %s", err)
	}

	trace.RegisterExporter(exporter)
	return nil
}

// RegisterDevelopmentExportersFromEnv registers exporters based on environment
// variables "TRACING_ZAP_EXPORTER" (zap exporter) and
// `TRACING_ZIPKIN_EXPORTER=zipkinURL` for Zipkin exporter.
func RegisterDevelopmentExportersFromEnv(serviceName string, sampler trace.Sampler) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: sampler})

	zapExporterEnv := os.Getenv("TRACING_ZAP_EXPORTER")
	zipkinExporterEnv := os.Getenv("TRACING_ZIPKIN_EXPORTER")

	if zapExporterEnv != "" {
		zlog.Info("registering zap exporter")
		RegisterZapExporter()
	}

	if zipkinExporterEnv != "" {
		zlog.Info("registering Zipkin exporter", zap.String("url", zipkinExporterEnv))
		err := RegisterZipkinExporter(serviceName, zipkinExporterEnv)
		if err != nil {
			return fmt.Errorf("failed to register ZipKin exporter: %s", err)
		}
	}

	return nil
}

// RegisterZapExporter registers a Zap exporter that exports all traces
// to zlog instance of this package.
func RegisterZapExporter() {
	exporter := new(zapExporter)
	trace.RegisterExporter(exporter)
}

// RegisterZipkinExporter registers a ZipKin exporter that exports all traces
// to a zipkin instance pointed by `zipkinURL`. Note the `zipkinURL` must be
// the full path of the export function.
func RegisterZipkinExporter(serviceName string, zipkinURL string) error {
	_, err := url.Parse(zipkinURL)
	if err != nil {
		return fmt.Errorf("invalid zipkin exporter url: %s", err)
	}

	localEndpoint, err := openzipkin.NewEndpoint(serviceName, "")
	if err != nil {
		return fmt.Errorf("unable to create local endpoint: %s", err)
	}

	reporter := zipkinHTTP.NewReporter(zipkinURL)
	zipkinExporter := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(zipkinExporter)
	return nil
}

// IsProductionEnvironment determines if we are in a production or
// a development environment. When file `/.dockerenv` is present,
// assuming it's production, development otherwise
func IsProductionEnvironment() bool {
	gcp := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	_, err := os.Stat("/.dockerenv")

	return gcp != "" && !os.IsNotExist(err)
}

func samplerOptionOrDefault(options []interface{}, defaultSampler trace.Sampler) trace.Sampler {
	for _, option := range options {
		if sampler, ok := option.(trace.Sampler); ok {
			return sampler
		}
	}

	return defaultSampler
}

func traceAttributesOptionOrDefault(options []interface{}, defaultAttributes TraceAttributes) TraceAttributes {
	for _, option := range options {
		if attributes, ok := option.(TraceAttributes); ok {
			return attributes
		}
	}

	return defaultAttributes
}
