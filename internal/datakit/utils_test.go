// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration(t *T.T) {
	d := Duration{Duration: time.Second}
	assert.Equal(t, "1s", d.UnitString(time.Second))
	assert.Equal(t, "1000000000ns", d.UnitString(time.Nanosecond))
	assert.Equal(t, "1000000mics", d.UnitString(time.Microsecond))
	assert.Equal(t, "1000ms", d.UnitString(time.Millisecond))
	assert.Equal(t, "0m", d.UnitString(time.Minute))
	assert.Equal(t, "0h", d.UnitString(time.Hour))
}
