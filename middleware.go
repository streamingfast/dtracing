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
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"net/http"

	strackdriverPropagation "contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/streamingfast/logging"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"go.uber.org/zap"
)

var defaultFormat propagation.HTTPFormat = &strackdriverPropagation.HTTPFormat{}
var traceIDGenerator *defaultIDGenerator

func init() {
	traceIDGenerator = &defaultIDGenerator{}
	// initialize traceID and spanID generators.
	var rngSeed int64
	for _, p := range []interface{}{
		&rngSeed, &traceIDGenerator.traceIDAdd, &traceIDGenerator.nextSpanID, &traceIDGenerator.spanIDInc,
	} {
		binary.Read(crand.Reader, binary.LittleEndian, p)
	}

	traceIDGenerator.traceIDRand = rand.New(rand.NewSource(rngSeed))
	traceIDGenerator.spanIDInc |= 1
}

// NewAddTraceIDAwareLoggerMiddleware returns a http.Handler wrapper so that all requests of
// processed by your HTTP handlers will an attached a `zap.Logger` (via the `request.Context()` value, extractable
// using `logging.Logger(ctx, <fallbackLogger>)` that is properly instrumented with the TraceID
// extracted from the actual HTTP request if present.
//
// This handler is aware of the incoming request's trace id, reading it from request headers as configured
// using the Propagation field. The extracted trace id if present is used to configure the actual logger
// with the field `trace_id`.
//
// If the trace id cannot be extracted from the request, a random request id is
// generated and used under the field `trace_id`.
func NewAddTraceIDAwareLoggerMiddleware(next http.Handler, rootLogger *zap.Logger, propagation propagation.HTTPFormat) *addTraceIDMiddleware {
	if rootLogger == nil {
		panic("root logger must not be nil")
	}

	return &addTraceIDMiddleware{
		next:        next,
		rootLogger:  rootLogger,
		propagation: propagation,
	}
}

type addTraceIDMiddleware struct {
	// Handler is the handler used to handle the incoming request.
	next http.Handler

	// Propagation defines how traces are propagated. If unspecified,
	// Stackdriver propagation will be used.
	propagation propagation.HTTPFormat

	// Actual root logger to instrument with request information
	rootLogger *zap.Logger
}

func (h *addTraceIDMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rootLogger := *h.rootLogger
	spanContext, ok := extractSpanContext(r, h.propagation)

	var logger *zap.Logger
	if !ok {
		// Not found in the header, check from the context directly than
		span := trace.FromContext(r.Context())
		if span == nil {
			traceIDField := zap.Stringer("trace_id", traceID(traceIDGenerator.NewTraceID()))
			logger = rootLogger.With(traceIDField)
		} else {
			spanContext := span.SpanContext()
			traceID := hex.EncodeToString(spanContext.TraceID[:])
			logger = rootLogger.With(zap.String("trace_id", traceID))
		}
	} else {
		traceID := hex.EncodeToString(spanContext.TraceID[:])
		logger = rootLogger.With(zap.String("trace_id", traceID))
	}

	ctx := logging.WithLogger(r.Context(), logger)
	h.next.ServeHTTP(w, r.WithContext(ctx))
}

func extractSpanContext(r *http.Request, propagation propagation.HTTPFormat) (trace.SpanContext, bool) {
	if propagation == nil {
		return defaultFormat.SpanContextFromRequest(r)
	}

	return propagation.SpanContextFromRequest(r)
}

type traceID [16]byte

func (t traceID) String() string {
	return hex.EncodeToString(t[:])
}
