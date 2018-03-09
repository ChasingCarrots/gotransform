package handlers

import (
	"go/ast"
	"gotransform"
	"gotransform/tagproc"
	"path/filepath"
	"text/template"
)

// NewTemplater returns a TagHandler that instantiates a Go-template for each tagged
// struct or interface with the declared name as the template's argument. The parameters
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
	outputPath    string
	template      *template.Template
	templatePath  string
	templateValue string
	makeName      func(string) string
}

func (_ *Templater) BeginFile(context tagproc.TagContext) error  { return nil }
func (_ *Templater) FinishFile(context tagproc.TagContext) error { return nil }
func (_ *Templater) Finalize() error                             { return nil }

func (ct *Templater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	name := ct.makeName(obj.Name)
	output := filepath.Join(ct.outputPath, name)
	return gotransform.WriteGoTemplate(output, ct.template, obj.Name)
}
