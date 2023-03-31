// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

//nolint:lll
const (
	winUpgradeCmd      = `$env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; Remove-item .install.ps1 -erroraction silentlycontinue; start-bitstransfer -source %s -destination .install.ps1; powershell .install.ps1;`
	winUpgradeCmdProxy = `$env:HTTPS_PROXY="%s"; $env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; Remove-item .install.ps1 -erroraction silentlycontinue; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source %s -destination .install.ps1; powershell .install.ps1;`

	unixUpgradeCmd      = `DK_UPGRADE=1 bash -c "$(curl -L %s)"`
	unixUpgradeCmdProxy = `HTTPS_PROXY="%s" DK_UPGRADE=1 bash -c "$(curl -x "%s" -L %s)"`
)

func runVersionFlags() error {
	showVersion(ReleaseVersion, InputsReleaseType)

	if !*flagVersionDisableUpgradeInfo {
		vis, err := checkNewVersion(ReleaseVersion, *flagVersionUpgradeTestingVersion)
		if err != nil {
			return err
		}

		for _, vi := range vis {
			cp.Infof("\n\n%s version available: %s, commit %s (release at %s)\n\nUpgrade:\n\t",
				vi.versionType, vi.newVersion.VersionString, vi.newVersion.Commit, vi.newVersion.ReleaseDate)
			cp.Infof("%s\n", getUpgradeCommand(vi.newVersion.DownloadURL))
		}
	}

	return nil
}

func checkUpdate(curverStr string, acceptRC bool) int {
	l = logger.SLogger("ota-update")

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

func showVersion(curverStr, releaseType string) {
	fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
ReleasedInputs: %s
`, curverStr, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader, releaseType)
}

type newVersionInfo struct {
	versionType string
	upgrade     bool
	install     bool
	newVersion  *version.VerInfo
}

func (vi *newVersionInfo) String() string {
	if vi.newVersion == nil {
		return ""
	}

	return fmt.Sprintf("%s/%v/%v\n%s",
		vi.versionType,
		vi.upgrade,
		vi.install,
		getUpgradeCommand(vi.newVersion.DownloadURL))
}

func checkNewVersion(curverStr string, showTestingVer bool) (map[string]*newVersionInfo, error) {
	vers, err := getOnlineVersions(showTestingVer)
	if err != nil {
		return nil, fmt.Errorf("getOnlineVersions: %w", err)
	}

	curver, err := getLocalVersion(curverStr)
	if err != nil {
		return nil, fmt.Errorf("getLocalVersion: %w", err)
	}

	vis := map[string]*newVersionInfo{}

	for k, v := range vers {
		// always show testing version if showTestingVer is true
		l.Debugf("compare %s <=> %s", v, curver)

		if version.IsNewVersion(v, curver, true) {
			vis[k] = &newVersionInfo{
				versionType: k,
				upgrade:     true,
				newVersion:  v,
			}
		}
	}
	return vis, nil
}

const (
	versionTypeOnline  = "Online"
	versionTypeTesting = "Testing"

	testingBaseURL = "https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com"
)

func getUpgradeCommand(dlurl string) string {
	var upgradeCmd string

	proxy := config.Cfg.Dataway.HTTPProxy

	switch runtime.GOOS {
	case datakit.OSWindows:
		if proxy != "" {
			upgradeCmd = fmt.Sprintf(winUpgradeCmdProxy, proxy, dlurl)
		} else {
			upgradeCmd = fmt.Sprintf(winUpgradeCmd, dlurl)
		}

	default: // Linux/Mac

		if proxy != "" {
			upgradeCmd = fmt.Sprintf(unixUpgradeCmdProxy, proxy, proxy, dlurl)
		} else {
			upgradeCmd = fmt.Sprintf(unixUpgradeCmd, dlurl)
		}
	}

	return upgradeCmd
}

func getLocalVersion(ver string) (*version.VerInfo, error) {
	v := &version.VerInfo{
		VersionString: strings.TrimPrefix(ver, "v"),
		Commit:        git.Commit,
		ReleaseDate:   git.BuildAt,
	}
	if err := v.Parse(); err != nil {
		return nil, err
	}
	return v, nil
}

func getVersion(addr string) (*version.VerInfo, error) {
	cli := getcli()
	cli.Timeout = time.Second * 5
	urladdr := addr + "/version"

	req, err := nhttp.NewRequest("GET", urladdr, nil)
	if err != nil {
		return nil, fmt.Errorf("http new request err=%w", err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http Do request err=%w", err)
	}

	defer resp.Body.Close() //nolint:errcheck
	infobody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http read body err=%w", err)
	}

	var ver version.VerInfo
	if err = json.Unmarshal(infobody, &ver); err != nil {
		return nil, fmt.Errorf("json unmarshal err=%w", err)
	}

	if err := ver.Parse(); err != nil {
		return nil, err
	}

	ver.DownloadURL = fmt.Sprintf("%s/install.sh", addr)

	if runtime.GOOS == datakit.OSWindows {
		ver.DownloadURL = fmt.Sprintf("%s/install.ps1", addr)
	}
	return &ver, nil
}

func getOnlineVersions(showTestingVer bool) (map[string]*version.VerInfo, error) {
	res := map[string]*version.VerInfo{}

	if v := datakit.GetEnv("DK_INSTALLER_BASE_URL"); v != "" {
		cp.Warnf("setup base URL to %s\n", v)
		OnlineBaseURL = v
	}

	versionInfos := map[string]string{
		versionTypeOnline:  (OnlineBaseURL + "/datakit"),
		versionTypeTesting: (testingBaseURL + "/datakit"),
	}

	for k, v := range versionInfos {
		if k == versionTypeTesting && !showTestingVer {
			continue
		}

		vi, err := getVersion(v)
		if err != nil {
			return nil, fmt.Errorf("get version from %s failed: %w", v, err)
		}
		res[k] = vi
		l.Debugf("get %s version: %s", k, vi)
	}

	return res, nil
}
