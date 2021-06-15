package cmds

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

func PromDebugger(configFile string) error {
	var configPath = configFile
	if !path.IsAbs(configFile) {
		currentDir, _ := os.Getwd()
		configPath = filepath.Join(currentDir, configFile)
		_, err := os.Stat(configPath)
		if err != nil {
			configPath = filepath.Join(datakit.ConfdDir, "prom", configFile)
			fmt.Printf("config is not found in current dir, using %s instead\n", configPath)
		}
	}

	input, err := GetPromInput(configPath)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = ShowInput(input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil

}

func GetPromInput(configPath string) (*prom.Input, error) {
	inputList, err := config.LoadInputConfigFile(configPath, func() inputs.Input {
		return prom.NewProm("")
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	if len(inputList) != 1 {
		return nil, fmt.Errorf("should test only one prom config, now get %v", len(inputList))
	}

	input, ok := inputList[0].(*prom.Input)

	if !ok {
		return nil, fmt.Errorf("invalid prom instance")
	}

	return input, nil
}

func ShowInput(input *prom.Input) error {
	// init client
	err := input.InitClient()
	if err != nil {
		return err
	}

	// collect points
	err = input.Collect()
	if err != nil {
		return err
	}

	// get collected points
	points := input.GetCachedPoints()

	fmt.Printf("\n================= Line Protocol Points ==================\n\n")
	// measurements collected
	measurements := make(map[string]string)
	timeSeries := make(map[string]string)
	for _, point := range points {
		lp := point.String()
		fmt.Printf(" %s\n", lp)

		influxPoint, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
		if len(influxPoint) != 1 {
			return fmt.Errorf("parse point error")
		}
		if err != nil {
			return err
		}
		timeSeries[fmt.Sprint(influxPoint[0].HashID())] = "true"
		name := point.Name()
		measurements[name] = "true"
	}
	mKeys := make([]string, len(measurements))
	i := 0
	for name := range measurements {
		mKeys[i] = name
		i++
	}
	fmt.Printf("\n================= Summary ==================\n\n")
	fmt.Printf("Total time series: %v\n", len(timeSeries))
	fmt.Printf("Total line protocol points: %v\n", len(points))
	fmt.Printf("Total measurements: %v (%s)\n\n", len(measurements), strings.Join(mKeys, ", "))

	return nil
}
