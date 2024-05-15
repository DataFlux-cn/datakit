// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package fileprovider

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreFiles(t *testing.T) {
	testcases := []struct {
		ignore, in, out []string
		fail            bool
	}{
		{
			ignore: []string{"/tmp/abc"},
			in:     []string{"/tmp/123"},
			out:    []string{"/tmp/123"},
			fail:   false,
		},
		{
			ignore: []string{"/tmp/*"},
			in:     []string{"/tmp/123"},
			out:    []string{},
			fail:   false,
		},
		{
			ignore: []string{"C:/Users/admin/Desktop/tmp/*"},
			in:     []string{"C:/Users/admin/Desktop/tmp/123"},
			out:    []string{},
			fail:   false,
		},
	}

	for _, tc := range testcases {
		p := NewProvider()
		p.list = tc.in

		result, err := p.IgnoreFiles(tc.ignore).Result()
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, tc.out, result)
	}
}

func TestSearchFiles(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	filename := file.Name()

	testcases := []struct {
		in, out []string
	}{
		{
			in:  []string{filename[:len(filename)-1] + "*"},
			out: []string{filename},
		},
	}

	for _, tc := range testcases {
		fmt.Println(tc.in)
		p := NewProvider()
		result, err := p.SearchFiles(tc.in).Result()
		assert.NoError(t, err)

		assert.Equal(t, tc.out, result)
	}
}
