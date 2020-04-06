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
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type zapExporter struct{}

// Compile time assertion that the exporter implements trace.Exporter
var _ trace.Exporter = (*zapExporter)(nil)

func (exporter *zapExporter) ExportSpan(span *trace.SpanData) {
	elapsed := span.EndTime.Sub(span.StartTime)

	zlog.Debug("trace span",
		zap.String("name", span.Name),
		zap.Stringer("trace_id", span.TraceID),
		zap.Stringer("span_id", span.SpanID),
		zap.Stringer("parent_span_id", span.ParentSpanID),
		zap.Duration("elapsed", elapsed),
		zap.Reflect("annotations", span.Annotations),
	)
}

func (exporter *zapExporter) ExportView(data *view.Data) {
	elapsed := data.End.Sub(data.Start)

	zlog.Debug("view metrics data",
		zap.Reflect("view", data.View),
		zap.Reflect("rows", data.Rows),
		zap.Duration("elapsed", elapsed),
	)
}
