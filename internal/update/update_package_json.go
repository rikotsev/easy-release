package update

import (
	"encoding/json"
	"github.com/rikotsev/easy-release/internal/config"
	"github.com/tidwall/sjson"
)

type updatePackageJson struct {
	cfg config.Update
}

func (u *updatePackageJson) Run(currentContent []byte, newVersion string) ([]byte, error) {
	var parsedContent map[string]interface{}
	if err := json.Unmarshal(currentContent, &parsedContent); err != nil {
		return nil, err
	}

	if _, ok := parsedContent["version"]; ok {
		output, err := sjson.Set(string(currentContent), "version", newVersion)
		if err != nil {
			return nil, err
		}

		return []byte(output), nil
	}

	return currentContent, nil
}
