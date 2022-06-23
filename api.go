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

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	ttrace "go.opentelemetry.io/otel/trace"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// GetTraceID try to find from the context the correct TraceID associated
// with it. When none is found, returns an randomly generated one.
func GetTraceID(ctx context.Context) (out ttrace.TraceID) {
	span := trace.FromContext(ctx)
	if span == nil {
		return config.Load().(*defaultIDGenerator).NewTraceID()
	}

	out = span.SpanContext().TraceID()
	return
}

// NewRandomTraceID returns a random trace ID using OpenCensus default config IDGenerator.
func NewRandomTraceID() ttrace.TraceID {
	return config.Load().(*defaultIDGenerator).NewTraceID()
}

// NewZeroedTraceID returns a mocked, fixed trace ID containing only 0s.
func NewZeroedTraceID() ttrace.TraceID {
	return NewFixedTraceID("00000000000000000000000000000000")
}

// NewFixedTraceID returns a mocked, fixed trace ID from an hexadecimal string.
// The string in question must be a valid hexadecimal string containing exactly
// 32 characters (16 bytes). Any invalid input results in a panic.
func NewFixedTraceID(hexTraceID string) (out ttrace.TraceID) {
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

// SetupTracing sets up tracers based on the `DTRACING` environment variable.
//
// Options are:
//   - stdout://
//   - cloudtrace://
//
// For cloudtrace, the default sampling rate is 0.25, you can specify it with:
//    cloudtrace://?sample=0.50 (UNIMPLEMENTED!)
//
func SetupTracing(serviceName string, options ...interface{}) error {
	// FIXME(abourget): is `options` still necessary? We want to keep the abstraction
	// to ourselves, I know. So let's not pass any `opentelemetry` stuff upstreams?

	conf := os.Getenv("DTRACING")
	if conf == "" {
		return nil
	}
	u, err := url.Parse(conf)
	if err != nil {
		return fmt.Errorf("parsing env var DTRACING with value %q: %w", conf, err)
	}

	switch u.Scheme {
	case "stdout":
		registerStdout(serviceName, u)
	case "cloudtrace":
		registerCloudTrace(serviceName, u)
	default:
	}

	return nil
}

func registerStdout(serviceName string, u *url.URL) {
	// FIXME(abourget): have all of this depend on `u`

	exp := stdouttrace.New(
		stdouttrace.WithWriter(os.Stderr),
		// Use human readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)

	res := buildResource(serviceName)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
}

func registerCloudTrace(serviceName string, u *url.URL) {
	exp, err := texporter.New()
	if err != nil {
		return err
	}

	res := buildResource(serviceName)

	// FIXME(abourget): use the `sample` querystring param from `u` if specified!
	sampler := trace.TraceIDRatioBased(1 / 4.0)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
}

func buildResource(serviceName string) *resource.Resource {
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			//semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", os.Getenv("NAMESPACE") /* that won't work, whatever */),
		),
	)
	return res
}

func samplerOptionOrDefault(options []interface{}, defaultSampler trace.Sampler) trace.Sampler {
	for _, option := range options {
		if sampler, ok := option.(trace.Sampler); ok {
			return sampler
		}
	}

	return defaultSampler
}
