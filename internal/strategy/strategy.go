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

type (
	StrategyResult string
	VcsPlatform    string
)

const (
	Done          StrategyResult = "Done"
	Error         StrategyResult = "Error"
	NotApplicable StrategyResult = "NotApplicable"
	AzureDevops   VcsPlatform    = "azuredevops"
	Github        VcsPlatform    = "github"
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
	Vcs     VcsPlatform
	Token   string
	Org     string
	Project string
	Repo    string
	Branch  string
}

func LoadEasyReleaseArgs() (*EasyReleaseArgs, error) {
	vcsPlatform := flag.String("vcs", string(AzureDevops), "")
	token := flag.String("token", "", "Access token to authenticate to the API")
	org := flag.String("org", "", "Azure DevOps Organization Identifier / Empty for Github")
	project := flag.String("project", "", "Azure DevOps Project Identifier / Github Owner")
	repo := flag.String("repo", "", "The Repository Name")
	branch := flag.String("branch", "", "The branch used for versioning")

	flag.Parse()

	if *token == "" || (*org == "" && *vcsPlatform == string(AzureDevops)) || *project == "" || *repo == "" || *branch == "" {
		flag.PrintDefaults()
		return nil, errors.New("all arguments are required")
	}

	return &EasyReleaseArgs{
		Vcs:     VcsPlatform(*vcsPlatform),
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

	if args.Vcs == AzureDevops {
		result.Api, err = vcs.NewAzureDevops(result.Cfg,
			vcs.ApiOpts{
				Token:   args.Token,
				Org:     args.Org,
				Project: args.Project,
				Repo:    args.Repo,
				Branch:  args.Branch,
			})
	} else if args.Vcs == Github {
		result.Api, err = vcs.NewGithub(result.Cfg,
			vcs.ApiOpts{
				Token:   args.Token,
				Project: args.Project,
				Repo:    args.Repo,
				Branch:  args.Branch,
			})
	} else {
		return nil, fmt.Errorf("unrecognized vcs platform")
	}

	if err != nil {
		return nil, fmt.Errorf("could not instantiate devops api client: %w", err)
	}

	return &result, nil
}
