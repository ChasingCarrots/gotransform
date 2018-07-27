package handlers

import (
	"go/ast"
	"path/filepath"
	"text/template"

	"github.com/chasingcarrots/gotransform"
	"github.com/chasingcarrots/gotransform/tagproc"
)

// NewTemplater returns a TagHandler that instantiates a Go-template for each tagged
// struct or interface with the declared name and its tags as the template's argument. The parameters
// outputPath and makeName control the path to write the instantiated templates to and how
// the declared name of a struct or interface corresponds to the file name. The template
// parameter specifies the template that should be instantiated.
func NewTemplater(outputPath string, template *template.Template, makeName func(string) string, formatGoCode bool) *Templater {
	return &Templater{
		outputPath:   outputPath,
		template:     template,
		makeName:     makeName,
		formatGoCode: formatGoCode,
	}
}

type Templater struct {
	templateCollection
	outputPath   string
	template     *template.Template
	makeName     func(string) string
	formatGoCode bool
	delay        bool
}

// Delay prevents any generation of files in the Finalize function and instead requires you to call
// WriteTemplates at another time that is more suitable for your uses. This might be required when
// using an inception-based approach.
func (t *Templater) Delay() {
	t.delay = true
}

func (_ *Templater) BeginFile(context tagproc.TagContext) error  { return nil }
func (_ *Templater) FinishFile(context tagproc.TagContext) error { return nil }

func (t *Templater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	return t.addEntry(context, obj, tagLiteral)
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
		if t.formatGoCode {
			if err := gotransform.WriteGoTemplate(output, t.template, entry); err != nil {
				return err
			}
		} else {
			if err := gotransform.WriteTemplate(output, t.template, entry); err != nil {
				return err
			}
		}
	}
	return nil
}
