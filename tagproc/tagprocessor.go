package tagproc

import (
	"go/ast"
	"go/token"
	"gotransform"
	"strings"

	"github.com/pkg/errors"
)

// TagHandler is implemented by all types that want to handle tags embedded in structs.
// A TagHandler gets a chance to react to each file and to each tag embedded in that file.
type TagHandler interface {
	// BeginFile is called whenever a new file is processed.
	BeginFile(context TagContext) error
	// HandleTag is called once for each struct with a tag in each file.
	// The object parameter then points to the definition of the struct in the source file,
	// the literalTag parameter contains the struct tag associated to the field in the struct.
	// It is the empty string if there is no tag attached to the struct field.
	HandleTag(context TagContext, object *ast.Object, literalTag string) error
	// FinishFile is called whenever the processing of a file finishes.
	FinishFile(context TagContext) error
	// Finalize is called when all files have been processed.
	Finalize() error
}

// TagContext contains all the data available to a tag processor.
type TagContext struct {
	File    *ast.File
	FileSet *token.FileSet
	// An import map that maps package scopes to the full package paths. This work under
	// the assumption that a path github.com/foo/bar actually points to a package called
	// bar.
	Imports map[string]string
}

// HandlerMap maps path types to lists of tag handlers.
type HandlerMap = map[string][]TagHandler

// TagProcessor is the FileTransformation that handles all tags.
type TagProcessor struct {
	HandlerMap HandlerMap
}

func New() *TagProcessor {
	return &TagProcessor{make(map[string][]TagHandler)}
}

// AddHandler adds a new handler for the given tag.
// For example:
//    tp.AddHandler("github.com/foo/tags/bartag", handler)
// registers a handler to be called for each tags.bartag typed tag in a struct.
func (tp *TagProcessor) AddHandler(tag string, handler TagHandler) {
	handlers, ok := tp.HandlerMap[tag]
	if !ok {
		tp.HandlerMap[tag] = []TagHandler{handler}
	} else {
		tp.HandlerMap[tag] = append(handlers, handler)
	}
}

// Finalize finalizes all tag handlers.
func (tp *TagProcessor) Finalize() error {
	for _, handlers := range tp.HandlerMap {
		for _, h := range handlers {
			if err := h.Finalize(); err != nil {
				return errors.Wrapf(err, "TagProcessor Finalize")
			}
		}
	}
	return nil
}

// Apply goes through the given file, looks for structs with tags, and calls
// all the respective tag processors on the structs.
func (tp *TagProcessor) Apply(fileContext gotransform.FileContext) error {
	context := TagContext{
		File:    fileContext.File,
		FileSet: fileContext.FileSet,
		Imports: importMap(fileContext.File),
	}
	if err := tp.beginFile(&context); err != nil {
		return errors.Wrapf(err, "TagProcessor Apply/beginFile")
	}
	if err := tp.handleFile(&context); err != nil {
		return errors.Wrapf(err, "TagProcessor Apply/handleFile")
	}
	if err := tp.finishFile(&context); err != nil {
		return errors.Wrapf(err, "TagProcessor Apply/finishFile")
	}
	return nil
}

func (tp *TagProcessor) beginFile(context *TagContext) error {
	for _, handlers := range tp.HandlerMap {
		for _, h := range handlers {
			if err := h.BeginFile(*context); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tp *TagProcessor) handleFile(context *TagContext) error {
	taggedTypes := tp.findTaggedTypes(context.File.Scope, context.Imports)
	for _, s := range taggedTypes {
		if handlers, ok := tp.HandlerMap[s.TagType]; ok {
			for _, h := range handlers {
				if err := h.HandleTag(*context, s.Object, s.LiteralTag); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (tp *TagProcessor) finishFile(context *TagContext) error {
	for _, handlers := range tp.HandlerMap {
		for _, h := range handlers {
			if err := h.FinishFile(*context); err != nil {
				return err
			}
		}
	}
	return nil
}

type taggedDeclaration struct {
	TagType    string
	LiteralTag string
	Object     *ast.Object
}

// findTaggedTypes looks for all structs and interfaces within a given scope that have types from a 'tags' package
// as an anonymous field, removes them, and passes them to specific tag handling functions.
func (tp TagProcessor) findTaggedTypes(scope *ast.Scope, imports map[string]string) []taggedDeclaration {
	output := make([]taggedDeclaration, 0)
	for _, obj := range scope.Objects {
		if obj.Kind != ast.Typ {
			continue
		}
		typeSpec := obj.Decl.(*ast.TypeSpec)
		switch spec := typeSpec.Type.(type) {
		case *ast.StructType:
			tp.findStructTags(obj, spec, imports, &output)
		case *ast.InterfaceType:
			tp.findInterfaceTags(obj, spec, imports, &output)
		}
	}
	return output
}

func (tp TagProcessor) findStructTags(obj *ast.Object, struc *ast.StructType, imports map[string]string, output *[]taggedDeclaration) {
	// find anonymous fields that are of tag-type
	toRemove := make([]int, 0)
	for i, f := range struc.Fields.List {
		if f.Names != nil {
			continue
		}
		isTag, typ := findTag(f.Type, imports)
		if !isTag {
			continue
		}
		// get the struct field tag value
		fieldTag := ""
		if f.Tag != nil {
			// remove the "" from the tag
			fieldTag = f.Tag.Value[1 : len(f.Tag.Value)-1]
		}
		toRemove = append(toRemove, i)
		*output = append(*output, taggedDeclaration{typ, fieldTag, obj})
	}
	// NB: we can only delete afterwards because we don't want to screw with the iteration
	for i := range toRemove {
		removeTag(struc.Fields, toRemove[len(toRemove)-1-i])
	}
}

func (tp TagProcessor) findInterfaceTags(obj *ast.Object, interfac *ast.InterfaceType, imports map[string]string, output *[]taggedDeclaration) {
	// find anonymous fields that are of tag-type
	toRemove := make([]int, 0)
	for i, f := range interfac.Methods.List {
		if f.Names != nil {
			continue
		}
		isTag, typ := findTag(f.Type, imports)
		if !isTag {
			continue
		}
		toRemove = append(toRemove, i)
		// note that interface do not have struct-tags
		*output = append(*output, taggedDeclaration{typ, "", obj})
	}
	// NB: we can only delete afterwards because we don't want to screw with the iteration
	for i := range toRemove {
		removeTag(interfac.Methods, toRemove[len(toRemove)-1-i])
	}
}

func findTag(expr ast.Expr, imports map[string]string) (isTag bool, typ string) {
	path := parseSelector(expr)
	if len(path) == 0 {
		return false, ""
	}
	var importPath string
	if len(path) == 1 {
		// just an identifier
		importPath = imports[""]
		typ = importPath + "/" + path[0]
	} else {
		// selector expression
		importPath = imports[path[0]]
		typ = importPath + "/" + strings.Join(path[1:], "/")
	}
	isTag = strings.HasSuffix(importPath, "tags")
	return
}

// removeTag removes a tag-type from a struct. The tag type is embedded as an anonymous field
// in the struct.
func removeTag(fieldList *ast.FieldList, tagIndex int) {
	fields := fieldList.List
	fieldList.List = append(fields[:tagIndex], fields[tagIndex+1:]...)
}

// reverse reverses a slice of strings.
func reverse(s []string) {
	n := len(s)
	for i := 0; i < n/2; i++ {
		s[i], s[n-i-1] = s[n-i-1], s[i]
	}
}

// parseSelector converts an expression (Ident/Selector) into a slice of strings.
func parseSelector(e ast.Expr) []string {
	current := e
	path := make([]string, 0)
loop:
	for {
		if current == nil {
			break
		}
		switch v := current.(type) {
		case *ast.SelectorExpr:
			current = v.X
			path = append(path, v.Sel.Name)
		case *ast.Ident:
			path = append(path, v.Name)
			break loop
		default:
			break loop
		}
	}
	reverse(path)
	return path
}
