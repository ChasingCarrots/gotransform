package handlers

import (
	"go/ast"
	"go/token"
	"gotransform/tagproc"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/tools/go/ast/astutil"
)

// NewFieldAdder creates a new tag handler that adds a field to whatever struct the target
// tag appears on. The name, typ, and importPath parameters are used to give the name and type
// of the field and the import path for the type. The tag parameter sets the field's tag.
// This handler will also add imports to the target type if necessary. Set the import path to
// the empty string to prevent this.
func NewFieldAdder(name, typ, importPath, tag string) *FieldAdder {
	return &FieldAdder{
		fieldName:  name,
		fieldType:  typ,
		importPath: importPath,
		fieldTag:   tag,
	}
}

type FieldAdder struct {
	importPath string
	fieldName  string
	fieldType  string
	fieldTag   string
}

func (_ *FieldAdder) BeginFile(context tagproc.TagContext) error  { return nil }
func (_ *FieldAdder) FinishFile(context tagproc.TagContext) error { return nil }
func (_ *FieldAdder) Finalize() error                             { return nil }

func (fa *FieldAdder) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	typeSpec := obj.Decl.(*ast.TypeSpec)
	struc, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return errors.Errorf("The tagged object is no struct! Object name: %s", obj.Name)
	}
	if len(fa.importPath) > 0 {
		astutil.AddImport(context.FileSet, context.File, fa.importPath)
	}
	addField(struc, fa.fieldName, fa.fieldType, fa.fieldTag)
	return nil
}

func addField(struc *ast.StructType, name, typ, tag string) {
	field := new(ast.Field)
	field.Type = makeSelector(typ)
	field.Names = []*ast.Ident{makeIden(name)}
	if struc.Fields == nil {
		struc.Fields = new(ast.FieldList)
	}
	if len(tag) > 0 {
		field.Tag = new(ast.BasicLit)
		field.Tag.Value = `"` + tag + `"`
		field.Tag.Kind = token.STRING
	}
	struc.Fields.List = append(struc.Fields.List, field)
}

func makeIden(value string) *ast.Ident {
	return &ast.Ident{
		Name: value,
	}
}

func makeSelector(path string) ast.Expr {
	idx := strings.Index(path, ".")
	fst := path[:idx]
	snd := path[idx+1:]
	return &ast.SelectorExpr{
		X:   makeIden(fst),
		Sel: makeIden(snd),
	}
}
