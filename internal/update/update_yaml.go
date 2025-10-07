package update

import (
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/rikotsev/easy-release/internal/config"
)

type updateYaml struct {
	cfg config.Update
}

func (upd *updateYaml) Run(currentContent []byte, newVersion string) ([]byte, error) {
	exprString := fmt.Sprintf("%s = \"%s\"", upd.cfg.YamlPath, newVersion)
	preferences := yqlib.ConfiguredYamlPreferences

	evaluator := yqlib.NewStringEvaluator()

	output, err := evaluator.Evaluate(exprString, string(currentContent), yqlib.NewYamlEncoder(preferences), yqlib.NewYamlDecoder(preferences))
	if err != nil {
		return nil, fmt.Errorf("could not execute expression: %s with: %w", exprString, err)
	}

	return []byte(output), nil
}
