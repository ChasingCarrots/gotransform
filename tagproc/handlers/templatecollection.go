package handlers

import (
	"go/ast"

	"github.com/chasingcarrots/gotransform/tagparser"
	"github.com/chasingcarrots/gotransform/tagproc"
)

type templateCollection struct {
	entries        []templateEntry
	templateMapper func(*templateEntry)
}

type templateEntry struct {
	Name    string
	Package string
	Tags    map[string]string
	Data    map[string]interface{}
}

// SetTemplateMapper allows you to specify a function that will be applied to the map containing the
// data that will be passed to the template engine. Use this to provide custom data.
// It can be accessed in the template via `Data.yourname`
func (tc *templateCollection) SetTemplateMapper(mapper func(*templateEntry)) {
	tc.templateMapper = mapper
}

func (tc *templateCollection) addEntry(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	entry, err := makeTemplateEntry(context, obj, tagLiteral)
	if err != nil {
		return err
	}
	if tc.templateMapper != nil {
		tc.templateMapper(&entry)
	}
	tc.entries = append(tc.entries, entry)
	return nil
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
		Data:    make(map[string]interface{}),
	}, nil
}
