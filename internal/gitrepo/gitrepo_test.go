// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitrepo

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	T "testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

func TestGetGitClonePathFromGitURL(t *T.T) {
	originInstallDir := datakit.InstallDir
	originGitReposDir := datakit.GitReposDir

	datakit.InstallDir = "/usr/local/datakit"
	datakit.GitReposDir = filepath.Join(datakit.InstallDir, datakit.StrGitRepos)

	cases := []struct {
		name          string
		gitURL        string
		expect        string
		shouldBeError bool
	}{
		{
			name:          "http_test_url",
			gitURL:        "http://username:password@github.com/path/to/repository1.git",
			expect:        "/usr/local/datakit/gitrepos/repository1",
			shouldBeError: false,
		},

		{
			name:          "git_test_url",
			gitURL:        "git@github.com:path/to/repository4.git",
			expect:        "/usr/local/datakit/gitrepos/repository4",
			shouldBeError: false,
		},

		{
			name:          "ssh_test_url",
			gitURL:        "ssh://git@github.com:9000/path/to/repository5.git",
			expect:        "/usr/local/datakit/gitrepos/repository5",
			shouldBeError: false,
		},

		{
			name:          "empty_test_url",
			gitURL:        "",
			expect:        "",
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			repoName, err := getGitClonePathFromGitURL(tc.gitURL)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			assert.Equal(t, tc.expect, repoName)
		})
	}

	datakit.InstallDir = originInstallDir
	datakit.GitReposDir = originGitReposDir
}

func TestIsUserNamePasswordAuth(t *T.T) {
	cases := []struct {
		name          string
		gitURL        string
		expect        bool
		shouldBeError bool
	}{
		{
			name:          "http_test_url",
			gitURL:        "http://username:password@github.com/path/to/repository.git",
			expect:        true,
			shouldBeError: false,
		},

		{
			name:          "git_test_url",
			gitURL:        "git@github.com:path/to/repository.git",
			expect:        false,
			shouldBeError: false,
		},

		{
			name:          "ssh_test_url",
			gitURL:        "ssh://git@github.com:9000/path/to/repository.git",
			expect:        false,
			shouldBeError: false,
		},

		{
			name:          "invalid_test_url",
			gitURL:        "ok",
			expect:        false,
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			isPassword, err := isUserNamePasswordAuth(tc.gitURL)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			assert.Equal(t, tc.expect, isPassword)
		})
	}
}

func TestGetUserNamePasswordFromGitURL(t *T.T) {
	originInstallDir := datakit.InstallDir
	originGitReposDir := datakit.GitReposDir

	datakit.InstallDir = "/usr/local/datakit"
	datakit.GitReposDir = filepath.Join(datakit.InstallDir, datakit.StrGitRepos)

	cases := []struct {
		name          string
		gitURL        string
		expect        map[string]string
		expectAuth    int
		shouldBeError bool
	}{
		{
			name:   "http_test_url",
			gitURL: "http://username:password@github.com/path/to/repository.git",
			expect: map[string]string{
				"username": "password",
			},
			expectAuth:    authUseHTTP,
			shouldBeError: false,
		},

		{
			name:   "git_test_url",
			gitURL: "git@github.com:path/to/repository.git",
			expect: map[string]string{
				"": "",
			},
			expectAuth:    authUseSSH,
			shouldBeError: false,
		},

		{
			name:   "ssh_test_url",
			gitURL: "ssh://git@github.com:9000/path/to/repository.git",
			expect: map[string]string{
				"": "",
			},
			expectAuth:    authUseSSH,
			shouldBeError: false,
		},

		{
			name:   "http_test_url_empty_username",
			gitURL: "http://:password@github.com/path/to/repository.git",
			expect: map[string]string{
				"": "password",
			},
			expectAuth:    authUseHTTP,
			shouldBeError: true,
		},

		{
			name:   "http_test_url_empty_all",
			gitURL: "http://:@github.com/path/to/repository.git",
			expect: map[string]string{
				"": "",
			},
			expectAuth:    authUseHTTP,
			shouldBeError: true,
		},

		{
			name: "special-password",
			gitURL: fmt.Sprintf("http://username:%s@github.com/path/to/repository.git",
				url.QueryEscape(`abc!@#$%^&*()_-+={}\|"':;,.<>/?`)),
			expect: map[string]string{
				"username": `abc!@#$%^&*()_-+={}\|"':;,.<>/?`,
			},
			expectAuth:    authUseHTTP,
			shouldBeError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			as, err := getUserNamePasswordFromGitURL(tc.gitURL)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}

			t.Logf("authOpt: %+#v, gitURL: %s", as, tc.gitURL)

			mVal := map[string]string{
				as.GitUserName: as.GitPassword,
			}
			assert.Equal(t, tc.expect, mVal)
			assert.Equal(t, tc.expectAuth, as.Auth)
		})
	}

	datakit.InstallDir = originInstallDir
	datakit.GitReposDir = originGitReposDir
}

func TestGetAuthMethod(t *T.T) {
	cases := []struct {
		name          string
		as            *authOpt
		c             *config.GitRepository
		expect        transport.AuthMethod
		shouldBeError bool
	}{
		{
			name: "auth_username_password",
			as: &authOpt{
				Auth:        authUseHTTP,
				GitUserName: "user",
				GitPassword: "pass",
			},
			c: &config.GitRepository{},
			expect: &http.BasicAuth{
				Username: "user",
				Password: "pass",
			},
			shouldBeError: false,
		},

		{
			name: "auth_empty_ssh_path",
			as: &authOpt{
				Auth: authUseSSH,
			},
			c:             &config.GitRepository{},
			expect:        nil,
			shouldBeError: true,
		},

		{
			name: "auth_empty_http_path",
			as: &authOpt{
				Auth:        authUseHTTP,
				GitUserName: "",
				GitPassword: "",
			},
			c:             &config.GitRepository{},
			expect:        nil,
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			authM, err := getAuthMethod(tc.as, tc.c)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			assert.Equal(t, tc.expect, authM)
		})
	}
}

func TestGitPull(t *T.T) {
	if !checkDevHost() {
		return
	}

	as := &authOpt{Auth: 2}
	c := &config.GitRepository{
		SSHPrivateKeyPath:     "/Users/user/.ssh/id_rsa",
		SSHPrivateKeyPassword: "",
	}
	authM, err := getAuthMethod(as, c)
	assert.NoError(t, err)

	const clonePath = "/usr/local/datakit/gitrepos/repository"
	const branch = "empty"

	_, err = gitPull(clonePath, branch, authM)
	assert.NoError(t, err)
}
