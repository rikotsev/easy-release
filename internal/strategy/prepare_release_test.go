package strategy

import (
	"context"
	"github.com/rikotsev/easy-release/internal/changelog"
	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/rikotsev/easy-release/internal/devops"
	"github.com/rikotsev/easy-release/internal/version"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
)

type PrepareReleaseTestSuite struct {
	suite.Suite
	ctx           context.Context
	cfg           *config.Config
	api           *mockApi
	git           *mockGitCli
	changelogFile *os.File
	args          *EasyReleaseArgs
	appCtx        *EasyReleaseContext
}

func (s *PrepareReleaseTestSuite) SetupSuite() {
	s.ctx = context.Background()
	mockedApi := mockApi{
		refs:               make([]string, 0),
		updateDescriptions: make([]string, 0),
		createDescriptions: make([]string, 0),
	}
	git := mockGitCli{
		tags: make([][]string, 0),
		log:  make([][]string, 0),
	}
	s.api = &mockedApi
	s.git = &git
	s.args = &EasyReleaseArgs{
		Branch: "master",
	}
	s.cfg = config.Default()
	s.cfg.Updates = make([]config.Update, 0)
	commitParser, err := commits.NewParser(s.cfg)
	s.Require().NoError(err)
	sections, err := config.PivotSections(s.cfg)
	s.Require().NoError(err)
	changelogBuilder, err := changelog.NewBuilder(s.cfg, sections)
	s.Require().NoError(err)
	versionManager, err := version.New(s.cfg, sections)
	s.appCtx = &EasyReleaseContext{
		Cfg:                 s.cfg,
		Api:                 &mockedApi,
		Git:                 &git,
		CommitParser:        commitParser,
		CommitTypeToSection: sections,
		ChangelogBuilder:    changelogBuilder,
		VersionManager:      versionManager,
	}
}

func (s *PrepareReleaseTestSuite) SetupTest() {
	file, err := os.CreateTemp("", "test-file")
	s.Require().NoError(err)
	s.changelogFile = file
	s.appCtx.Cfg.ChangelogPath = file.Name()
}

func (s *PrepareReleaseTestSuite) TearDownTest() {
	err := os.Remove(s.changelogFile.Name())
	s.Require().NoError(err)
}

func (s *PrepareReleaseTestSuite) TestPullRequestDescriptionIsTruncated() {
	logs := s.git.generateRandomLogs(10 * 10 * 10)
	logsSize := len(strings.Join(logs, ""))
	s.git.tags = append(s.git.tags, []string{"1.0.0"})
	s.git.log = append(s.git.log, logs)
	s.api.refs = append(s.api.refs, "master-sha", "release-sha")
	strategy := PrepareRelease(s.args, s.appCtx)

	res, err := strategy.Execute(s.ctx)
	pullRequestDescriptionSize := len(s.api.updateDescriptions[0])

	s.Require().True(devops.PullRequestDescriptionLimit < logsSize)
	s.Require().NoError(err)
	s.Require().Equal(Done, res)
	s.Require().True(pullRequestDescriptionSize <= devops.PullRequestDescriptionLimit,
		"inputSize", logsSize, "outputSize", pullRequestDescriptionSize,
		"description", s.api.updateDescriptions[0])
}

func TestPrepareReleaseTestSuite(t *testing.T) {
	suite.Run(t, new(PrepareReleaseTestSuite))
}
