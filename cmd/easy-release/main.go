package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/rikotsev/easy-release/internal/strategy"
)

const configFileName = ".easy-release.json"

func main() {
	args, err := strategy.LoadEasyReleaseArgs()
	if err != nil {
		slog.Error("failed to load args", "err", err)
		os.Exit(1)
	}

	appCtx, err := strategy.CreateEasyReleaseContext(args)
	if err != nil {
		slog.Error("failed to create application context", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	performRelease := strategy.PerformRelease(args, appCtx)
	result, err := performRelease.Execute(ctx)
	if err != nil {
		slog.Error("failed to execute perform release strategy", "err", err)
		os.Exit(1)
	}
	if result == strategy.Done {
		slog.Info("a release was performed. no need to perform a new release on this run. exiting")
		os.Exit(0)
	}

	prepareRelease := strategy.PrepareRelease(args, appCtx)
	_, err = prepareRelease.Execute(ctx)
	if err != nil {
		slog.Error("failed to execute prepare release strategy", "err", err)
		os.Exit(1)
	}

}
