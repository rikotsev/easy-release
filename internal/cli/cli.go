package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rikotsev/easy-release/internal/config"
)

type CommandLineClient interface {
	Tags(context.Context) ([]string, error)
	Log(context.Context, string) ([]string, error)
}

func New(cfg *config.Config) CommandLineClient {
	return &commandLineClientImpl{
		cfg: cfg,
	}
}

type commandLineClientImpl struct {
	cfg *config.Config
}

func (client *commandLineClientImpl) Tags(ctx context.Context) ([]string, error) {

	stdout, _, err := client.runSync(ctx, client.cfg.GitCommand, client.cfg.GitTagCommand)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(stdout), "\n"), nil
}

func (client *commandLineClientImpl) Log(ctx context.Context, startingSha string) ([]string, error) {
	args := []string{}
	args = append(args, "log")
	if startingSha != "" {
		args = append(args, startingSha+"..HEAD")
	}
	args = append(args, "--pretty=format:%s")

	stdout, _, err := client.runSync(ctx, client.cfg.GitCommand, args...)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(stdout), "\n"), nil
}

func (client *commandLineClientImpl) runSync(ctx context.Context, externalCmd string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, externalCmd, args...)

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to acquire stderr pipe: %w", err)
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to acquire stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to execute `%s %s` with: %w", externalCmd, strings.Join(args, " "), err)
	}

	stderr, err := io.ReadAll(errPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read stderr: %w", err)
	}

	if len(stderr) > 0 {
		return nil, nil, fmt.Errorf("command was not executed successfully. output was: %s", string(stderr))
	}

	stdout, err := io.ReadAll(outPipe)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read stdout: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, nil, fmt.Errorf("command did not run successfully: %w", err)
	}

	return stdout, stderr, nil
}
