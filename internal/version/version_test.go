package version

import (
	"testing"

	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/stretchr/testify/suite"
)

type VersionTestSuite struct {
	suite.Suite
	cfg     *config.Config
	manager *Manager
}

func (suite *VersionTestSuite) SetupSuite() {
	suite.cfg = config.Default()
	commitTypeToSection, err := config.PivotSections(suite.cfg)
	suite.NoError(err)
	manager, err := New(suite.cfg, commitTypeToSection)
	suite.NoError(err)
	suite.manager = manager
}

func (suite *VersionTestSuite) TestCurrent() {

	suite.Run("determine the current version from a list", func() {
		vers := suite.manager.Current([]string{
			"1.0.0",
			"0.0.1",
			"1.2.3",
		})
		suite.Equal("1.2.3", vers)
	})

	suite.Run("determine the current version if the list is empty", func() {
		vers := suite.manager.Current([]string{})
		suite.Equal("", vers)
	})

	suite.Run("determine the current version if the list is nil", func() {
		vers := suite.manager.Current(nil)
		suite.Equal("", vers)
	})

	suite.Run("determine the last version if the list does not contain semantic versions", func() {
		vers := suite.manager.Current([]string{
			"v1.0",
			"2.0",
			"3",
			"0.2-SNAPSHOT",
			"a-cool-tag-i-did-for-fun",
		})
		suite.Equal("", vers)
	})
}

func (suite *VersionTestSuite) TestNext() {

	suite.Run("increment patch because of a fix", func() {
		newVersion, err := suite.manager.Next("1.0.0", []commits.Commit{
			{
				Type:  "fix",
				Title: "a nasty bug was fixed",
				Link:  "JIRA-001",
			},
			{
				Type:  "fix",
				Title: "more fixes",
				Link:  "JIRA-002",
			},
		})

		suite.NoError(err)
		suite.Equal("1.0.1", newVersion)
	})

	suite.Run("increase minor because of a feat", func() {
		newVersion, err := suite.manager.Next("1.0.1", []commits.Commit{
			{
				Type:  "fix",
				Title: "a nasty bug was fixed",
				Link:  "JIRA-003",
			},
			{
				Type:  "feat",
				Title: "a new endpoint",
				Link:  "JIRA-004",
			},
		})

		suite.NoError(err)
		suite.Equal("1.1.0", newVersion)
	})

	suite.Run("increase minor once with multiple feat", func() {
		newVersion, err := suite.manager.Next("1.0.1", []commits.Commit{
			{
				Type:  "fix",
				Title: "a nasty bug was fixed",
				Link:  "JIRA-003",
			},
			{
				Type:  "feat",
				Title: "a new endpoint",
				Link:  "JIRA-004",
			},
			{
				Type:  "feat",
				Title: "a second endpoint",
			},
		})

		suite.NoError(err)
		suite.Equal("1.1.0", newVersion)
	})

	suite.Run("increase major because of a breaking change", func() {
		newVersion, err := suite.manager.Next("1.1.0", []commits.Commit{
			{
				Type:  "feat",
				Title: "a cool new endpoint",
			},
			{
				Type:  "feat!",
				Title: "an endpoint the changes everything",
			},
		})

		suite.NoError(err)
		suite.Equal("2.0.0", newVersion)
	})

	suite.Run("increase major once with multiple breaking changes", func() {
		newVersion, err := suite.manager.Next("1.1.0", []commits.Commit{
			{
				Type:  "feat",
				Title: "a cool new endpoint",
			},
			{
				Type:  "feat!",
				Title: "an endpoint the changes everything",
			},
			{
				Type:  "feat!",
				Title: "another big big change",
			},
		})

		suite.NoError(err)
		suite.Equal("2.0.0", newVersion)
	})

	suite.Run("version is not increased if type is not tracked", func() {
		newVersion, err := suite.manager.Next("1.0.0", []commits.Commit{
			{
				Type:  "build",
				Title: "a change to the pipeline",
			},
			{
				Type:  "chore",
				Title: "another release",
			},
			{
				Type:  "doc",
				Title: "improved the readme",
			},
		})

		suite.NoError(err)
		suite.Equal("1.0.0", newVersion)
	})

	suite.Run("if there is no current version - next should be the default one", func() {
		newVersion, err := suite.manager.Next("", []commits.Commit{})

		suite.NoError(err)
		suite.Equal(suite.cfg.StartingVersion, newVersion)
	})

}

func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
