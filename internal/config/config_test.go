package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPivotSectionsErrorsOut(t *testing.T) {
	input := Config{
		ChangelogSections: []ChangelogSection{
			{
				Section:   "Features",
				Includes:  []string{"feat"},
				Increment: IncrementVersionMinor,
			},
			{
				Section:   "Breaking Changes",
				Includes:  []string{"feat", "feat!"},
				Increment: IncrementVersionMajor,
			},
		},
	}

	_, err := PivotSections(&input)

	assert.ErrorIs(t, ErrDuplicateType, err)
}

func TestPivotSections(t *testing.T) {
	input := Config{
		ChangelogSections: []ChangelogSection{
			{
				Section:   "Features",
				Includes:  []string{"feat"},
				Increment: IncrementVersionMinor,
			},
			{
				Section:   "Breaking Changes",
				Includes:  []string{"feat!", "fix!"},
				Increment: IncrementVersionMajor,
			},
		},
	}

	result, err := PivotSections(&input)

	assert.NoError(t, err)

	featuresSection, ok := result["feat"]
	assert.True(t, ok)
	assert.Equal(t, IncrementVersionMinor, featuresSection.Increment)

	breakingChangesSection1, ok := result["feat!"]
	assert.True(t, ok)
	assert.Equal(t, IncrementVersionMajor, breakingChangesSection1.Increment)

	breakingChangesSection2, ok := result["fix!"]
	assert.True(t, ok)
	assert.Equal(t, IncrementVersionMajor, breakingChangesSection2.Increment)
}

func TestPartialConfigLoad(t *testing.T) {
	cfg := Default()

	json.Unmarshal([]byte("{\"linkPrefix\":\"http://test.com/\",\"extractCommitRegex\":\".*\"}"), cfg)

	assert.Equal(t, "http://test.com/", cfg.LinkPrefix)
	assert.Equal(t, ".*", cfg.ExtractCommitRegex)
	assert.Equal(t, 3, len(cfg.ChangelogSections))
}
