// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"os"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/gobwas/glob"
)

type Provider struct {
	list    []string
	lastErr error
}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) SearchFiles(patterns []string) *Provider {
	if len(patterns) == 0 {
		return p
	}

	for _, pattern := range patterns {
		info, err := os.Stat(pattern)
		if err == nil && !info.IsDir() { // Is file
			p.list = append(p.list, pattern)
			continue
		}

		paths, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			p.lastErr = err
			continue
		}
		p.list = append(p.list, paths...)
	}

	p.list = unique(p.list)
	return p
}

func (p *Provider) IgnoreFiles(patterns []string) *Provider {
	if len(patterns) == 0 {
		return p
	}

	var ignores []glob.Glob
	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			p.lastErr = err
			continue
		}
		ignores = append(ignores, g)
	}

	pass := []string{}
	for _, path := range p.list {
		matched := false
		func() {
			for _, g := range ignores {
				if g.Match(path) {
					matched = true
					return
				}
			}
		}()
		if !matched {
			pass = append(pass, path)
		}
	}

	p.list = pass
	return p
}

func (p *Provider) Result() (list []string, lastErr error) {
	return p.list, p.lastErr
}

func unique(slice []string) []string {
	var res []string
	keys := make(map[string]interface{})
	for _, str := range slice {
		if _, ok := keys[str]; !ok {
			keys[str] = nil
			res = append(res, str)
		}
	}
	return res
}
