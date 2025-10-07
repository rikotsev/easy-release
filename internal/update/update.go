package update

import (
	"errors"
	"fmt"
	"os"

	"github.com/rikotsev/easy-release/internal/config"
)

var ErrNotSupportedUpdateKind = errors.New("the update kind is not supported")

type Update interface {
	Run(currentContent []byte, newVersion string) ([]byte, error)
}

func Execute(nextVersion string, updateConfig config.Update) (string, []byte, error) {
	updater, err := getUpdater(updateConfig)
	if err != nil {
		return "", nil, err
	}

	currentContent, err := os.ReadFile(updateConfig.FilePath)
	if err != nil {
		return "", nil, fmt.Errorf("could not find file to update on path %s with error %w", updateConfig.FilePath, err)
	}

	newContent, err := updater.Run(currentContent, nextVersion)
	if err != nil {
		return "", nil, fmt.Errorf("update failed with %w", err)
	}

	return updateConfig.FilePath, newContent, nil
}

func getUpdater(updateConfig config.Update) (Update, error) {
	if updateConfig.Kind == config.UpdateKindMaven {
		return &updateMaven{
			cfg: updateConfig,
		}, nil
	}

	if updateConfig.Kind == config.UpdateKindYaml {
		return &updateYaml{
			cfg: updateConfig,
		}, nil
	}

	if updateConfig.Kind == config.UpdateKindPackageJson {
		return &updatePackageJson{
			cfg: updateConfig,
		}, nil
	}

	if updateConfig.Kind == config.UpdateKindToml {
		return &updateToml{
			cfg: updateConfig,
		}, nil
	}

	return nil, ErrNotSupportedUpdateKind
}
