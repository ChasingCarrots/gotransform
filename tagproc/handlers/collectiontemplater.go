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
		collected:  nil,
	}
}

type CollectionTemplater struct {
	template   *template.Template
	outputPath string
	collected  []templateEntry
}

func (ct *CollectionTemplater) BeginFile(context tagproc.TagContext) error  { return nil }
func (ct *CollectionTemplater) FinishFile(context tagproc.TagContext) error { return nil }

func (ct *CollectionTemplater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	entry, err := makeTemplateEntry(obj.Name, tagLiteral)
	if err != nil {
		return err
	}
	ct.collected = append(ct.collected, entry)
	return nil
}

func (ct *CollectionTemplater) Finalize() error {
	return gotransform.WriteGoTemplate(ct.outputPath, ct.template, ct.collected)
}
