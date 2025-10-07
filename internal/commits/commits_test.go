package commits

import (
	"fmt"
	"slices"
	"testing"

	"github.com/rikotsev/easy-release/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CommitsTestSuite struct {
	suite.Suite
	cfg    *config.Config
	parser *CommitParser
	linter *CommitLinter
}

func (suite *CommitsTestSuite) SetupSuite() {
	suite.cfg = config.Default()
	parser, err := NewParser(suite.cfg)
	suite.NoError(err)
	suite.parser = parser
	linter, err := NewLinter(suite.cfg)
	suite.NoError(err)
	suite.linter = linter
}

func (suite *CommitsTestSuite) TestParse() {

	inputs := []struct {
		commit  Commit
		err     error
		builder func(commit Commit) string
	}{
		{
			commit: Commit{
				Type:  "feat",
				Title: "My cool new endpoint that does XYZ",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("%s: %s", commit.Type, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "feat",
				Title: "New endpoint",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("%s:%s", commit.Type, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "feat!",
				Title: "New endpoint",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("%s:%s", commit.Type, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "build",
				Title: "test handling of merged pr message",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("Merged PR 82203: %s: %s", commit.Type, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "doc",
				Title: "test handling of merged pr random numbers",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("Merged PR 123123123123123: %s: %s", commit.Type, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "fix",
				Title: "fixing a nasty bug",
				Link:  "ITEM-0003",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("%s: [%s] %s", commit.Type, commit.Link, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "fix!",
				Title: "fixing a nasty bug",
				Link:  "ITEM-0003",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("%s: [%s] %s", commit.Type, commit.Link, commit.Title)
			},
		},
		{
			commit: Commit{
				Type:  "customtype",
				Title: "This is something very custom that may happen",
				Link:  "JIRAITEM-012335",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("Merged PR 123123:%s:[%s]%s", commit.Type, commit.Link, commit.Title)
			},
		},
		{
			commit: Commit{},
			err:    CannotParseErr,
			builder: func(commit Commit) string {
				return "This is just a random message that was committed"
			},
		},
		{
			commit: Commit{},
			err:    CannotParseErr,
			builder: func(commit Commit) string {
				return "fix!!: [ITEM-0003] fixing a nasty bug"
			},
		},
		{
			commit: Commit{
				Type:  "chore",
				Title: "bump org.projectlombok:lombok from 1.18.30 to 1.18.34",
			},
			err: nil,
			builder: func(commit Commit) string {
				return fmt.Sprintf("%s: %s", commit.Type, commit.Title)
			},
		},
	}

	for idx, input := range inputs {
		suite.Run(fmt.Sprintf("parse commit test [%d] with input %s", idx, input.builder(input.commit)), func() {
			commit, err := suite.parser.extract(input.builder(input.commit))
			if input.err != nil {
				suite.Equal(input.err, err)
			} else {
				suite.NoError(err)
			}

			suite.Equal(input.commit.Type, commit.Type)
			suite.Equal(input.commit.Title, commit.Title)
			suite.Equal(input.commit.Link, commit.Link)
		})
	}

}

func (suite *CommitsTestSuite) TestLint() {

	var (
		ErrFollowConventionalCommits = suite.linter.conventionalCommitMessage()
		ErrNoJiraReference           = suite.linter.requiredJiraMessage()
		//ErrReservedType              = suite.linter.easyReleaseReservedType()
		Subject = "a cool subject"
	)

	tests := []struct {
		input  string
		status int
		output string
	}{
		{Subject, 1, ErrFollowConventionalCommits},
		{"", 1, ErrFollowConventionalCommits},
		{fmt.Sprintf("Merged PR 5431: %s", Subject), 1, ErrFollowConventionalCommits},
		//{suite.cfg.ReleaseCommitPrefix + "1.0.0", 1, ErrReservedType},
		//{suite.cfg.SnapshotCommitPrefix + "1.0.1-SNAPSHOT", 1, ErrReservedType},
	}

	for _, commitType := range suite.cfg.PrLint.AllowedTypes {
		subject := Subject
		if slices.Contains(suite.cfg.PrLint.TypesRequiringJira, commitType) {
			subject = fmt.Sprintf("%s %s", "[JIRA-135]", subject)
			tests = append(tests, struct {
				input  string
				status int
				output string
			}{
				fmt.Sprintf("%s: %s", commitType, Subject),
				1,
				ErrNoJiraReference,
			})
			tests = append(tests, struct {
				input  string
				status int
				output string
			}{
				fmt.Sprintf("%s: JIRA-135 %s", commitType, Subject),
				1,
				ErrNoJiraReference,
			})
		}

		tests = append(tests, struct {
			input  string
			status int
			output string
		}{
			fmt.Sprintf("%s: %s", commitType, subject),
			0,
			"",
		})

		tests = append(tests, struct {
			input  string
			status int
			output string
		}{
			fmt.Sprintf("%s: %s: %s", "something", commitType, subject),
			1,
			ErrFollowConventionalCommits,
		})

	}

	for _, testCase := range tests {
		suite.Run(fmt.Sprintf("linting input: %s", testCase.input), func() {
			println(testCase.input)
			status, resp := suite.linter.Lint(testCase.input)
			suite.Equal(testCase.status, status)
			suite.Equal(testCase.output, resp, fmt.Sprintf("input was: %s", testCase.input))
		})
	}
}

func TestCommitsTestSuite(t *testing.T) {
	suite.Run(t, new(CommitsTestSuite))
}

func TestCannotCreateParser(t *testing.T) {
	cfg := config.Default()
	cfg.ExtractCommitRegex = "(abc"

	_, err := NewParser(cfg)
	assert.Error(t, err)
}
