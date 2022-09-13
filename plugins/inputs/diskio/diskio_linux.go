// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskio

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

var udevPath = "/run/udev/data"

func (i *Input) diskInfo(devName string) (map[string]string, error) {
	var err error
	var stat unix.Stat_t

	path := devName
	err = unix.Stat(path, &stat)
	if err != nil {
		return nil, err
	}

	if i.infoCache == nil {
		i.infoCache = map[string]diskInfoCache{}
	}
	ic, ok := i.infoCache[devName]

	if ok && stat.Mtim.Nano() == ic.modifiedAt {
		return ic.values, nil
	}

	major := unix.Major(stat.Rdev)
	minor := unix.Minor(stat.Rdev)
	udevDataPath := fmt.Sprintf("%s/b%d:%d", udevPath, major, minor)

	di := map[string]string{}

	i.infoCache[devName] = diskInfoCache{
		modifiedAt:   stat.Mtim.Nano(),
		udevDataPath: udevDataPath,
		values:       di,
	}

	f, err := os.Open(filepath.Clean(udevDataPath))
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck,gosec

	scnr := bufio.NewScanner(f)
	var devlinks bytes.Buffer
	for scnr.Scan() {
		l := scnr.Text()
		if len(l) < 4 {
			continue
		}
		if l[:2] == "S:" {
			if devlinks.Len() > 0 {
				devlinks.WriteString(" ")
			}
			devlinks.WriteString("/dev/")
			devlinks.WriteString(l[2:])
			continue
		}
		if l[:2] != "E:" {
			continue
		}
		kv := strings.SplitN(l[2:], "=", 2)
		if len(kv) < 2 {
			continue
		}
		di[kv[0]] = kv[1]
	}

	if devlinks.Len() > 0 {
		di["DEVLINKS"] = devlinks.String()
	}

	return di, nil
}
