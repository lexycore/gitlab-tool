package operation

import (
	"errors"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/lexycore/gitlab-tools/internal/client"
	"github.com/lexycore/gitlab-tools/internal/config"
)

// take a look:
// "github.com/gabrie30/ghorg"
// "github.com/korovkin/limiter"

var (
	v = struct {
		updated string
		desc    string
		created string
		beta    string
		master  string
		merged  string
	}{
		updated: "updated",
		desc:    "desc",
		created: "created_at",
		beta:    "beta",
		master:  "master",
		merged:  "merged",
	}
)

var (
	ErrRepoNotFound   = errors.New("Project/Group not found")
	ErrServerNotFound = errors.New("GitLab server not found")
)

func NewClient(ctx *cli.Context) error {
	c := &config.App{
		Config: &config.Config{
			GitLabURL:       ctx.String("gitlab-url"),
			GitLabToken:     ctx.String("gitlab-token"),
			GitLabGroup:     ctx.String("gitlab-group"),
			ExcludeProjects: ctx.StringSlice("exclude-projects"),
		},
	}
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(c.Config.GitLabURL, "/") {
		c.Config.GitLabURL += "/"
	}
	var err error
	c.Client, err = client.InitClient(c.Config)
	if err != nil {
		return err
	}
	return nil
}
