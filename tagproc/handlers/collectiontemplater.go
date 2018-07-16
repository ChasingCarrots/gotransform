package handlers

import (
	"go/ast"
	"text/template"

	"github.com/chasingcarrots/gotransform"
	"github.com/chasingcarrots/gotransform/tagproc"
)

// NewCollectionTemplater creates a TagHandler that will collect the names and tags of all tagged
// declarations and pass them to a Go template as a collection.
// The outputPath parameter determines the path of the output file.
func NewCollectionTemplater(outputPath string, template *template.Template) *CollectionTemplater {
	return &CollectionTemplater{
		outputPath: outputPath,
		template:   template,
	}
}

type CollectionTemplater struct {
	templateCollection
	template   *template.Template
	outputPath string
	delay      bool
}

func (ct *CollectionTemplater) Delay() {
	ct.delay = true
}

func (ct *CollectionTemplater) BeginFile(context tagproc.TagContext) error  { return nil }
func (ct *CollectionTemplater) FinishFile(context tagproc.TagContext) error { return nil }

func (ct *CollectionTemplater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	return ct.addEntry(context, obj, tagLiteral)
}

func (ct *CollectionTemplater) Finalize() error {
	if !ct.delay {
		return ct.WriteTemplate()
	}
	return nil
}

func (ct *CollectionTemplater) WriteTemplate() error {
	return gotransform.WriteGoTemplate(ct.outputPath, ct.template, ct.entries)
}
