package strategy

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rikotsev/easy-release/internal/devops"
)

var noMoreStubs = errors.New("no more stubs")

type mockApi struct {
	refs               []string
	createDescriptions []string
	updateDescriptions []string
}

func (m *mockApi) GetLastRef(ctx context.Context, branch string) (string, error) {
	if len(m.refs) > 0 {
		result, whatsLeft := m.refs[0], m.refs[1:]
		m.refs = whatsLeft

		return result, nil
	}

	return "", noMoreStubs
}

func (m *mockApi) UpdateRef(ctx context.Context, branch string, newSha string, oldSha string) (string, error) {
	//do nothing
	return "", nil
}

func (m *mockApi) PushCommit(ctx context.Context, branch string, lastSha string, message string, changes []devops.RemoteChange) error {
	//do nothing
	return nil
}

func (m *mockApi) GetPR(ctx context.Context, toBranch string, fromBranch string) (int, error) {
	//do nothing
	return 1, nil
}

func (m *mockApi) CreatePR(ctx context.Context, toBranch string, fromBranch string, title string, description string) (int, error) {
	m.createDescriptions = append(m.createDescriptions, description)

	return 1, nil
}

func (m *mockApi) UpdatePR(ctx context.Context, prId int, title string, description string) (int, error) {
	m.updateDescriptions = append(m.updateDescriptions, description)

	return 1, nil
}

func (m *mockApi) GetLastCommitMessage(ctx context.Context, branch string) (string, string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockApi) CreateAnnotatedTag(ctx context.Context, sha string, version string) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockApi) GetPRTitle(ctx context.Context, prId int) (string, error) {
	//TODO implement me
	panic("implement me")
}

type mockGitCli struct {
	tags [][]string
	log  [][]string
}

func (git *mockGitCli) Tags(ctx context.Context) ([]string, error) {
	if len(git.tags) > 0 {
		result, whatsLeft := git.tags[0], git.tags[1:]
		git.tags = whatsLeft

		return result, nil
	}
	return nil, noMoreStubs
}

func (git *mockGitCli) Log(ctx context.Context, startingSha string) ([]string, error) {
	if len(git.log) > 0 {
		result, whatsLeft := git.log[0], git.log[1:]
		git.log = whatsLeft

		return result, nil
	}

	return nil, noMoreStubs
}

func (git *mockGitCli) generateRandomLogs(numberOfLogs int) []string {
	var result []string

	for _ = range numberOfLogs {
		result = append(result, fmt.Sprintf("feat: %s", uuid.New().String()))
	}

	return result
}
