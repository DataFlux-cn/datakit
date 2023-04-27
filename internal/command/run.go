// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package command wraps command functions.
package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"
	"time"
)

func RunWithTimeout(timeout time.Duration, sudo bool, command string, args ...string) ([]byte, error) {
	if sudo {
		args = append([]string{"-n", command}, args...)
		command = "sudo"
	}
	cmd := exec.Command(command, args...) //nolint:gosec

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	after := time.AfterFunc(timeout, func() {
		err = cmd.Process.Kill()
	})

	cmderr := cmd.Wait()

	if !after.Stop() {
		if err == nil {
			return nil, fmt.Errorf("command %q process overtime", command)
		}

		return nil, err
	}

	output, err := ioutil.ReadAll(&buf)
	if err != nil {
		return nil, err
	}

	return output, cmderr
}

// ExitStatus check cmd errors.
func ExitStatus(err error) (int, error) {
	if exiterr, ok := err.(*exec.ExitError); ok { //nolint:errorlint
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}

	return 0, err
}
