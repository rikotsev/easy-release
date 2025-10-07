package version

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
)

var ErrDuplicateType = errors.New("cannot have the same commit type perform different version increments")

type Manager struct {
	cfg                 *config.Config
	commitTypeToSection map[string]*config.ChangelogSection
}

func New(cfg *config.Config, commitTypeToSection map[string]*config.ChangelogSection) (*Manager, error) {
	return &Manager{
		cfg:                 cfg,
		commitTypeToSection: commitTypeToSection,
	}, nil
}

// Will determine the last strict semantic version from all tags if any.
func (m *Manager) Current(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	semVersions := make([]*semver.Version, 0, len(versions))

	for _, vers := range versions {
		sv, err := semver.StrictNewVersion(vers)

		if err != nil {
			//TODO log something maybe
			continue
		}

		semVersions = append(semVersions, sv)
	}

	if len(semVersions) == 0 {
		slog.Info("could not find any strict semantic versions. using the initial one from the config. This will be a first release for the repository.")
		return ""
	}

	sort.Sort(semver.Collection(semVersions))

	return semVersions[len(semVersions)-1].String()
}

func (m *Manager) Next(initialSha string, parsedCommits []commits.Commit) (string, error) {
	if initialSha == "" {
		return m.cfg.StartingVersion, nil
	}

	sv, err := semver.StrictNewVersion(initialSha)

	if err != nil {
		return "", fmt.Errorf("failed to parse the current version with %w", err)
	}

	var (
		increaseMajor = false
		increaseMinor = false
		increasePatch = false
	)

	for _, commit := range parsedCommits {

		section, ok := m.commitTypeToSection[commit.Type]
		if !ok {
			// this commit type is not tracked in a section
			continue
		}

		if section.Increment == config.IncrementVersionMajor {
			increaseMajor = true
			break
		}

		if section.Increment == config.IncrementVersionMinor {
			increaseMinor = true
		}

		if section.Increment == config.IncrementVersionPatch {
			increasePatch = true
		}

	}

	if increaseMajor {
		return sv.IncMajor().String(), nil
	}

	if increaseMinor {
		return sv.IncMinor().String(), nil
	}

	if increasePatch {
		return sv.IncPatch().String(), nil
	}

	return sv.String(), nil
}
