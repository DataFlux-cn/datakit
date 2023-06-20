// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package downloader wrap HTTP download function
package downloader

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	humanize "github.com/dustin/go-humanize"
)

var (
	CurDownloading string
	l              = logger.DefaultSLogger("downloader")
)

type WriteCounter struct {
	Total   uint64
	current uint64
	last    float64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	if n > 0 && wc.current >= wc.Total {
		fmt.Println()
	}
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	if wc.last > float64(wc.Total)*0.01 || wc.current == wc.Total { // update progress-bar each 1%
		fmt.Printf("\r%s", strings.Repeat(" ", 100)) //nolint:gomnd
		fmt.Printf("\rDownloading(% 7s)... %s/%s", CurDownloading, humanize.Bytes(wc.current), humanize.Bytes(wc.Total))
		wc.last = 0.0
	}
}

// Extract unzip files from @r to directory @to.
//nolint:cyclop
func Extract(r io.Reader, to string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		l.Error(err)
		return err
	}

	defer gzr.Close() //nolint:errcheck

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			l.Error(err)
			return err
		}

		if hdr == nil {
			continue
		}

		target := filepath.Join(to, hdr.Name) //nolint:gosec

		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.ModePerm); err != nil {
					l.Error(err)
					return err
				}
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				l.Error(err)
				return err
			}

			// TODO: lock file before extracting, to avoid `text file busy` error
			f, err := os.OpenFile(filepath.Clean(target), os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				l.Error(err)
				return err
			}

			if _, err := io.Copy(f, tr); err != nil { //nolint:gosec
				l.Error(err)
				return err
			}

			if err := f.Close(); err != nil {
				l.Warnf("f.Close(): %v, ignored", err)
			}

		default:
			l.Warnf("unexpected file %s", target)
		}
	}
}

func Download(cli *http.Client, from, to string, progress, downloadOnly bool) error {
	l.Debugf("Downloading %s => %s", from, to)
	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		l.Error(err)
		return err
	}

	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close() //nolint:errcheck
	progbar := &WriteCounter{
		Total: uint64(resp.ContentLength),
	}

	if downloadOnly {
		if to == "" {
			to = filepath.Base(from)
		}
		return doDownload(io.TeeReader(resp.Body, progbar), to)
	}

	if !progress {
		return Extract(resp.Body, to)
	}

	return Extract(io.TeeReader(resp.Body, progbar), to)
}

func doDownload(r io.Reader, to string) error {
	f, err := os.OpenFile(filepath.Clean(to), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, r); err != nil { //nolint:gosec
		return err
	}

	return f.Close()
}
