package commits

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/rikotsev/easy-release/internal/config"
)

var CannotParseErr = errors.New("could not parse commit from raw log")

type CommitParser struct {
	cfg          *config.Config
	extractRegex *regexp.Regexp
}

type Commit struct {
	Title string
	Type  string
	Link  string
}

type CommitLinter struct {
	parser *CommitParser
}

func NewParser(cfg *config.Config) (*CommitParser, error) {
	compiledRegex, err := regexp.Compile(cfg.ExtractCommitRegex)
	if err != nil {
		return nil, fmt.Errorf("could not compile regex: %s with %w", cfg.ExtractCommitRegex, err)
	}
	return &CommitParser{
		cfg:          cfg,
		extractRegex: compiledRegex,
	}, nil
}

func NewLinter(cfg *config.Config) (*CommitLinter, error) {
	parser, err := NewParser(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser for linter with: %w", err)
	}

	return &CommitLinter{
		parser: parser,
	}, nil
}

func (parser *CommitParser) extract(rawLog string) (Commit, error) {
	matches := parser.extractRegex.FindStringSubmatch(rawLog)

	if len(matches) > 5 {
		return Commit{
			Type:  fmt.Sprintf("%s%s", matches[1], matches[3]),
			Link:  matches[4],
			Title: matches[5],
		}, nil
	}

	return Commit{}, CannotParseErr
}

func (parser *CommitParser) Extract(ctx context.Context, rawLogEntries []string) []Commit {

	result := []Commit{}

	for _, rawLog := range rawLogEntries {
		commit, err := parser.extract(rawLog)

		if err != nil && errors.Is(err, CannotParseErr) && ctx.Value("verbose") != nil {
			slog.Warn("failed to parse log entry", "entry", rawLog)
			continue
		}

		if err != nil {
			slog.Error("failed to extract from raw log", "entry", rawLog, "err", err)
			continue
		}

		result = append(result, commit)
	}

	return result

}

func (linter *CommitLinter) Lint(input string) (int, string) {
	isBadStart := true

	for _, commitType := range linter.parser.cfg.PrLint.AllowedTypes {
		if strings.HasPrefix(input, commitType) {
			isBadStart = false
			break
		}
	}

	if isBadStart {
		return 1, linter.conventionalCommitMessage()
	}
	/*
		if strings.Contains(input, strings.Trim(linter.parser.cfg.ReleaseCommitPrefix, " ")) ||
			strings.Contains(input, strings.Trim(linter.parser.cfg.SnapshotCommitPrefix, " ")) {
			return 1, linter.easyReleaseReservedType()
		}
	*/
	commit, err := linter.parser.extract(input)
	if err != nil {
		return 1, linter.conventionalCommitMessage()
	}

	if !slices.Contains(linter.parser.cfg.PrLint.AllowedTypes, commit.Type) {
		return 1, linter.conventionalCommitMessage()
	}

	if slices.Contains(linter.parser.cfg.PrLint.TypesRequiringJira, commit.Type) && commit.Link == "" {
		return 1, linter.requiredJiraMessage()
	}

	return 0, ""
}

func (linter *CommitLinter) conventionalCommitMessage() string {
	allowedTypes := linter.parser.cfg.PrLint.AllowedTypes
	allowedTypesMessage := fmt.Sprintf("[%s]", strings.Join(allowedTypes, ", "))

	return fmt.Sprintf("Follow conventional commits! `type(scope): [JIRA-XXX] message` - scope and Jira item are optional. Allowed types are: %s", allowedTypesMessage)
}

func (linter *CommitLinter) requiredJiraMessage() string {
	typesWithJiras := linter.parser.cfg.PrLint.TypesRequiringJira
	typesWithJirasMessage := fmt.Sprintf("[%s]", strings.Join(typesWithJiras, ", "))

	return fmt.Sprintf("You have to specify a Jira in []. e.g. `feat: [JIRA-135] new endpoint`. Types that require a Jira reference: %s", typesWithJirasMessage)
}

func (linter *CommitLinter) easyReleaseReservedType() string {
	reservedTypes := []string{
		strings.Trim(linter.parser.cfg.ReleaseCommitPrefix, " "),
		strings.Trim(linter.parser.cfg.SnapshotCommitPrefix, " "),
	}
	reservedTypesMessage := fmt.Sprintf("[%s]", strings.Join(reservedTypes, ", "))

	return fmt.Sprintf("The types: %s are to be used only by easy-release", reservedTypesMessage)
}
