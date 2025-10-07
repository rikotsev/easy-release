package update

import (
	"errors"

	"github.com/beevik/etree"
	"github.com/rikotsev/easy-release/internal/config"
)

var ErrCannotFindElementInPom = errors.New("cannot find version element in pom.xml")

type updateMaven struct {
	cfg config.Update
}

func (upd *updateMaven) Run(currentContent []byte, newVersion string) ([]byte, error) {
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes(currentContent); err != nil {
		return nil, err
	}

	version := doc.FindElement(upd.cfg.PomPath)
	if version == nil {
		return nil, ErrCannotFindElementInPom
	}
	doc.WriteSettings.CanonicalText = true

	version.SetText(newVersion)

	return doc.WriteToBytes()
}
