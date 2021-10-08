package changelog

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/xanzy/go-gitlab"

	"github.com/lexycore/gitlab-tools/internal/config"
)

const defaultChanges = "  some changes were made"

func determineChanges(prevItem *Item, gitClient *gitlab.Client, cfg *config.Config, path *config.GitLabPath) (string, error) {
	// if prevItem == nil || prevItem.Changes == "" {
	// 	return "some changes were made", nil
	// }
	repo, err := git.PlainOpen(".")
	if err != nil {
		return defaultChanges, nil
	}

	// TODO: get merged MRs between previous and new changelog items
	fmt.Println("Remotes")
	remotes, err := repo.Remotes()
	if err != nil {
		return defaultChanges, nil
	}

	for i, remote := range remotes {
		for j, url := range remote.Config().URLs {
			fmt.Println(i, j, url)
		}
	}

	fmt.Println("Tags")
	tags, err := repo.Tags()
	if err != nil {
		return defaultChanges, nil
	}

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		fmt.Println(ref.Name(), ref.Hash(), ref.Type(), ref.String(), ref.Target())
		return nil
	})

	fmt.Println("CommitObjects")
	objs, err := repo.CommitObjects()
	if err != nil {
		return defaultChanges, nil
	}

	err = objs.ForEach(func(commit *object.Commit) error {
		fmt.Println(commit.Committer, commit.Hash, commit.ParentHashes, commit.Message, commit.Author)
		return nil
	})

	return prevItem.Changes, nil
}
