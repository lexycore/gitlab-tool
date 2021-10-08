package operation

import (
	"fmt"

	"github.com/xanzy/go-gitlab"

	"github.com/lexycore/gitlab-tools/internal/config"
	"github.com/lexycore/gitlab-tools/internal/util"
)

func Clone(git *gitlab.Client, cfg *config.Config, path *config.GitLabPath, exclude string) error {
	if path.Server == "" {
		return ErrServerNotFound
	}
	if path.Group == "" && path.Project == "" {
		return ErrRepoNotFound
	}

	if path.Project != "" {
		return cloneRepo(git, cfg, path, exclude)
	}
	return cloneGroup(git, cfg, path, exclude)
}

func cloneGroup(git *gitlab.Client, cfg *config.Config, path *config.GitLabPath, exclude string) error {
	group, _, err := git.Groups.GetGroup(cfg.GitLabGroup, nil)
	if err != nil {
		return err
	}
	tagOpt := &gitlab.ListTagsOptions{
		OrderBy: &v.updated,
		Sort:    &v.desc,
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 1,
		},
	}
	i := 0
	for _, repo := range group.Projects {
		if util.ContainsString(&cfg.ExcludeProjects, repo.Name) {
			continue
		}
		i++
		fmt.Println(i, ":", repo.Name)
		tags, _, err := git.Tags.ListTags(repo.PathWithNamespace, tagOpt)
		if err != nil {
			return err
		} // 64b24f9f8744cc4fe47638cb534cf656a164d59c
		var mrRelease *gitlab.MergeRequest
		for _, tag := range tags {
			fmt.Println("\t-", tag.Commit.CreatedAt, ":", tag.Name, ":", tag.Commit.ID, ":", tag.Release)
			mrOpt := &gitlab.ListProjectMergeRequestsOptions{
				ListOptions: gitlab.ListOptions{
					Page:    1,
					PerPage: 5,
				},
				State:        &v.merged,
				OrderBy:      &v.created,
				Sort:         &v.desc,
				TargetBranch: &v.beta,
				Search:       nil,
			}
			readMRs := true
			for readMRs {
				mrs, response, err := git.MergeRequests.ListProjectMergeRequests(repo.PathWithNamespace, mrOpt)
				if err != nil {
					return err
				}
				mrOpt.ListOptions.Page = response.NextPage
				for _, mr := range mrs {
					if mr.SourceBranch == v.master {
						mrRelease = mr
						fmt.Println("\t\t- beta :", mr.CreatedAt, ":", mr.ID, ":", mr.SHA, ":", mr.MergeCommitSHA, ":", mr.Title)
						readMRs = false
						break
					}
				}
			}
			mrOpt = &gitlab.ListProjectMergeRequestsOptions{
				ListOptions: gitlab.ListOptions{
					Page:    1,
					PerPage: 20,
				},
				State:        &v.merged,
				OrderBy:      &v.created,
				Sort:         &v.desc,
				TargetBranch: &v.master,
				Search:       nil,
			}
			readMRs = true
			for mrOpt.ListOptions.Page > 0 {
				mrsM, response, err := git.MergeRequests.ListProjectMergeRequests(repo.PathWithNamespace, mrOpt)
				if err != nil {
					return err
				}
				mrOpt.ListOptions.Page = response.NextPage
				for _, mr := range mrsM {
					if mrRelease != nil {
						if mrRelease.SHA == mr.MergeCommitSHA {
							fmt.Println("\t\t--------------------------------------")
							mrOpt.ListOptions.Page = 0
							readMRs = false
						}
					}
					fmt.Println("\t\t- master :", mr.CreatedAt, ":", mr.ID, ":", mr.SHA, ":", mr.MergeCommitSHA, ":", mr.Title)
					if !readMRs {
						break
					}
				}
			}
		}
	}
	return nil
}

func cloneRepo(git *gitlab.Client, cfg *config.Config, path *config.GitLabPath, exclude string) error {
	return nil
}
