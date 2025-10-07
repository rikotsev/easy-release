# Easy Release

A couple of tools that should do - with few lines - the heavy lifting behind 
releasing a version.

`easy-release` - should be part of the build pipeline.
`pr-lint` - should be part of the PR pipeline - preferably a first steps. Enforces 

This is inspired mainly from [release-pelase](https://github.com/googleapis/release-please) and used extensively in Azure Devops (unforunately).

## Quick Start

Pre-requisites:
* your base branch should follow a linear history (use squash and merge)
* your project has to be the only project to be released in the repository (no monorepo)
* you have to have a CHANGELOG.md file in the repository (it will be populated)
* you need to set-up a pipeline for building & releasing your project
* the pipeline should also fetch the tags
* you need to set-up a pipeline for validating your pull requests

The pipeline for building & releasing should have the following steps:
```yaml
variables:
    ORG_NAME: '<your org>'
    EASY_RELEASE_VERSION: '0.0.0-beta.3' # this should be the latest version

steps:
    - checkout: self
      fetchDepth: 100 # this is the number of commits back you want to look. For long lasting project this can be very slow 
      fetchTags: true # (you need to set this to true)

    - task: UniversalPackages@0
      inputs:
        command: 'download'
        downloadDirectory: '$(Build.SourcesDirectory)'
        feedsToUse: 'internal'
        feedListDownload: '<your feed>'
        packageListDownload: 'easy-release'
        versionListDownload: '$(EASY_RELEASE_VERSION)'

    - script: |
        chmod +x easy-release
        ./easy-release  -token $(System.AccessToken)    \
                      -org $(ORG_NAME)                  \
                      -project $(System.TeamProject)    \
                      -repo $(Build.Repository.Name)    \
                      -branch $(Build.SourceBranchName)
      displayName: 'Easy Release'
```

The pipeline for validating a pull request should have the following steps:

```yaml
variables:
    ORG_NAME: '<your org>'
    EASY_RELEASE_VERSION: '0.0.0-beta.4' # this should be the latest version

steps:
    - task: UniversalPackages@0
      inputs:
        command: 'download'
        downloadDirectory: '$(Build.SourcesDirectory)'
        feedsToUse: 'internal'
        feedListDownload: '<your feed>'
        packageListDownload: 'pr-lint'
        versionListDownload: '$(EASY_RELEASE_VERSION)'

    - script: |
        chmod +x pr-lint
        ./pr-lint  -token $(System.AccessToken)                 \
                        -org $(ORG_NAME)                        \
                        -project $(System.TeamProject)          \
                        -repo $(Build.Repository.Name)          \
                        -branch $(Build.SourceBranchName)       \
                        -id $(System.PullRequest.PullRequestId)
      displayName: 'PR Lint'
```
You can set-up the PR linting even before checking out the repository and setting up anything else.


## Default Configuration Values

You can specify a configuration for easy-release by setting up a `.easy-release.json` file in your repository
If you do not specify a configuration - easy-release will use the standard one. 
If you decide to override a property - you can override only that property.

```json
{
    "gitCommand": "git",
    "gitTagCommand": "tag",
    "startingVersion": "1.0.0",
    "extractCommitRegex": ".*\\b(\\w+)(?:\\(([^)]+)\\))?(!?)\\s*:\\s*(?:\\[(.*?)\\]\\s*)?(.+)$",
    "linkPrefix": "http://example.com/",
    "releaseCommitPrefix": "chore(release): ",
    "snapshotCommitPrefix": "chore(snapshot): ",
    "changelogPath": "CHANGELOG.md",
    "releaseBranchPrefix": "easy-release--",
    "changelogSections": [
        {
            "section": "Breaking Changes",
            "hidden": false,
            "increment": "MAJOR",
            "includes": ["feat!", "fix!"]
        },
        {
            "section": "Features",
            "hidden": false,
            "increment": "MINOR",
            "includes": ["feat"]
        },
        {
            "section": "Fixes",
            "hidden": false,
            "increment": "PATCH",
            "includes": ["fix"]
        }
    ],
    "updates": [
        {
            "filePath": "pom.xml",
            "kind": "MAVEN",
            "pomPath": "//project/properties/revision"
        }
    ],
    "prLint": {
        "allowedType": [
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
            "revert"
        ],
        "typesRequiringJira": [
            "feat",
            "feat!",
            "fix"
        ]
}
```

## Additional Reading
 * [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
 * [Semantic Versioning](https://semver.org/)
