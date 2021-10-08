package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"github.com/xanzy/go-gitlab"

	"github.com/lexycore/gitlab-tools/internal/changelog"
	"github.com/lexycore/gitlab-tools/internal/client"
	"github.com/lexycore/gitlab-tools/internal/config"
	"github.com/lexycore/gitlab-tools/internal/operation"
	"github.com/lexycore/gitlab-tools/version"
)

const (
	envPrefix = "GT_"

	gitLabURLDefault         = "https://gitlab.com/"
	gitLabTokenDefault       = "your-token-goes-here"
	gitLabGroupDefault       = ""
	cloneNonEmptyOnlyDefault = false
	configFileNameDefault    = ".gitlab-tool.yml"
)

var (
	excludeProjectsDefault = make([]string, 0)
)

// CLI is the command line interface app object structure
type CLI struct {
	app      *cli.App
	Config   *config.Config
	Git      *gitlab.Client
	BasePath *config.GitLabPath
}

// Run is the entry point to the CLI app
func (c *CLI) Run(args []string) error {
	return c.app.Run(args)
}

// CreateCLI creates a new CLI Application with Name, Usage, Version and Actions
func CreateCLI() *CLI {
	c := &CLI{
		app: cli.NewApp(),
	}
	c.app.Name = version.Description
	c.app.Usage = version.Usage
	c.app.Version = version.Version()
	c.app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "gitlab-url",
			Aliases: []string{"u"},
			Value:   gitLabURLDefault,
			EnvVars: []string{envPrefix + "GITLAB_URL"},
			Usage:   "Your GitLab server URL",
		},
		&cli.StringFlag{
			Name:    "gitlab-token",
			Aliases: []string{"t"},
			Value:   gitLabTokenDefault,
			EnvVars: []string{envPrefix + "GITLAB_TOKEN"},
			Usage:   "Your GitLab access token",
		},
		&cli.StringFlag{
			Name:    "gitlab-group",
			Aliases: []string{"g"},
			Value:   gitLabGroupDefault,
			EnvVars: []string{envPrefix + "GITLAB_GROUP"},
			Usage:   "GitLab project group",
		},
		&cli.StringSliceFlag{
			Name:    "exclude-projects",
			Aliases: []string{"e"},
			Value:   cli.NewStringSlice(excludeProjectsDefault...),
			EnvVars: []string{envPrefix + "EXCLUDE_PROJECTS"},
			Usage:   "GitLab projects to exclude",
		},
		&cli.StringFlag{
			Name:    "config-file",
			Aliases: []string{"c"},
			Value:   ".gitlab-tool.yml",
			Usage:   "Application config file",
		},
	}
	c.app.Commands = cli.Commands{
		{
			Name:  "get",
			Usage: "get objects from gitlab server",
			// Action: c.get,
			Subcommands: cli.Commands{
				{
					Name:   "projects",
					Usage:  "get projects from gitlab group",
					Action: c.getProjects,
				},
				{
					Name:   "tags",
					Usage:  "get projects latest tags",
					Action: c.getTags,
				},
				{
					Name:    "merge-requests",
					Aliases: []string{"mrs"},
					Usage:   "get projects latest tags",
					Action:  c.getMRs,
				},
			},
		},
		{
			Name:   "clone",
			Usage:  "clone project or group of projects",
			Action: c.clone,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "exclude",
					Aliases: []string{"e"},
					Usage:   "projects to exclude",
				},
				// &cli.BoolFlag{
				// 	Name:    "non-empty",
				// 	Aliases: []string{"n"},
				// 	Value:   cloneNonEmptyOnlyDefault,
				// 	EnvVars: []string{envPrefix + "CLONE_NON_EMPTY"},
				// 	Usage:   "Clone only non-empty projects",
				// },
			},
		},
		{
			Name:    "changelog",
			Aliases: []string{"chl"},
			Usage:   "changelog operations",
			Subcommands: cli.Commands{
				{
					Name:   "add",
					Usage:  "add changelog section",
					Action: c.addChangelog,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "package",
							Aliases: []string{"p"},
							Usage:   "package name",
						},
						&cli.StringFlag{
							Name:    "version",
							Aliases: []string{"v"},
							Usage:   "version number",
						},
						&cli.StringFlag{
							Name:    "release",
							Aliases: []string{"r"},
							Usage:   "release string",
						},
						&cli.StringFlag{
							Name:    "urgency",
							Aliases: []string{"u"},
							Usage:   "urgency string",
						},
						&cli.StringFlag{
							Name:    "changes",
							Aliases: []string{"c"},
							Usage:   "changes multi-string",
						},
						&cli.StringFlag{
							Name:    "maintainer",
							Aliases: []string{"m"},
							Usage:   "maintainer name and email",
						},
						&cli.StringFlag{
							Name:    "date",
							Aliases: []string{"d"},
							Usage:   "update date",
						},
					},
				},
			},
		},
	}
	c.app.Before = altsrc.InitInputSourceWithContext(c.app.Flags, altsrc.NewYamlSourceFromFlagFunc("config-file"))
	c.app.Action = c.main
	return c
}

func (c *CLI) main(ctx *cli.Context) error {
	// Config := &Config{
	// 	GitLabURL:   ctx.String("gitlab-url"),
	// 	GitLabToken: ctx.String("gitlab-token"),
	// 	GitLabGroup: ctx.String("gitlab-group"),
	// }
	fmt.Println("main")
	return nil
}

func determineGitLabPath(path string) (*config.GitLabPath, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	p := config.GitLabPath{
		Server:  fmt.Sprintf("%s://%s/", u.Scheme, u.Host),
		Group:   strings.Join(parts[0:len(parts)-1], "/"),
		Project: parts[len(parts)-1],
	}

	return &p, nil
}

func (c *CLI) initClient(ctx *cli.Context, checkArg bool) (*config.GitLabPath, error) {
	var path *config.GitLabPath
	var err error
	c.Config = &config.Config{
		GitLabURL:       ctx.String("gitlab-url"),
		GitLabToken:     ctx.String("gitlab-token"),
		GitLabGroup:     ctx.String("gitlab-group"),
		ExcludeProjects: ctx.StringSlice("exclude-projects"),
	}
	if checkArg {
		path, err = determineGitLabPath(ctx.Args().First())
		if err == nil {
			c.Config.GitLabURL = path.Server
			c.Config.GitLabGroup = path.Group
		}
	}
	// Make sure the given URL ends with a slash
	if !strings.HasSuffix(c.Config.GitLabURL, "/") {
		c.Config.GitLabURL += "/"
	}
	c.Git, err = client.InitClient(c.Config)
	if err != nil {
		return nil, err
	}
	return path, nil
}

func (c *CLI) getProjects(ctx *cli.Context) error {
	_, err := c.initClient(ctx, false)
	if err != nil {
		return err
	}
	return operation.GetProjectRepos(c.Git, c.Config)
}

func (c *CLI) getTags(ctx *cli.Context) error {
	_, err := c.initClient(ctx, false)
	if err != nil {
		return err
	}
	return operation.GetProjectReposTags(c.Git, c.Config)
}

func (c *CLI) getMRs(ctx *cli.Context) error {
	_, err := c.initClient(ctx, false)
	if err != nil {
		return err
	}
	return operation.GetProjectReposMRs(c.Git, c.Config)
}

func (c *CLI) clone(ctx *cli.Context) error {
	path, err := c.initClient(ctx, true)
	if err != nil {
		return err
	}
	exclude := ctx.String("exclude")
	return operation.Clone(c.Git, c.Config, path, exclude)
}

func (c *CLI) addChangelog(ctx *cli.Context) error {
	// path, err := c.initClient(ctx, true)
	// if err != nil {
	// 	return err
	// }
	path := &config.GitLabPath{}
	newItem := changelog.Item{
		Package:    ctx.String("package"),
		Version:    ctx.String("version"),
		Release:    ctx.String("release"),
		Urgency:    ctx.String("urgency"),
		Changes:    ctx.String("changes"),
		Maintainer: ctx.String("maintainer"),
		Date:       ctx.String("date"),
	}
	return changelog.Add(c.Git, c.Config, path, &newItem)
}
