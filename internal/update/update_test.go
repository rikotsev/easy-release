package update

import (
	"testing"

	"github.com/rikotsev/easy-release/internal/config"
	"github.com/stretchr/testify/suite"
)

type UpdateTestSuite struct {
	suite.Suite
}

func (suite *UpdateTestSuite) TestInvalidKindError() {
	_, _, err := Execute("1.0.0", config.Update{
		Kind: "Random",
	})

	suite.ErrorIs(ErrNotSupportedUpdateKind, err)
}

func (suite *UpdateTestSuite) TestValidKind() {
	suite.Run("maven updater exists", func() {
		_, err := getUpdater(config.Update{
			Kind: config.UpdateKindMaven,
		})
		suite.NoError(err)
	})

	suite.Run("yaml updater exists", func() {
		_, err := getUpdater(config.Update{
			Kind: config.UpdateKindYaml,
		})
		suite.NoError(err)
	})

	suite.Run("package.json updater exists", func() {
		_, err := getUpdater(config.Update{
			Kind: config.UpdateKindPackageJson,
		})
		suite.Require().NoError(err)
	})

	suite.Run("toml updater exists", func() {
		_, err := getUpdater(config.Update{
			Kind: config.UpdateKindToml,
		})
		suite.Require().NoError(err)
	})
}

func TestUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateTestSuite))
}
