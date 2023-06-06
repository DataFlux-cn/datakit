// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestGroup(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name: "normal",
			pl: `json(_, status)
			 group_between(status, [200, 400], false, newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: false,
			outkey:   "newkey",
		},
		{
			name: "normal",
			pl: `json(_, status)
			 group_between(status, [200, 400], 10, newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: int64(10),
			outkey:   "newkey",
		},
		{
			name: "normal",
			pl: `json(_, status)
			 group_between(status, [200, 400], "ok", newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name: "normal",
			pl: `json(_, status)
			 group_between(status, [200, 299], "ok")`,
			in:       `{"status": 200,"age":47}`,
			expected: "ok",
			outkey:   "status",
		},
		{
			name: "normal",
			pl: `json(_, status)
			 group_between(status, [200, 299], "ok", newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name: "normal",
			pl: `json(_, status)
			 group_between(status, [300, 400], "ok", newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: nil,
			outkey:   "newkey",
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}

			pt := ptinput.GetPoint()
			ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				ptinput.PutPoint(pt)
				t.Fatal(errR)
			}

			v := pt.Fields[tc.outkey]
			// tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, v)

			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}