package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

var ErrDuplicateType = errors.New("duplicated commit type in section")

type Config struct {
	GitCommand           string             `json:"gitCommand,omitempty"`
	GitTagCommand        string             `json:"gitTagCommand,omitempty"`
	StartingVersion      string             `json:"startingVersion,omitempty"`
	ExtractCommitRegex   string             `json:"extractCommitRegex,omitempty"`
	LinkPrefix           string             `json:"linkPrefix,omitempty"`
	ReleaseCommitPrefix  string             `json:"releaseCommitPrefix,omitempty"`
	SnapshotCommitPrefix string             `json:"snapshotCommitPrefix,omitempty"`
	ChangelogPath        string             `json:"changelogPath,omitempty"`
	ReleaseBranchPrefix  string             `json:"releaseBranchPrefix,omitempty"`
	ChangelogSections    []ChangelogSection `json:"changelogSections,omitempty"` //the order here will be applied in the resulting changelog
	Updates              []Update           `json:"updates,omitempty"`
	PrLint               PrLint             `json:"prLint,omitempty"`
}

type ChangelogSection struct {
	Section   string   `json:"section,omitempty"`
	Hidden    bool     `json:"hidden,omitempty"`
	Increment string   `json:"increment,omitempty"` // possible values - MAJOR, MINOR, PATCH
	Includes  []string `json:"includes,omitempty"`
}

type Update struct {
	FilePath string `json:"filePath,omitempty"`
	Kind     string `json:"kind,omitempty"` //Only supports Maven
	PomPath  string `json:"pomPath,omitempty"`
	YamlPath string `json:"yamlPath,omitempty"`
	TomlPath string `json:"tomlPath,omitempty"`
}

type PrLint struct {
	AllowedTypes       []string `json:"allowedTypes,omitempty"`
	TypesRequiringJira []string `json:"typesRequiringJira,omitempty"`
}

const (
	configFileName        = ".easy-release.json"
	IncrementVersionMajor = "MAJOR"
	IncrementVersionMinor = "MINOR"
	IncrementVersionPatch = "PATCH"
	IncrementVersionNone  = "NONE"
	UpdateKindMaven       = "MAVEN"
	UpdateKindYaml        = "YAML"
	UpdateKindPackageJson = "PACKAGE_JSON"
	UpdateKindToml        = "TOML"
)

func LoadConfig() (*Config, error) {
	result := Default()

	if _, err := os.Stat(configFileName); errors.Is(err, os.ErrNotExist) {
		return result, nil
	}

	content, err := os.ReadFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s with %w", configFileName, err)
	}

	if err = json.Unmarshal(content, result); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s with %w", configFileName, err)
	}

	return result, nil
}

func Default() *Config {
	return &Config{
		GitCommand:      "git",
		GitTagCommand:   "tag",
		StartingVersion: "1.0.0",
		// Expression breakdown:
		// .* - leading text before the commit type
		// \b(\w+) - the commit type as a whole word
		// (?:\(([^)]+)\))? - optionally - the scope inside ()
		// : colon after the type or type(scope)
		// \s* - any white space after the :
		// (?:\[(.*?)\]\s*)? -  a jira ticket inside []
		// (.+)$ - the rest of the line as a subject
		ExtractCommitRegex:   "^(?:Merged PR(?: \\d+)?:\\s*)?(\\w+)(?:\\(([^)]+)\\))?(!?)\\s*:\\s*(?:\\[(.*?)\\]\\s*)?(.+)$",
		LinkPrefix:           "http://example.com/",
		ReleaseCommitPrefix:  "chore(release): ",
		SnapshotCommitPrefix: "chore(snapshot): ",
		ChangelogPath:        "CHANGELOG.md",
		ReleaseBranchPrefix:  "easy-release--",
		ChangelogSections: []ChangelogSection{
			{
				Section:   "Breaking Changes",
				Hidden:    false,
				Increment: IncrementVersionMajor,
				Includes:  []string{"feat!", "fix!"},
			},
			{
				Section:   "Features",
				Hidden:    false,
				Increment: IncrementVersionMinor,
				Includes:  []string{"feat"},
			},
			{
				Section:   "Fixes",
				Hidden:    false,
				Increment: IncrementVersionPatch,
				Includes:  []string{"fix"},
			},
		},
		Updates: []Update{
			{
				FilePath: "pom.xml",
				Kind:     UpdateKindMaven,
				PomPath:  "//project/properties/revision",
			},
		},
		PrLint: PrLint{
			AllowedTypes: []string{
				"feat",
				"feat!",
				"fix",
				"docs",
				"style",
				"refactor",
				"perf",
				"test",
				"build",
				"ci",
				"chore",
				"revert",
			},
			TypesRequiringJira: []string{
				"feat",
				"feat!",
				"fix",
			},
		},
	}
}

func PivotSections(cfg *Config) (map[string]*ChangelogSection, error) {
	result := make(map[string]*ChangelogSection)

	for idxSection, section := range cfg.ChangelogSections {
		for _, commitType := range section.Includes {
			if _, ok := result[commitType]; ok {
				slog.Error("a section has duplicated commit type", "section index", idxSection, "type", commitType)
				return nil, ErrDuplicateType
			}

			result[commitType] = &section
		}
	}

	return result, nil
}
