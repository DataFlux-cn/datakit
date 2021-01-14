package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	flagVersion        = flag.Bool("version", false, `show version info`)
	flagCheckConfigDir = flag.Bool("check-config-dir", false, `check datakit conf.d, list configired and mis-configured collectors`)
	flagInputFilters   = flag.String("input-filter", "", "filter the inputs to enable, separator is :")
	flagDocker         = flag.Bool("docker", false, "run within docker")

	flagListCollectors    = flag.Bool("tree", false, `list vailable collectors`)
	flagDumpConfigSamples = flag.String("dump-samples", "", `dump all config samples`)

	flagCmd      = flag.Bool("cmd", false, "run datakit under command line mode")
	flagPipeline = flag.String("pl", "", "pipeline script to test(name only, do not use file path)")
	flagText     = flag.String("pltxt", "", "input text for the pipeline(json or raw text)")

	ReleaseType = ""
)

var (
	inputFilters = []string{}
	l            = logger.DefaultSLogger("main")
)

func main() {
	flag.Parse()

	applyFlags()

	tryLoadConfig()

	// This may throw `Unix syslog delivery error` within docker, so we just
	// start the entry under docker.
	if *flagDocker {
		run()
	} else {
		datakit.Entry = run
		if err := datakit.StartService(); err != nil {
			l.Errorf("start service failed: %s", err.Error())
			return
		}
	}

	l.Info("datakit exited")
}

func applyFlags() {

	if *flagVersion {
		fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
ReleasedInputs: %s
`, git.Version, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader, ReleaseType)
		os.Exit(0)
	}

	datakit.ReleaseType = ReleaseType

	if *flagCmd {
		runDatakitWithCmd()
		os.Exit(0)
	}

	if *flagDumpConfigSamples != "" {
		dumpAllConfigSamples(*flagDumpConfigSamples)
		os.Exit(0)
	}

	if *flagCheckConfigDir {
		config.CheckConfd()
		os.Exit(0)
	}

	if *flagListCollectors {
		listCollectors()
		os.Exit(0)
	}

	if *flagInputFilters != "" {
		inputFilters = strings.Split(":"+strings.TrimSpace(*flagInputFilters)+":", ":")
	}

	if *flagDocker {
		datakit.Docker = true
	}
}

func listCollectors() {
	collectors := map[string][]string{}

	for k, v := range inputs.Inputs {
		cat := v().Catalog()
		collectors[cat] = append(collectors[cat], k)
	}

	star := " * "
	uncheck := " ? "

	ndk := 0
	nuncheck := 0

	output := []string{}

	for k, vs := range collectors {
		output = append(output, k)
		for _, v := range vs {
			checked, ok := inputs.AllInputs[v]
			if !ok {
				l.Errorf("datakit input %s not exists in check list", v)
			}

			if !checked && datakit.ReleaseType == datakit.ReleaseCheckedInputs {
				continue
			}

			if checked {
				output = append(output, fmt.Sprintf("  |--[d]%s%s", star, v))
			} else {
				nuncheck++
				output = append(output, fmt.Sprintf("  |--[d]%s%s", uncheck, v))
			}
			ndk++
		}
	}

	collectors = map[string][]string{}
	for k, v := range tgi.TelegrafInputs {
		collectors[v.Catalog] = append(collectors[v.Catalog], k)
	}

	ntg := 0
	for k, vs := range collectors {
		output = append(output, k)
		for _, v := range vs {

			checked, ok := inputs.AllInputs[v]
			if !ok {
				l.Errorf("telegraf input %s not exists in check list", v)
			}

			if !checked && datakit.ReleaseType == datakit.ReleaseCheckedInputs {
				continue
			}

			if checked {
				output = append(output, fmt.Sprintf("  |--[t]%s%s", star, v))
			} else {
				nuncheck++
				output = append(output, fmt.Sprintf("  |--[t]%s%s", uncheck, v))
			}

			ntg++
		}
	}

	fmt.Println(strings.Join(output, "\n"))
	fmt.Printf("total %d, datakit: %d, telegraf: %d, uncheck: %d\n", ntg+ndk, ndk, ntg, nuncheck)
}

func dumpAllConfigSamples(fpath string) {

	if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
		panic(err)
	}

	for k, v := range inputs.Inputs {
		sample := v().SampleConfig()
		if err := ioutil.WriteFile(filepath.Join(fpath, k+".conf"), []byte(sample), os.ModePerm); err != nil {
			panic(err)
		}
	}

	for k, v := range tgi.TelegrafInputs {
		sample := v.SampleConfig()
		if err := ioutil.WriteFile(filepath.Join(fpath, k+".conf"), []byte(sample), os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func run() {

	if datakit.Cfg.MainCfg.EnablePProf {
		go func() {
			if err := nhttp.ListenAndServe(":6060", nil); err != nil {
				l.Fatalf("pprof server error: %s", err.Error())
			}
		}()
	}

	inputs.StartTelegraf()

	l.Info("datakit start...")
	if err := runDatakitWithHTTPServer(); err != nil {
		return
	}

	l.Info("datakit start ok. Wait signal or service stop...")

	// NOTE:
	// Actually, the datakit process been managed by system service, no matter on
	// windows/UNIX, datakit should exit via `service-stop' operation, so the signal
	// branch should not reached, but for daily debugging(ctrl-c), we kept the signal
	// exit option.
	signals := make(chan os.Signal, datakit.CommonChanCap)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		if sig == syscall.SIGHUP {
			// TODO: reload configures
		} else {
			l.Infof("get signal %v, wait & exit", sig)
			datakit.Quit()
		}

	case <-datakit.StopCh:
		l.Infof("service stopping")
		datakit.Quit()
	}

	l.Info("datakit exit.")
}

func tryLoadConfig() {
	datakit.Cfg.InputFilters = inputFilters

	for {
		if err := config.LoadCfg(datakit.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	l = logger.SLogger("main")
}

func runDatakitWithHTTPServer() error {

	io.Start()

	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	go func() {
		http.Start(datakit.Cfg.MainCfg.HTTPBind)
	}()

	return nil
}

func runDatakitWithCmd() {
	if *flagPipeline != "" {

		if *flagText == "" {
			l.Fatal("-pltxt required")
		}

		if err := pipeline.Init(); err != nil {
			l.Fatalf("pipeline init failed: %s", err.Error())
		}

		start := time.Now()
		pl, err := pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, *flagPipeline))
		if err != nil {
			l.Fatalf("new pipeline failed: %s", err.Error())
		}

		res, err := pl.Run(*flagText).Result()
		if err != nil {
			l.Fatalf("run pipeline failed: %s", err.Error())
		}

		if len(res) == 0 {
			fmt.Println("No data extracted from pipeline")
			return
		}

		if j, err := json.MarshalIndent(res, "", "    "); err != nil {
			l.Fatal(err)
		} else {
			fmt.Printf("Extracted data(cost: %v):\n", time.Since(start))
			fmt.Printf("%s\n", string(j))
		}
	}
}
