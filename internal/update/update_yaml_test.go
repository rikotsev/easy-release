package update

import (
	"fmt"

	"github.com/rikotsev/easy-release/internal/config"
)

func (suite *UpdateTestSuite) TestYamlUpdate() {

	tests := []struct {
		oldContent string
		newVersion string
		newContent string
		cfg        config.Update
	}{
		{
			oldContent: `
openapi: 3.0.3
info:
  title: Employee Life REST API
  version: 0.0.1
  description: "Rest Api definitions for employee life."
`,
			newVersion: "1.0.0",
			newContent: `
openapi: 3.0.3
info:
  title: Employee Life REST API
  version: 1.0.0
  description: "Rest Api definitions for employee life."
`,
			cfg: config.Update{
				YamlPath: ".info.version",
			},
		},
		{
			oldContent: `
openapi:
  info:
    title:
      description:
        version: 0.0.1
`,
			newVersion: "1.2.3",
			newContent: `
openapi:
  info:
    title:
      description:
        version: 1.2.3
`,
			cfg: config.Update{
				YamlPath: ".openapi.info.title.description.version",
			},
		},
		{
			oldContent: `
openapi: 3.0.3
info:
  title: Employee Life REST API
  version: 0.0.1
  description: |
    Rest Api definitions for employee life.
    Yes.
`,
			newVersion: "1.0.0",
			newContent: `
openapi: 3.0.3
info:
  title: Employee Life REST API
  version: 1.0.0
  description: |
    Rest Api definitions for employee life.
    Yes.
`,
			cfg: config.Update{
				YamlPath: ".info.version",
			},
		},
	}

	for idx, testCase := range tests {
		suite.Run(fmt.Sprintf("testing yaml update: [%d]", idx), func() {
			updater := updateYaml{
				cfg: testCase.cfg,
			}

			actual, err := updater.Run([]byte(testCase.oldContent), testCase.newVersion)
			suite.NoError(err)

			suite.Equal(testCase.newContent, string(actual))
		})
	}

}
