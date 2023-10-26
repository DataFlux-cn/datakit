// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GetReq func(map[string]string, string) (*http.Request, error)

var AuthMaps = map[string]GetReq{
	"bearer_token": BearerToken,
}

func BearerToken(auth map[string]string, url string) (*http.Request, error) {
	token, ok := auth["token"]
	if !ok {
		tokenFile, ok := auth["token_file"]
		if !ok {
			return nil, fmt.Errorf("invalid token")
		}
		tokenBytes, err := os.ReadFile(filepath.Clean(tokenFile))
		if err != nil {
			return nil, fmt.Errorf("invalid token file")
		}
		token = string(tokenBytes)
		token = strings.ReplaceAll(token, "\n", "")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	return req, err
}
