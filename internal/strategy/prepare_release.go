package strategy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/devops"
	"github.com/rikotsev/easy-release/internal/update"
)

type PrepareReleaseImpl struct {
	args             *EasyReleaseArgs
	appCtx           *EasyReleaseContext
	startingSha      string
	extractedCommits []commits.Commit
	nextVersion      string
	newChangelog     string
	remoteChanges    []devops.RemoteChange
	baseBranch       string
	releaseBranch    string
	releaseLastSha   string
}

func PrepareRelease(args *EasyReleaseArgs, applicationContext *EasyReleaseContext) Strategy {
	return &PrepareReleaseImpl{
		args:          args,
		appCtx:        applicationContext,
		remoteChanges: []devops.RemoteChange{},
		baseBranch:    args.Branch,
		releaseBranch: fmt.Sprintf("%s%s", applicationContext.Cfg.ReleaseBranchPrefix, args.Branch),
	}
}

func (strat *PrepareReleaseImpl) Execute(ctx context.Context) (StrategyResult, error) {
	if err := strat.walkGitHistory(ctx); err != nil {
		return NotApplicable, err
	}

	if strat.nextVersion == strat.startingSha {
		slog.Info("Nothing worth tracking has happened!")
		return NotApplicable, nil
	}

	if err := strat.updateChangelog(); err != nil {
		return Error, err
	}

	if err := strat.updatePathsWithNewVersion(); err != nil {
		return Error, err
	}

	if err := strat.keepReleaseBranchUpToDate(ctx); err != nil {
		return Error, err
	}

	if err := strat.makeCommitWithReleaseChanges(ctx); err != nil {
		return Error, err
	}

	if err := strat.makeOrUpdateThePR(ctx); err != nil {
		return Error, err
	}

	return Done, nil
}

func (strat *PrepareReleaseImpl) walkGitHistory(ctx context.Context) error {
	tags, err := strat.appCtx.Git.Tags(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	strat.startingSha = strat.appCtx.VersionManager.Current(tags)
	logEntries, err := strat.appCtx.Git.Log(ctx, strat.startingSha)
	if err != nil {
		return fmt.Errorf("failed to get log entries: %w", err)
	}

	strat.extractedCommits = strat.appCtx.CommitParser.Extract(ctx, logEntries)
	strat.nextVersion, err = strat.appCtx.VersionManager.Next(strat.startingSha, strat.extractedCommits)
	if err != nil {
		return fmt.Errorf("failed to determine next version: %w", err)
	}

	return nil
}

func (strat *PrepareReleaseImpl) updateChangelog() error {
	chnglog, err := strat.appCtx.ChangelogBuilder.Generate(strat.nextVersion, strat.extractedCommits, time.Now())
	if err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	currentChangelog, err := os.ReadFile(strat.appCtx.Cfg.ChangelogPath)
	if err != nil {
		return fmt.Errorf("make sure a %s file exists. failed to read changelog: %w", strat.appCtx.Cfg.ChangelogPath, err)
	}

	strat.newChangelog = string(chnglog)
	strat.remoteChanges = append(strat.remoteChanges, devops.RemoteChange{
		Path:    strat.appCtx.Cfg.ChangelogPath,
		Content: string(append(chnglog[:], currentChangelog[:]...)),
	})

	return nil
}

func (strat *PrepareReleaseImpl) updatePathsWithNewVersion() error {
	for idx, updCfg := range strat.appCtx.Cfg.Updates {
		updatedFile, newContent, err := update.Execute(strat.nextVersion, updCfg)
		if err != nil {
			return fmt.Errorf("failed to perform update for %s [%d] with %w", updCfg.FilePath, idx, err)
		}
		strat.remoteChanges = append(strat.remoteChanges, devops.RemoteChange{
			Path:    updatedFile,
			Content: string(newContent),
		})
	}

	return nil
}

func (strat *PrepareReleaseImpl) keepReleaseBranchUpToDate(ctx context.Context) error {
	baseLastSha, err := strat.appCtx.Api.GetLastRef(ctx, strat.baseBranch)
	if err != nil {
		return fmt.Errorf("could not get base branch: %s last sha with: %w", strat.baseBranch, err)
	}
	if baseLastSha == "" {
		return fmt.Errorf("base branch: %s should have commits. something is terribly wrong!", strat.baseBranch)
	}

	releaseLastSha, err := strat.appCtx.Api.GetLastRef(ctx, strat.releaseBranch)
	if err != nil {
		return fmt.Errorf("could not get release branch: %s last sha with: %w", strat.releaseBranch, err)
	}

	if releaseLastSha == "" {
		releaseLastSha = "0000000000000000000000000000000000000000"
	}

	strat.releaseLastSha, err = strat.appCtx.Api.UpdateRef(ctx, strat.releaseBranch, baseLastSha, releaseLastSha)
	if err != nil {
		return fmt.Errorf("failed to update release branch: %s with current sha: %s to new sha: %s with: %w",
			strat.releaseBranch, releaseLastSha, baseLastSha, err)
	}

	return nil
}

func (strat *PrepareReleaseImpl) makeCommitWithReleaseChanges(ctx context.Context) error {
	message := strat.releaseMessage()

	return strat.appCtx.Api.PushCommit(ctx, strat.releaseBranch, strat.releaseLastSha, message, strat.remoteChanges)
}

func (strat *PrepareReleaseImpl) makeOrUpdateThePR(ctx context.Context) error {
	prId, err := strat.appCtx.Api.GetPR(ctx, strat.baseBranch, strat.releaseBranch)
	if err != nil {
		return fmt.Errorf("could not get a pr: %w", err)
	}
	prContent := strat.newChangelog

	if len(prContent) > devops.PullRequestDescriptionLimit {
		prContent = prContent[:devops.PullRequestDescriptionLimit]
	}

	if prId == -1 {
		_, err := strat.appCtx.Api.CreatePR(ctx, strat.baseBranch, strat.releaseBranch, strat.releaseMessage(), prContent)
		if err != nil {
			return fmt.Errorf("failed to create pr: %w", err)
		}
	} else {
		_, err := strat.appCtx.Api.UpdatePR(ctx, prId, strat.releaseMessage(), prContent)
		if err != nil {
			return fmt.Errorf("failed to edit pr: %w", err)
		}
	}

	return nil
}

func (strat *PrepareReleaseImpl) releaseMessage() string {
	return fmt.Sprintf("%s%s", strat.appCtx.Cfg.ReleaseCommitPrefix, strat.nextVersion)
}
