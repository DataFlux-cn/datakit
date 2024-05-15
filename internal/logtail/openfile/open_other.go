// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

// Package openfile wrap process file state on systems
package openfile

import (
	"os"
)

func OpenFile(path string) (*os.File, error) {
	flag := os.O_RDONLY
	perm := os.FileMode(0)
	return os.OpenFile(path, flag, perm) //nolint:gosec
}
