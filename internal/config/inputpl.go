// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	pipelineWarning = `
#------------------------------------   警告   -------------------------------------
# 不要修改本文件，如果要更新，请拷贝至其它文件，最好以某种前缀区分，避免重启后被覆盖
#-----------------------------------------------------------------------------------
`
)

func initPluginPipeline() error {
	scriptMap, err := GetScriptMap()
	if err != nil {
		l.Errorf(err.Error())
		return err
	}

	for name, script := range scriptMap {
		plPath := filepath.Join(datakit.PipelineDir, name)
		if err := ioutil.WriteFile(plPath, []byte(script), datakit.ConfPerm); err != nil {
			l.Errorf("failed to create pipeline script for %s: %s", name, err.Error())
			return err
		}
	}

	if err := pipeline.Init(Cfg.Pipeline); err != nil {
		return err
	}

	return nil
}

func GetScriptMap() (map[string]string, error) {
	scriptMap := map[string]string{}
	for _, c := range inputs.Inputs {
		if v, ok := c().(inputs.PipelineInput); ok {
			scripts := v.PipelineConfig()
			for n, script := range scripts {
				// Ignore empty pipeline script.
				if script == "" {
					continue
				}
				name := n + ".p"
				if _, has := scriptMap[name]; has {
					return nil, fmt.Errorf("duplicated pipeline script name: %s", name)
				}
				scriptMap[name] = pipelineWarning + script
			}
		}
	}
	return scriptMap, nil
}
