package changelog

import (
	"fmt"
	"testing"
	"time"

	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/stretchr/testify/suite"
)

type ChangelogTestSuite struct {
	suite.Suite
	cfg     *config.Config
	builder *ChangelogBuilder
}

func (suite *ChangelogTestSuite) SetupSuite() {
	cfg := config.Default()
	commitTypeToSection, err := config.PivotSections(cfg)
	suite.NoError(err)
	suite.cfg = cfg
	builder, err := NewBuilder(cfg, commitTypeToSection)
	suite.NoError(err)
	suite.builder = builder
}

func (suite *ChangelogTestSuite) TestGenerate() {
	inputs := []struct {
		version  string
		date     time.Time
		log      []commits.Commit
		expected string
	}{
		{
			version: "1.0.1",
			date:    time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			log: []commits.Commit{
				{
					Type:  "fix",
					Title: "fixed a nasty bug",
					Link:  "JIRA-001",
				},
				{
					Type:  "build",
					Title: "changed a pipeline",
				},
				{
					Type:  "fix",
					Title: "another nasty bug fix",
					Link:  "JIRA-002",
				},
			},
			expected: `
## 1.0.1 (2024-01-20)

### Fixes
* [JIRA-001](http://example.com/JIRA-001) fixed a nasty bug
* [JIRA-002](http://example.com/JIRA-002) another nasty bug fix
`,
		},
		{
			version: "1.1.0",
			date:    time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
			log: []commits.Commit{
				{
					Type:  "fix",
					Title: "fixed a nasty bug",
					Link:  "JIRA-003",
				},
				{
					Type:  "doc",
					Title: "added some info",
				},
				{
					Type:  "feat",
					Title: "added a new endpoint for creating",
					Link:  "JIRA-004",
				},
			},
			expected: `
## 1.1.0 (2024-08-12)

### Features
* [JIRA-004](http://example.com/JIRA-004) added a new endpoint for creating

### Fixes
* [JIRA-003](http://example.com/JIRA-003) fixed a nasty bug
`,
		},
		{
			version: "3.0.0",
			date:    time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			log: []commits.Commit{
				{
					Type:  "feat!",
					Title: "modified the schema for creating",
					Link:  "JIRA-005",
				},
				{
					Type:  "refactor",
					Title: "changed naming to be more explicit",
				},
				{
					Type:  "feat",
					Title: "added a new endpoint for creating",
					Link:  "JIRA-006",
				},
				{
					Type:  "feat",
					Title: "a new endpoint for deleting",
				},
				{
					Type:  "fix",
					Title: "fixed a nasty bug",
				},
				{
					Type:  "fix",
					Title: "incorrect calculation",
					Link:  "JIRA-007",
				},
			},
			expected: `
## 3.0.0 (2024-12-25)

### Breaking Changes
* [JIRA-005](http://example.com/JIRA-005) modified the schema for creating

### Features
* [JIRA-006](http://example.com/JIRA-006) added a new endpoint for creating
* a new endpoint for deleting

### Fixes
* fixed a nasty bug
* [JIRA-007](http://example.com/JIRA-007) incorrect calculation
`,
		},
	}

	for idx, input := range inputs {
		suite.Run(fmt.Sprintf("test generating log [%d]", idx), func() {
			actual, err := suite.builder.Generate(input.version, input.log, input.date)
			suite.NoError(err)
			suite.Equal(input.expected, string(actual))
		})
	}

}

func TestChangelogTestSuite(t *testing.T) {
	suite.Run(t, new(ChangelogTestSuite))
}
