// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package engine

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
)

type searchPath struct {
	m map[string]struct{}
	l []string
}

func (p *searchPath) Push(nodeName string) error {
	p.l = append(p.l, nodeName)
	if _, ok := p.m[nodeName]; ok {
		defer func() {
			p.l = p.l[:len(p.l)-1]
		}()
		return fmt.Errorf("circular dependency: %s", p)
	}
	p.m[nodeName] = struct{}{}
	return nil
}

func (p *searchPath) Pop() {
	if len(p.l) == 0 {
		return
	}

	nodeName := p.l[len(p.l)-1]

	p.l = p.l[:len(p.l)-1]
	delete(p.m, nodeName)
}

func NewSearchPath() *searchPath {
	return &searchPath{
		m: map[string]struct{}{},
		l: []string{},
	}
}

func (p *searchPath) String() string {
	if len(p.l) == 0 {
		return ""
	}

	return strings.Join(p.l, " -> ")
}

func EngineCallRefLinkAndCheck(allNg map[string]*runtime.Script) (map[string]*runtime.Script, map[string]error) {
	retMap := map[string]*runtime.Script{}
	retErrMap := map[string]error{}

	for name, ng := range allNg {
		sPath := NewSearchPath()
		if err := dfs(name, ng, allNg, sPath, retMap, retErrMap); err != nil {
			retErrMap[name] = err
		} else {
			retMap[name] = ng
		}
	}

	return retMap, retErrMap
}

func dfs(name string, procc *runtime.Script, allNg map[string]*runtime.Script,
	sPath *searchPath, retMap map[string]*runtime.Script, retErrMap map[string]error,
) error {
	if err := sPath.Push(name); err != nil {
		return err
	}
	if err, ok := retErrMap[name]; ok {
		return fmt.Errorf("%s: %w", sPath.String(), err)
	}

	if _, ok := retMap[name]; ok {
		return nil
	}

	for cName := range procc.CallRef {
		if cNg, ok := allNg[cName]; !ok {
			return fmt.Errorf(sPath.String()+": script %s not found", cName)
		} else {
			procc.CallRef[cName] = cNg
			if err := dfs(cName, cNg, allNg, sPath, retMap, retErrMap); err != nil {
				return err
			}
		}
	}

	retMap[name] = procc
	sPath.Pop()

	return nil
}
