package cmds

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

const (
	winUpgradeCmd = `Import-Module bitstransfer; ` +
		`start-bitstransfer -source %s -destination .dk-installer.exe; ` +
		`.dk-installer.exe -upgrade; ` +
		`rm .dk-installer.exe`
	unixUpgradeCmd = `sudo -- sh -c ` +
		`"curl %s -o dk-installer ` +
		`&& chmod +x ./dk-installer ` +
		`&& ./dk-installer -upgrade ` +
		`&& rm -rf ./dk-installer"`
)

func CheckUpdate(curverStr string, acceptRC bool) int {

	l = logger.SLogger("ota-update")

	install.Init()

	l.Debugf("get online version...")
	vers, err := getOnlineVersions(false)
	if err != nil {
		l.Errorf("Get online version failed: \n%s\n", err.Error())
		return 0
	}

	ver := vers["Online"]

	curver, err := getLocalVersion(curverStr)
	if err != nil {
		l.Errorf("Get online version failed: \n%s\n", err.Error())
		return -1
	}

	l.Debugf("online version: %v, local version: %v", ver, curver)

	if ver != nil && version.IsNewVersion(ver, curver, acceptRC) {
		l.Infof("New online version available: %s, commit %s (release at %s)",
			ver.VersionString, ver.Commit, ver.ReleaseDate)
		return 42
	} else {
		if acceptRC {
			l.Infof("Up to date(%s)", curver.VersionString)
		} else {
			l.Infof("Up to date(%s), RC version skipped", curver.VersionString)
		}
	}
	return 0
}

func ShowVersion(curverStr, releaseType string, showTestingVer bool) {
	fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
ReleasedInputs: %s
`, curverStr, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader, releaseType)
	vers, err := getOnlineVersions(showTestingVer)
	if err != nil {
		fmt.Printf("Get online version failed: \n%s\n", err.Error())
		os.Exit(-1)
	}
	curver, err := getLocalVersion(curverStr)
	if err != nil {
		fmt.Printf("Get local version failed: \n%s\n", err.Error())
		os.Exit(-1)
	}

	for k, v := range vers {

		// always show testing verison if showTestingVer is true
		l.Debugf("compare %s <=> %s", v, curver)
		if k == "Testing" || version.IsNewVersion(v, curver, true) { // show version info, also show RC verison info
			fmt.Println("---------------------------------------------------")
			fmt.Printf("\n\n%s version available: %s, commit %s (release at %s)\n",
				k, v.VersionString, v.Commit, v.ReleaseDate)
			switch runtime.GOOS {
			case "windows":
				cmdWin := fmt.Sprintf(winUpgradeCmd, v.DownloadURL)
				fmt.Printf("\nUpgrade:\n\t%s\n\n", cmdWin)
			default:
				cmd := fmt.Sprintf(unixUpgradeCmd, v.DownloadURL)
				fmt.Printf("\nUpgrade:\n\t%s\n\n", cmd)
			}
		}
	}
}

func getLocalVersion(ver string) (*version.VerInfo, error) {
	v := &version.VerInfo{
		VersionString: strings.TrimPrefix(datakit.Version, "v"),
		Commit:        git.Commit,
		ReleaseDate:   git.BuildAt}
	if err := v.Parse(); err != nil {
		return nil, err
	}
	return v, nil
}

func getVersion(addr string) (*version.VerInfo, error) {
	resp, err := nhttp.Get("http://" + path.Join(addr, "version"))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	infobody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ver version.VerInfo
	if err = json.Unmarshal(infobody, &ver); err != nil {
		return nil, err
	}

	if err := ver.Parse(); err != nil {
		return nil, err
	}
	ver.DownloadURL = fmt.Sprintf("https://%s/installer-%s-%s",
		addr, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		ver.DownloadURL += ".exe"
	}
	return &ver, nil
}

func getOnlineVersions(showTestingVer bool) (res map[string]*version.VerInfo, err error) {

	nhttp.DefaultTransport.(*nhttp.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	res = map[string]*version.VerInfo{}

	onlineVer, err := getVersion("static.dataflux.cn/datakit")
	if err != nil {
		return nil, err
	}
	res["Online"] = onlineVer
	l.Debugf("online version: %s", onlineVer)

	if showTestingVer {
		testVer, err := getVersion("zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit")
		if err != nil {
			return nil, err
		}
		res["Testing"] = testVer
		l.Debugf("testing version: %s", testVer)
	}

	return
}
