package vcs

import (
	"context"
	"fmt"

	"github.com/google/go-github/v75/github"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/rikotsev/easy-release/internal/util"
)

type githubApiImpl struct {
	cfg    *config.Config
	opts   ApiOpts
	client *github.Client
}

var _ Api = &githubApiImpl{}

func NewGithub(cfg *config.Config, opts ApiOpts) (Api, error) {
	client := github.NewClient(nil).WithAuthToken(opts.Token)

	return &githubApiImpl{
		cfg:    cfg,
		opts:   opts,
		client: client,
	}, nil
}

func (g *githubApiImpl) GetLastRef(ctx context.Context, branch string) (string, error) {
	ref, _, err := g.client.Git.GetRef(ctx, g.opts.Project, g.opts.Repo, "refs/heads/"+branch)

	if err != nil {
		return "", fmt.Errorf("failed to get last ref for branch: %s with err: %w", branch, err)
	}

	return ref.GetRef(), nil
}

func (g *githubApiImpl) UpdateRef(ctx context.Context, branch string, newSha string, oldSha string) (string, error) {
	ref, _, err := g.client.Git.UpdateRef(ctx, g.opts.Project, g.opts.Repo, branch, github.UpdateRef{
		SHA:   newSha,
		Force: util.Bool(true),
	})

	if err != nil {
		return "", fmt.Errorf("failed to update branch: %s to sha: %s with err: %w", branch, newSha, err)
	}

	return ref.GetRef(), nil
}

func (g *githubApiImpl) PushCommit(ctx context.Context, branch string, lastSha string, message string, changes []RemoteChange) error {
	entries := make([]*github.TreeEntry, 0, len(changes))

	for _, change := range changes {
		entries = append(entries, &github.TreeEntry{
			Path:    &change.Path,
			Mode:    util.String("100644"),
			Type:    util.String("blob"),
			Content: util.String(change.Content),
		})
	}

	tree, _, err := g.client.Git.CreateTree(ctx, g.opts.Project, g.opts.Repo, lastSha, entries)
	if err != nil {
		return fmt.Errorf("failed to create github tree: %w", err)
	}

	commit, _, err := g.client.Git.CreateCommit(ctx, g.opts.Project, g.opts.Repo, github.Commit{
		SHA:  util.String(lastSha),
		Tree: tree,
		Parents: []*github.Commit{{
			SHA: util.Ptr(lastSha),
		}},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create github commit: %w", err)
	}

	_, _, err = g.client.Git.UpdateRef(ctx, g.opts.Project, g.opts.Repo, branch, github.UpdateRef{
		SHA:   commit.GetSHA(),
		Force: util.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to push github commit: %w", err)
	}

	return nil
}

func (g *githubApiImpl) GetPR(ctx context.Context, toBranch string, fromBranch string) (int, error) {
	list, _, err := g.client.PullRequests.List(ctx, g.opts.Project, g.opts.Repo, &github.PullRequestListOptions{
		Head: fromBranch,
		Base: toBranch,
	})
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve pull requests from branch: %s to branch: %s with err: %w", fromBranch, toBranch, err)
	}

	if len(list) == 0 {
		return -1, nil
	}

	return int(*list[0].ID), nil
}

func (g *githubApiImpl) CreatePR(ctx context.Context, toBranch string, fromBranch string, title string, description string) (int, error) {
	pullRequest, _, err := g.client.PullRequests.Create(ctx, g.opts.Project, g.opts.Repo, &github.NewPullRequest{
		Title: &title,
		Head:  &fromBranch,
		Base:  &toBranch,
		Body:  &description,
	})

	if err != nil {
		return -1, fmt.Errorf("github: %w: %v", ErrCannotCreatePullRequest, err)
	}

	return int(*pullRequest.ID), nil
}

func (g *githubApiImpl) UpdatePR(ctx context.Context, prId int, title string, description string) (int, error) {
	pullRequest, _, err := g.client.PullRequests.Edit(ctx, g.opts.Project, g.opts.Repo, prId, &github.PullRequest{
		Title: &title,
		Body:  &description,
	})

	if err != nil {
		return -1, fmt.Errorf("github: %w: %v", ErrCannotUpdatePullRequest, err)
	}

	return int(*pullRequest.ID), nil
}

func (g *githubApiImpl) GetLastCommitMessage(ctx context.Context, branch string) (string, string, error) {
	commit, _, err := g.client.Repositories.GetCommit(ctx, g.opts.Project, g.opts.Repo, "refs/heads/"+branch, nil)

	if err != nil {
		return "", "", fmt.Errorf("failed to get last commit msg: %w", err)
	}

	return commit.GetCommit().GetSHA(), commit.GetCommit().GetMessage(), nil
}

func (g *githubApiImpl) CreateAnnotatedTag(ctx context.Context, sha string, version string) error {
	//TODO implement me
	panic("implement me")
}

func (g *githubApiImpl) GetPRTitle(ctx context.Context, prId int) (string, error) {
	//TODO implement me
	panic("implement me")
}
