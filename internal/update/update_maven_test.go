package update

import (
	"fmt"

	"github.com/rikotsev/easy-release/internal/config"
)

func (suite *UpdateTestSuite) TestMavenUpdate() {
	inputs := []struct {
		oldContent string
		newVersion string
		newContent string
		cfg        config.Update
	}{
		{
			oldContent: `
<project>
	<version>0.0.1-SNAPHOST</version>
</project>`,
			newVersion: "1.0.0",
			newContent: `
<project>
	<version>1.0.0</version>
</project>`,
			cfg: config.Update{
				PomPath: "//project/version",
			},
		},
		{
			oldContent: `
<project>
	<version>${revision}</version>
	<properties>
		<revision>0.0.1-SNAPSHOT</revision>
	</properties>
</project>`,
			newVersion: "1.0.0",
			newContent: `
<project>
	<version>${revision}</version>
	<properties>
		<revision>1.0.0</revision>
	</properties>
</project>`,
			cfg: config.Update{
				PomPath: "//project/properties/revision",
			},
		},
		{
			oldContent: `
<project>
	<version>${revision}</version>
	<properties>
		<revision>0.0.1-SNAPSHOT</revision>
		<my.special.property>A cool "string"</my.special.property>
	</properties>
</project>`,
			newVersion: "1.0.0",
			newContent: `
<project>
	<version>${revision}</version>
	<properties>
		<revision>1.0.0</revision>
		<my.special.property>A cool "string"</my.special.property>
	</properties>
</project>`,
			cfg: config.Update{
				PomPath: "//project/properties/revision",
			},
		},
	}

	for idx, input := range inputs {

		suite.Run(fmt.Sprintf("running maven update: %d", idx), func() {
			updater := updateMaven{
				cfg: input.cfg,
			}

			output, err := updater.Run([]byte(input.oldContent), input.newVersion)
			suite.NoError(err)
			suite.Equal(input.newContent, string(output))
		})

	}

	suite.Run("running maven update - no element error", func() {
		updater := updateMaven{
			cfg: config.Update{
				PomPath: "//project/test",
			},
		}

		_, err := updater.Run([]byte("<project><version>0.1-SNAPSHOT</version></project>"), "1.0.0")
		suite.ErrorIs(err, ErrCannotFindElementInPom)
	})
}
