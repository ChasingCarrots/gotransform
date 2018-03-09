package gotransform

import (
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

type genericTransformation struct {
	prepare  func() error
	apply    func(FileContext) error
	finalize func() error
}

func (gt *genericTransformation) Prepare() error {
	if gt.prepare == nil {
		return nil
	}
	return gt.Prepare()
}

func (gt *genericTransformation) Apply(context FileContext) error {
	if gt.apply == nil {
		return nil
	}
	return gt.apply(context)
}

func (gt *genericTransformation) Finalize() error {
	if gt.finalize == nil {
		return nil
	}
	return gt.finalize()
}

// DropBuildIgnore removes // +build ignore comments from a file.
func DropBuildIgnore() FileTransformation {
	return &genericTransformation{
		apply: func(context FileContext) error {
			f := context.File
			if len(f.Comments) > 0 && strings.Trim(f.Comments[0].Text(), " ") == "+build ignore" {
				f.Comments = f.Comments[1:]
			}
			return nil
		},
	}
}

// ChangePackageName changes the name of the package in a file.
func ChangePackageName(name string) FileTransformation {
	return &genericTransformation{
		apply: func(context FileContext) error {
			context.File.Name.Name = name
			return nil
		},
	}
}

// AddImport adds an import to a file, if it is not already present.
func AddImport(path string) FileTransformation {
	return &genericTransformation{
		apply: func(context FileContext) error {
			astutil.AddImport(context.FileSet, context.File, path)
			return nil
		},
	}
}

// AddNamedImport adds a named import to a file, if it is not already present.
func AddNamedImport(name, path string) FileTransformation {
	return &genericTransformation{
		apply: func(context FileContext) error {
			astutil.AddNamedImport(context.FileSet, context.File, name, path)
			return nil
		},
	}
}
