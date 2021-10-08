package client

import (
	"fmt"

	"github.com/xanzy/go-gitlab"

	"github.com/lexycore/gitlab-tools/internal/config"
)

func InitClient(cfg *config.Config) (*gitlab.Client, error) {
	git, err := gitlab.NewClient(cfg.GitLabToken, gitlab.WithBaseURL(fmt.Sprintf("%sapi/v4/", cfg.GitLabURL)))
	if err != nil {
		return nil, err
	}
	return git, nil
}
