package operation

import (
	"fmt"

	"github.com/xanzy/go-gitlab"

	"github.com/lexycore/gitlab-tools/internal/config"
	"github.com/lexycore/gitlab-tools/internal/util"
)

func GetProjectRepos(git *gitlab.Client, cfg *config.Config) error {
	group, _, err := git.Groups.GetGroup(cfg.GitLabGroup, nil)
	if err != nil {
		return err
	}
	for i, repo := range group.Projects {
		if util.ContainsString(&cfg.ExcludeProjects, repo.Name) {
			continue
		}
		fmt.Println(i, ":", repo.Name)
	}
	return nil
}

func GetProjectReposTags(git *gitlab.Client, cfg *config.Config) error {
	group, _, err := git.Groups.GetGroup(cfg.GitLabGroup, nil)
	if err != nil {
		return err
	}
	opt := &gitlab.ListTagsOptions{}
	for i, repo := range group.Projects {
		if util.ContainsString(&cfg.ExcludeProjects, repo.Name) {
			continue
		}
		fmt.Println(i, ":", repo.Name)
		tags, _, err := git.Tags.ListTags(repo.PathWithNamespace, opt)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			fmt.Println("\t-", tag.Commit.CreatedAt, ":", tag.Name, ":", tag.Commit.ID, ":", tag.Release)
		}
	}
	return nil
}

func GetProjectReposMRs(git *gitlab.Client, cfg *config.Config) error {
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
