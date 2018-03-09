package handlers

import (
	"go/ast"
	"gotransform"
	"gotransform/tagproc"
	"text/template"
)

// NewCollectionTemplater creates a TagHandler that will collect the names of all tagged declarations
// and pass them to a Go template as a collection.
// The outputPath parameter determines the path of the output file.
func NewCollectionTemplater(outputPath string, template *template.Template) *CollectionTemplater {
	return &CollectionTemplater{
		outputPath: outputPath,
		template:   template,
		collected:  nil,
	}
}

type CollectionTemplater struct {
	template   *template.Template
	outputPath string
	collected  []string
}

func (ct *CollectionTemplater) BeginFile(context tagproc.TagContext) error  { return nil }
func (ct *CollectionTemplater) FinishFile(context tagproc.TagContext) error { return nil }

func (ct *CollectionTemplater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	ct.collected = append(ct.collected, obj.Name)
	return nil
}

func (ct *CollectionTemplater) Finalize() error {
	return gotransform.WriteGoTemplate(ct.outputPath, ct.template, ct.collected)
}
