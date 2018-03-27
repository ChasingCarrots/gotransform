package handlers

import (
	"go/ast"
	"path/filepath"
	"text/template"

	"github.com/chasingcarrots/gotransform/tagparser"

	"github.com/chasingcarrots/gotransform"
	"github.com/chasingcarrots/gotransform/tagproc"
)

// NewTemplater returns a TagHandler that instantiates a Go-template for each tagged
// struct or interface with the declared name and its tags as the template's argument. The parameters
// outputPath and makeName control the path to write the instantiated templates to and how
// the declared name of a struct or interface corresponds to the file name. The template
// parameter specifies the template that should be instantiated.
func NewTemplater(outputPath string, template *template.Template, makeName func(string) string) *Templater {
	return &Templater{
		outputPath: outputPath,
		template:   template,
		makeName:   makeName,
	}
}

type Templater struct {
	outputPath string
	template   *template.Template
	makeName   func(string) string
	entries    []templateEntry
	delay      bool
}

func (t *Templater) Delay() {
	t.delay = true
}

type templateEntry struct {
	Name    string
	Package string
	Tags    map[string]string
}

func makeTemplateEntry(context tagproc.TagContext, obj *ast.Object, tagLiteral string) (templateEntry, error) {
	tags, err := tagparser.Parse(tagLiteral)
	if err != nil {
		return templateEntry{}, err
	}
	return templateEntry{
		Name:    obj.Name,
		Package: context.File.Name.Name,
		Tags:    tagparser.Unique(tags),
	}, nil
}

func (_ *Templater) BeginFile(context tagproc.TagContext) error  { return nil }
func (_ *Templater) FinishFile(context tagproc.TagContext) error { return nil }

func (t *Templater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	entry, err := makeTemplateEntry(context, obj, tagLiteral)
	if err != nil {
		return err
	}
	t.entries = append(t.entries, entry)
	return nil
}

func (t *Templater) Finalize() error {
	if !t.delay {
		return t.WriteTemplates()
	}
	return nil
}

func (t *Templater) WriteTemplates() error {
	for _, entry := range t.entries {
		name := t.makeName(entry.Name)
		output := filepath.Join(t.outputPath, name)
		if err := gotransform.WriteGoTemplate(output, t.template, entry); err != nil {
			return err
		}
	}
	return nil
}
