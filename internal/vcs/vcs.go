package vcs

import (
	"context"
	"errors"
)

var ErrCannotCreateBranch = errors.New("cannot create a release branch")
var ErrCannotCreatePullRequest = errors.New("cannot create PR")
var ErrCannotUpdatePullRequest = errors.New("cannot update PR")

type Api interface {
	GetLastRef(ctx context.Context, branch string) (string, error)
	UpdateRef(ctx context.Context, branch string, newSha string, oldSha string) (string, error)
	PushCommit(ctx context.Context, branch string, lastSha string, message string, changes []RemoteChange) error
	GetPR(ctx context.Context, toBranch string, fromBranch string) (int, error)
	CreatePR(ctx context.Context, toBranch string, fromBranch string, title string, description string) (int, error)
	UpdatePR(ctx context.Context, prId int, title string, description string) (int, error)
	GetLastCommitMessage(ctx context.Context, branch string) (string, string, error)
	CreateAnnotatedTag(ctx context.Context, sha string, version string) error
	GetPRTitle(ctx context.Context, prId int) (string, error)
}

const PullRequestDescriptionLimit = 4000

type RemoteChange struct {
	Path    string
	Content string
}

type ApiOpts struct {
	Token   string
	Org     string
	Project string
	Repo    string
	Branch  string
}
