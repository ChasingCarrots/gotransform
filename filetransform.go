package gotransform

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// FileTransformation represents a general kind of in-place transformation that
// is applied file-wise. It doesn't have to do anything to the files themselves; it
// might just as well simply collect information etc.
type FileTransformation interface {
	// Prepare is called before any processing takes place.
	Prepare() error
	// Apply is where the action happens; called once for every file.
	Apply(FileContext) error
	// Finalize is called once all files have been visited. Use this to write out
	// summaries that depend on all files.
	Finalize() error
}

// FileContext contains all the data that is passed down to every call to Apply
// of a FileTransformation.
type FileContext struct {
	File         *ast.File
	FileSet      *token.FileSet
	RelativePath string
}

// Apply recursively walks the file-system, starting at the given input path,
// and applies the given transformations to all go-files encountered on the way.
// For each file, the transformations are executed in the order that they have in the
// array.
func Apply(inputPath string, transformations []FileTransformation) error {
	// parse the files
	collection := make([]FileContext, 0)
	fileset := token.NewFileSet()
	processFile := parseFiles(inputPath, fileset, &collection)
	if err := filepath.Walk(inputPath, processFile); err != nil {
		return errors.Wrapf(err, "Apply FileWalk")
	}

	// apply the transformations
	for _, t := range transformations {
		for _, context := range collection {
			if err := t.Apply(context); err != nil {
				return errors.Wrapf(err, "Apply: Failed to transform file %s", context.RelativePath)
			}
		}
	}

	// finalize transformations
	for _, t := range transformations {
		if err := t.Finalize(); err != nil {
			return errors.Wrapf(err, "Apply Finalize")
		}
	}
	return nil
}

func parseFiles(inputPath string, fileSet *token.FileSet, collection *[]FileContext) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// We don't do anything on directories, but note that this does *not* mean that
		// we do not descend into directories.
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		context, err := readFile(fileSet, inputPath, path)
		if err != nil {
			return err
		}
		*collection = append(*collection, *context)
		return nil
	}
}

func readFile(fileset *token.FileSet, inputPath, path string) (*FileContext, error) {
	file, err := parser.ParseFile(fileset, path, nil, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse file %s", path)
	}

	relativePath, err := filepath.Rel(inputPath, path)
	if err != nil {
		// this should never yield an error.
		panic(fmt.Sprintf("Error with filepath.Rel: %v", err))
	}
	return &FileContext{file, fileset, relativePath}, nil
}
