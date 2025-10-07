package devops

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	devopsgit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/rikotsev/easy-release/internal/util"
)

const (
	targetBranchPrefix = "easy-release--"
)

var ErrNoRefsOnBaseBranch = errors.New("the base branch does not have any history")
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

type PayloadChange struct {
	ChangeType string
	Item       PayloadChangeItem
	NewContent PayloadChangeContent
}

type PayloadChangeItem struct {
	Path string
}

type PayloadChangeContent struct {
	Content     string
	ContentType string
}

type ApiOpts struct {
	Token   string
	Org     string
	Project string
	Repo    string
	Branch  string
}

type apiImpl struct {
	cfg    *config.Config
	opts   ApiOpts
	client devopsgit.Client
}

func New(cfg *config.Config, opts ApiOpts) (Api, error) {
	ctx := context.Background()
	organizationUrl := fmt.Sprintf("https://dev.azure.com/%s", opts.Org)
	conn := azuredevops.NewPatConnection(organizationUrl, opts.Token)

	client, err := devopsgit.NewClient(ctx, conn)
	if err != nil {
		return nil, err
	}

	return &apiImpl{
		cfg:    cfg,
		opts:   opts,
		client: client,
	}, nil
}

func (api *apiImpl) GetLastRef(ctx context.Context, branch string) (string, error) {
	resp, err := api.client.GetRefs(ctx, devopsgit.GetRefsArgs{
		Top:          util.Int(1),
		Filter:       util.String(fmt.Sprintf("heads/%s", branch)),
		RepositoryId: &api.opts.Repo,
		Project:      &api.opts.Project,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get last ref for branch: %s with: %w", branch, err)
	}

	if resp != nil && len(resp.Value) == 1 && resp.Value[0].ObjectId != nil {
		return *resp.Value[0].ObjectId, nil
	}

	return "", nil
}

func (api *apiImpl) UpdateRef(ctx context.Context, branch string, newSha string, oldSha string) (string, error) {
	updateResp, err := api.client.UpdateRefs(ctx, devopsgit.UpdateRefsArgs{
		RefUpdates: &[]devopsgit.GitRefUpdate{
			{
				Name:        util.String(fmt.Sprintf("refs/heads/%s", branch)),
				NewObjectId: &newSha,
				OldObjectId: &oldSha,
			},
		},
		RepositoryId: &api.opts.Repo,
		Project:      &api.opts.Project,
	})

	if err != nil {
		return "", fmt.Errorf("failed to update new refs %w", err)
	}
	if updateResp != nil && len(*updateResp) > 0 && (*updateResp)[0].Success != nil && !(*(*updateResp)[0].Success) {
		return "", ErrCannotCreateBranch
	}

	return newSha, nil
}

func (api *apiImpl) PushCommit(ctx context.Context, branch string, lastSha string, message string, changes []RemoteChange) error {
	payloadChanges := []interface{}{}

	for _, chg := range changes {
		payloadChanges = append(payloadChanges, PayloadChange{
			ChangeType: "edit",
			Item: PayloadChangeItem{
				Path: chg.Path,
			},
			NewContent: PayloadChangeContent{
				Content:     chg.Content,
				ContentType: "rawtext",
			},
		})
	}

	_, err := api.client.CreatePush(ctx, devopsgit.CreatePushArgs{
		RepositoryId: &api.opts.Repo,
		Project:      &api.opts.Project,
		Push: &devopsgit.GitPush{
			RefUpdates: &[]devopsgit.GitRefUpdate{
				{
					Name:        util.String(fmt.Sprintf("refs/heads/%s", branch)),
					OldObjectId: &lastSha,
				},
			},
			Commits: &[]devopsgit.GitCommitRef{
				{
					Comment: util.String(message),
					Changes: &payloadChanges,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (api *apiImpl) GetPR(ctx context.Context, toBranch string, fromBranch string) (int, error) {
	resp, err := api.client.GetPullRequests(ctx, devopsgit.GetPullRequestsArgs{
		Project:      &api.opts.Project,
		RepositoryId: &api.opts.Repo,
		SearchCriteria: &devopsgit.GitPullRequestSearchCriteria{
			SourceRefName: util.String(fmt.Sprintf("refs/heads/%s", fromBranch)),
			TargetRefName: util.String(fmt.Sprintf("refs/heads/%s", toBranch)),
		},
		Top: util.Int(1),
	})
	if err != nil {
		return -1, fmt.Errorf("failed to query PRs with src: %s and tgt: %s with: %w", fromBranch, toBranch, err)
	}

	if resp != nil && len(*resp) == 1 && (*resp)[0].PullRequestId != nil {
		return *(*resp)[0].PullRequestId, nil
	}

	return -1, nil
}

func (api *apiImpl) CreatePR(ctx context.Context, toBranch string, fromBranch string, title string, description string) (int, error) {
	resp, err := api.client.CreatePullRequest(ctx, devopsgit.CreatePullRequestArgs{
		Project:      &api.opts.Project,
		RepositoryId: &api.opts.Repo,
		GitPullRequestToCreate: &devopsgit.GitPullRequest{
			SourceRefName: util.String(fmt.Sprintf("refs/heads/%s", fromBranch)),
			TargetRefName: util.String(fmt.Sprintf("refs/heads/%s", toBranch)),
			Title:         &title,
			Description:   &description,
		},
	})
	if err != nil {
		return -1, fmt.Errorf("failed to create PR with src:%s and tgt: %s with: %w", fromBranch, toBranch, err)
	}

	if resp != nil && resp.PullRequestId != nil {
		return *resp.PullRequestId, nil
	}

	return -1, ErrCannotCreatePullRequest
}

func (api *apiImpl) UpdatePR(ctx context.Context, prId int, title string, description string) (int, error) {
	resp, err := api.client.UpdatePullRequest(ctx, devopsgit.UpdatePullRequestArgs{
		Project:       &api.opts.Project,
		RepositoryId:  &api.opts.Repo,
		PullRequestId: util.Int(prId),
		GitPullRequestToUpdate: &devopsgit.GitPullRequest{
			Title:       &title,
			Description: &description,
		},
	})
	if err != nil {
		return -1, fmt.Errorf("failed to update PR with id: %d with: %w", prId, err)
	}

	if resp != nil && resp.PullRequestId != nil {
		return *resp.PullRequestId, nil
	}

	return -1, ErrCannotUpdatePullRequest
}

func (api *apiImpl) GetLastCommitMessage(ctx context.Context, branch string) (string, string, error) {
	resp, err := api.client.GetCommits(ctx, devopsgit.GetCommitsArgs{
		Project:      &api.opts.Project,
		RepositoryId: &api.opts.Repo,
		SearchCriteria: &devopsgit.GitQueryCommitsCriteria{
			Top: util.Int(1),
			ItemVersion: &devopsgit.GitVersionDescriptor{
				Version:     util.String(branch),
				VersionType: &devopsgit.GitVersionTypeValues.Branch,
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to get last commit for branch: %s with: %w", branch, err)
	}

	sha := ""
	message := ""

	if resp != nil && len(*resp) == 1 && (*resp)[0].Comment != nil {
		message = *(*resp)[0].Comment
	}

	if resp != nil && len(*resp) == 1 && (*resp)[0].CommitId != nil {
		sha = *(*resp)[0].CommitId
	}

	return sha, message, nil
}

func (api *apiImpl) CreateAnnotatedTag(ctx context.Context, sha string, version string) error {
	_, err := api.client.CreateAnnotatedTag(ctx, devopsgit.CreateAnnotatedTagArgs{
		Project:      &api.opts.Project,
		RepositoryId: &api.opts.Repo,
		TagObject: &devopsgit.GitAnnotatedTag{
			Name:    util.String(version),
			Message: util.String(version),
			TaggedObject: &devopsgit.GitObject{
				ObjectId: util.String(sha),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %s on sha: %s with: %w", version, sha, err)
	}

	return nil
}

func (api *apiImpl) GetPRTitle(ctx context.Context, prId int) (string, error) {
	resp, err := api.client.GetPullRequestById(ctx, devopsgit.GetPullRequestByIdArgs{
		PullRequestId: util.Int(prId),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get PR with id: %d with: %w", prId, err)
	}

	if resp.Title != nil {
		return *resp.Title, nil
	}

	return "", nil
}
