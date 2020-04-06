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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTraceID(t *testing.T) {
	traceID := GetTraceID(NewZeroedTraceIDInContext(context.Background()))
	assert.Equal(t, NewFixedTraceID("00000000000000000000000000000000"), traceID)

	traceID = GetTraceID(NewFixedTraceIDInContext(context.Background(), "000102030405060708090a0b0c0d0e0f"))
	assert.Equal(t, NewFixedTraceID("000102030405060708090a0b0c0d0e0f"), traceID)

	traceIDRandomOne := GetTraceID(context.Background())
	traceIDRandomTwo := GetTraceID(context.Background())

	assert.NotEqual(t, traceIDRandomOne, traceIDRandomTwo)
}
