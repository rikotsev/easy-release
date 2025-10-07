package strategy

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/rikotsev/easy-release/internal/changelog"
	"github.com/rikotsev/easy-release/internal/cli"
	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/rikotsev/easy-release/internal/vcs"
	"github.com/rikotsev/easy-release/internal/version"
)

type StrategyResult string

const (
	Done          StrategyResult = "Done"
	Error         StrategyResult = "Error"
	NotApplicable StrategyResult = "NotApplicable"
)

type Strategy interface {
	Execute(ctx context.Context) (StrategyResult, error)
}

type EasyReleaseContext struct {
	Cfg                 *config.Config
	CommitTypeToSection map[string]*config.ChangelogSection
	Git                 cli.CommandLineClient
	VersionManager      *version.Manager
	CommitParser        *commits.CommitParser
	ChangelogBuilder    *changelog.ChangelogBuilder
	Api                 vcs.Api
}

type EasyReleaseArgs struct {
	Token   string
	Org     string
	Project string
	Repo    string
	Branch  string
}

func LoadEasyReleaseArgs() (*EasyReleaseArgs, error) {
	token := flag.String("token", "", "Access token to authenticate to the API")
	org := flag.String("org", "", "The Azure DevOps Organization Identifier")
	project := flag.String("project", "", "The Azure DevOps Project Identifier")
	repo := flag.String("repo", "", "The Azure DevOps Repository Name")
	branch := flag.String("branch", "", "The branch used for versioning")

	flag.Parse()

	if *token == "" || *org == "" || *project == "" || *repo == "" || *branch == "" {
		flag.PrintDefaults()
		return nil, errors.New("all arguments are required!")
	}

	return &EasyReleaseArgs{
		Token:   *token,
		Org:     *org,
		Project: *project,
		Repo:    *repo,
		Branch:  *branch,
	}, nil
}

func CreateEasyReleaseContext(args *EasyReleaseArgs) (*EasyReleaseContext, error) {
	result := EasyReleaseContext{}
	var err error

	result.Cfg, err = config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	result.CommitTypeToSection, err = config.PivotSections(result.Cfg)
	if err != nil {
		return nil, fmt.Errorf("sections could not be pivoted: %w", err)
	}

	result.Git = cli.New(result.Cfg)

	result.VersionManager, err = version.New(result.Cfg, result.CommitTypeToSection)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate version manager: %w", err)
	}

	result.CommitParser, err = commits.NewParser(result.Cfg)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate commit parser: %w", err)
	}

	result.ChangelogBuilder, err = changelog.NewBuilder(result.Cfg, result.CommitTypeToSection)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate changelog builder: %w", err)
	}

	result.Api, err = vcs.NewAzureDevops(result.Cfg,
		vcs.ApiOpts{
			Token:   args.Token,
			Org:     args.Org,
			Project: args.Project,
			Repo:    args.Repo,
			Branch:  args.Branch,
		})
	if err != nil {
		return nil, fmt.Errorf("could not instantiate devops api client: %w", err)
	}

	return &result, nil
}
