package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/rikotsev/easy-release/internal/strategy"
	"github.com/rikotsev/easy-release/internal/vcs"
)

func main() {
	ctx := context.Background()
	prId := flag.Int("id", -1, "the pull request id to be validated")

	args, err := strategy.LoadEasyReleaseArgs()
	if err != nil {
		slog.Error("could not load standard args", "err", err)
	}

	if prId != nil && -1 == *prId {
		slog.Error("you have to provide a PR id to validate")
		os.Exit(1)
	}

	linter, api, err := initServices(args)
	if err != nil {
		slog.Error("could not initialize api", "err", err)
		os.Exit(1)
	}

	title, err := api.GetPRTitle(ctx, *prId)
	if err != nil {
		slog.Error("coult not retrieve PR title", "err", err)
		os.Exit(1)
	}

	status, resp := linter.Lint(title)

	if status == 0 {
		os.Exit(0)
	}

	slog.Error("PR Title is incorrect", "Reason", resp)
	os.Exit(status)
}

func initServices(args *strategy.EasyReleaseArgs) (*commits.CommitLinter, vcs.Api, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not load config with: %w", err)
	}

	linter, err := commits.NewLinter(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize linter: %w", err)
	}

	api, err := vcs.NewAzureDevops(cfg, vcs.ApiOpts{
		Token:   args.Token,
		Org:     args.Org,
		Project: args.Project,
		Repo:    args.Repo,
		Branch:  args.Branch,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize api: %w", err)
	}

	return linter, api, nil
}
