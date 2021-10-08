package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type App struct {
	BasePath *GitLabPath
	Config   *Config
	Client   *gitlab.Client
}

type GitLabPath struct {
	Server  string
	Group   string
	Project string
}

// Config represents common gitlab-tools settings
type Config struct {
	GitLabURL       string
	GitLabToken     string
	GitLabGroup     string
	ExcludeProjects []string
}

const configFileName = ".gitlab-tool.yml"

var cfgPaths = []string{
	"~/.gitlab",
	"/etc/gitlab",
}

func FindGitlabToolConfig(configFileName string) (string, error) {
	var cfgPathsCalc []string

	homeCfg := fmt.Sprintf("%s/%s", strings.TrimRight(os.ExpandEnv("~"), "/"), configFileName)
	curPath, err := os.Getwd()
	if err == nil {
		curPathItems := strings.Split(strings.TrimRight(curPath, "/"), "/")
		cfgPathsCalc = make([]string, 0, len(curPathItems)+len(cfgPaths))
		for i := len(curPathItems); i >= 0; i-- {
			iPath := strings.Join(append(curPathItems[0:i], configFileName), "/")
			cfgPathsCalc = append(cfgPathsCalc, iPath)
			if iPath == homeCfg {
				break
			}
		}
	}
	for _, cfgPath := range append(cfgPathsCalc, cfgPaths...) {
		if _, err := os.Stat(cfgPath); err != nil {
			return "", err
		}
	}
	return "", nil
}
