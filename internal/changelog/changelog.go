package changelog

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/rikotsev/easy-release/internal/commits"
	"github.com/rikotsev/easy-release/internal/config"
)

type ChangelogBuilder struct {
	cfg                 *config.Config
	commitTypeToSection map[string]*config.ChangelogSection
	tpl                 *template.Template
}

type SectionItem struct {
	Title       string
	HasLink     bool
	LinkPreview string
	Link        string
}

type TemplateSection struct {
	Title string
	Items []SectionItem
}

type Changelog struct {
	Version  string
	Date     string
	Sections []TemplateSection
}

const tplContent = `
## {{.Version}} ({{.Date}})
{{ range $is, $section := .Sections }}
### {{ $section.Title }}{{ range $ii, $item := $section.Items  }}
* {{ if $item.HasLink }}[{{ $item.LinkPreview }}]({{ $item.Link }}) {{ end }}{{ $item.Title }}{{ end  }}
{{ end  }}`

func NewBuilder(cfg *config.Config, commitTypeToSection map[string]*config.ChangelogSection) (*ChangelogBuilder, error) {

	tpl, err := template.New("changelog").Parse(tplContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &ChangelogBuilder{
		cfg:                 cfg,
		commitTypeToSection: commitTypeToSection,
		tpl:                 tpl,
	}, nil
}

func (builder *ChangelogBuilder) Generate(nextVersion string, extractedCommits []commits.Commit, date time.Time) ([]byte, error) {
	var output bytes.Buffer
	templateData := Changelog{}
	templateData.Version = nextVersion
	templateData.Date = date.Format(time.DateOnly)
	templateData.Sections = []TemplateSection{}

	templateSections := []*TemplateSection{}
	sectionNameToTemplateSection := map[string]*TemplateSection{}

	for _, section := range builder.cfg.ChangelogSections {
		templateSection := TemplateSection{
			Title: section.Section,
			Items: []SectionItem{},
		}
		templateSections = append(templateSections, &templateSection)
		sectionNameToTemplateSection[section.Section] = &templateSection
	}

	for _, ref := range extractedCommits {
		changelogSection, ok := builder.commitTypeToSection[ref.Type]
		if !ok {
			//This commit type is not tracked in changelog, we skip it
			continue
		}

		item := SectionItem{
			Title: ref.Title,
		}

		if ref.Link != "" {
			item.HasLink = true
			item.LinkPreview = ref.Link
			item.Link = fmt.Sprintf("%s%s", builder.cfg.LinkPrefix, ref.Link)
		}

		templateSection, ok := sectionNameToTemplateSection[changelogSection.Section]
		if !ok {
			return nil, fmt.Errorf("this should not be happening! a template section exists that does not exist in the config!")
		}

		templateSection.Items = append(templateSection.Items, item)
	}

	for _, tplSec := range templateSections {
		if len(tplSec.Items) == 0 {
			//This section has no items, we skip it
			continue
		}

		templateData.Sections = append(templateData.Sections, *tplSec)
	}

	if err := builder.tpl.Execute(&output, templateData); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
