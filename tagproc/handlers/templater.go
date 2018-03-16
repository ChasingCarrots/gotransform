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
}

type templateEntry struct {
	Name string
	Tags map[string]string
}

func makeTemplateEntry(name, tagLiteral string) (templateEntry, error) {
	tags, err := tagparser.Parse(tagLiteral)
	if err != nil {
		return templateEntry{}, err
	}
	return templateEntry{
		Name: name,
		Tags: tagparser.Unique(tags),
	}, nil
}

func (_ *Templater) BeginFile(context tagproc.TagContext) error  { return nil }
func (_ *Templater) FinishFile(context tagproc.TagContext) error { return nil }
func (_ *Templater) Finalize() error                             { return nil }

func (ct *Templater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	name := ct.makeName(obj.Name)
	output := filepath.Join(ct.outputPath, name)
	entry, err := makeTemplateEntry(obj.Name, tagLiteral)
	if err != nil {
		return err
	}
	return gotransform.WriteGoTemplate(output, ct.template, entry)
}
