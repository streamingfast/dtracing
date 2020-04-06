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
	"fmt"

	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

// StartSpan starts a `trace.Span` using a sugaring way of adding attributes. This has
// some drawback since we peform some computation to correctly resolve the attribute to
// their `trace.*Attribute(..., ...)` values.
//
// If you are creating your span in a tight loop, you are better off using `StartSpanA`
// which accepts `trace.Attribute` directly.
func StartSpan(ctx context.Context, name string, keyedAttributes ...interface{}) (context.Context, *trace.Span) {
	return StartSpanWithSamplerA(ctx, name, nil, keyedAttributesToTraceAttributes(keyedAttributes)...)
}

// StartSpanA starts a `trace.Span` which accepts a variadic list of `trace.Attribute` directly.
func StartSpanA(ctx context.Context, name string, attributes ...trace.Attribute) (context.Context, *trace.Span) {
	return StartSpanWithSamplerA(ctx, name, nil, attributes...)
}

// StartSpanWithSampler starts a `trace.Span` just like `StartSpan` accepting the same set of
// arguments alongside a new `sampler`.
func StartSpanWithSampler(ctx context.Context, name string, sampler trace.Sampler, keyedAttributes ...interface{}) (context.Context, *trace.Span) {
	return StartSpanWithSamplerA(ctx, name, sampler, keyedAttributesToTraceAttributes(keyedAttributes)...)
}

// StartSpanWithSamplerA starts a `trace.Span` just like `StartSpanA` accepting the same set of
// arguments alongside a new `sampler` value for the trace.
func StartSpanWithSamplerA(ctx context.Context, name string, sampler trace.Sampler, attributes ...trace.Attribute) (context.Context, *trace.Span) {
	var startOptions []trace.StartOption
	if sampler != nil {
		startOptions = append(startOptions, trace.WithSampler(sampler))
	}

	childCtx, span := trace.StartSpan(ctx, name, startOptions...)
	span.AddAttributes(attributes...)

	return childCtx, span
}

// StartFreshSpan has exact same behavior as StartSpan expect it always starts new fresh trace & span
func StartFreshSpan(ctx context.Context, name string, keyedAttributes ...interface{}) (context.Context, *trace.Span) {
	return StartFreshSpanWithSamplerA(ctx, name, nil, keyedAttributesToTraceAttributes(keyedAttributes)...)
}

// StartFreshSpanWithSamplerA has exact same behavior as StartSpanWithSamplerA expect it always starts new fresh trace & span
func StartFreshSpanA(ctx context.Context, name string, attributes ...trace.Attribute) (context.Context, *trace.Span) {
	return StartFreshSpanWithSamplerA(ctx, name, nil, attributes...)
}

// StartFreshSpanWithSampler has exact same behavior as StartSpanWithSampler expect it always starts new fresh trace & span
func StartFreshSpanWithSampler(ctx context.Context, name string, sampler trace.Sampler, keyedAttributes ...interface{}) (context.Context, *trace.Span) {
	return StartFreshSpanWithSamplerA(ctx, name, sampler, keyedAttributesToTraceAttributes(keyedAttributes)...)
}

var emptySpanContext = trace.SpanContext{}

// StartFreshSpanWithSamplerA has exact same behavior as StartSpanWithSamplerA expect it always starts new fresh trace & span
func StartFreshSpanWithSamplerA(ctx context.Context, name string, sampler trace.Sampler, attributes ...trace.Attribute) (context.Context, *trace.Span) {
	var startOptions []trace.StartOption
	if sampler != nil {
		startOptions = append(startOptions, trace.WithSampler(sampler))
	}

	childCtx, span := trace.StartSpanWithRemoteParent(ctx, name, emptySpanContext, startOptions...)
	span.AddAttributes(attributes...)

	return childCtx, span
}

func keyedAttributesToTraceAttributes(keyedAttributes []interface{}) []trace.Attribute {
	keyedAttributeCount := len(keyedAttributes)
	if keyedAttributeCount <= 0 {
		return nil
	}

	if keyedAttributeCount%2 != 0 {
		zlog.Panic("keyedAttributes parameters should be a multiple of 2", zap.Any("keyed_attributes", keyedAttributes))
	}

	attributes := make([]trace.Attribute, keyedAttributeCount/2)
	for i := 0; i < keyedAttributeCount; i += 2 {
		key := toString(keyedAttributes[i])
		value := keyedAttributes[i+1]
		attributeIndex := (i + 1) / 2

		switch v := value.(type) {
		case int, int8, int16, int32, int64, uintptr, uint, uint8, uint16, uint32, uint64:
			attributes[attributeIndex] = trace.Int64Attribute(key, toInt64(v))
		case bool:
			attributes[attributeIndex] = trace.BoolAttribute(key, v)
		case fmt.Stringer:
			attributes[attributeIndex] = trace.StringAttribute(key, v.String())
		case string:
			attributes[attributeIndex] = trace.StringAttribute(key, v)
		default:
			zlog.Panic("trace attribute must be a integer, a boolean or a string/stringer", zap.String("type", fmt.Sprintf("%T", value)))
		}
	}

	return attributes
}

func toString(input interface{}) string {
	switch v := input.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%T", input)
	}
}

func toInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return int64(v)
	case uint:
		return int64(v)
	case int32:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case uintptr:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	}

	zlog.Panic("Value should be castable to int64")
	return 0
}
