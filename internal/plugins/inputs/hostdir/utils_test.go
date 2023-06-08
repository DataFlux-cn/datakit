// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"os"
	"runtime"
	"testing"
)

func TestStartcollect(t *testing.T) {
	str, _ := os.Getwd()
	t.Log(str)
	a, b, c := Startcollect(str, []string{})
	t.Log(a, b, c)
}

func TestGetFileOwnership(t *testing.T) {
	host := runtime.GOOS
	str, _ := os.Getwd()
	a, _ := GetFileOwnership(str, host)
	t.Log(a)
}

func TestGetuid(t *testing.T) {
	str, _ := os.Getwd()
	host := runtime.GOOS
	u, _ := Getuid(str, host)
	t.Log(u)
}

func TestGetFileSystemType(t *testing.T) {
	str, _ := os.Getwd()
	systemtype, err := GetFileSystemType(str)
	if err != nil {
		t.Log(err)
	}
	t.Log(systemtype)
}
