package strategy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/rikotsev/easy-release/internal/devops"
	"github.com/rikotsev/easy-release/internal/update"
)

type PerformReleaseImpl struct {
	args            *EasyReleaseArgs
	appCtx          *EasyReleaseContext
	baseBranch      string
	releaseSha      string
	releasedVersion *semver.Version
}

func PerformRelease(args *EasyReleaseArgs, applicationContext *EasyReleaseContext) Strategy {
	return &PerformReleaseImpl{
		args:       args,
		appCtx:     applicationContext,
		baseBranch: args.Branch,
	}
}

func (strat *PerformReleaseImpl) Execute(ctx context.Context) (StrategyResult, error) {
	sha, message, err := strat.appCtx.Api.GetLastCommitMessage(ctx, strat.baseBranch)
	if err != nil {
		return Error, fmt.Errorf("failed to retrieve last ref for: %s with: %w", strat.baseBranch, err)
	}

	if !strings.Contains(message, strat.appCtx.Cfg.ReleaseCommitPrefix) {
		slog.Info("last commit does not appear to be a release commit. abandoning perform release process")
		return NotApplicable, nil
	}

	extractedCommits := strat.appCtx.CommitParser.Extract(ctx, []string{message})
	if len(extractedCommits) == 0 {
		slog.Info("last commit was not properly formatted. abandoning perform release process")
		return Error, nil
	}

	strat.releaseSha = sha
	strat.releasedVersion, err = semver.StrictNewVersion(extractedCommits[0].Title)
	if err != nil {
		return Error, fmt.Errorf("committed version - %s is not strict semver: %w", extractedCommits[0].Title, err)
	}

	if err := strat.appCtx.Api.CreateAnnotatedTag(ctx, sha, strat.releasedVersion.String()); err != nil {
		return Error, fmt.Errorf("failed to create tag: %w", err)
	}

	if err := strat.optionallyMakeSnapshot(ctx); err != nil {
		return Error, fmt.Errorf("failed to make snapshot: %w", err)
	}

	if err := strat.touchVersion(strat.releasedVersion.String()); err != nil {
		return Error, fmt.Errorf("failed to make version file: %w", err)
	}

	return Done, nil
}

func (strat *PerformReleaseImpl) optionallyMakeSnapshot(ctx context.Context) error {
	snapshotVersion := fmt.Sprintf("%s-%s", strat.releasedVersion.IncPatch().String(), "SNAPSHOT")

	changes := []devops.RemoteChange{}
	for idx, upd := range strat.appCtx.Cfg.Updates {
		if upd.Kind == config.UpdateKindMaven {
			filePath, content, err := update.Execute(snapshotVersion, upd)
			if err != nil {
				return fmt.Errorf("could not update file: %s [%d] with: %w", upd.FilePath, idx, err)
			}

			changes = append(changes, devops.RemoteChange{
				Path:    filePath,
				Content: string(content),
			})
		}
	}

	if len(changes) == 0 {
		//There are no maven projects to be bumped
		return nil
	}

	message := fmt.Sprintf("%s%s", strat.appCtx.Cfg.SnapshotCommitPrefix, snapshotVersion)

	return strat.appCtx.Api.PushCommit(ctx, strat.baseBranch, strat.releaseSha, message, changes)
}

func (strat *PerformReleaseImpl) touchVersion(version string) error {
	return os.WriteFile(".easy-release-version.txt", []byte(version), 0644)
}
