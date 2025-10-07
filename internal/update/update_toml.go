package update

import (
	"fmt"
	"github.com/rikotsev/easy-release/internal/config"
	"strconv"
	"strings"
)

type updateToml struct {
	cfg config.Update
}

func (t *updateToml) Run(currentContent []byte, newVersion string) ([]byte, error) {

	lines := strings.Split(string(currentContent), "\n")
	lineNumber, err := strconv.Atoi(t.cfg.TomlPath)
	if err != nil {
		return nil, err
	}

	lines[lineNumber] = fmt.Sprintf("version = \"%s\"", newVersion)

	return []byte(strings.Join(lines, "\n")), nil
}
