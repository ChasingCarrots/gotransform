package writeout

import (
	"bytes"
	"go/format"
	"gotransform"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type writeOut struct {
	outputPath, suffix string
}

// Transformation can be used at any stage to write each file out with an additional
// suffix. Files are written out with their relative path, so a file dir1/dir2/file.go
// will end up in outputPath/dir1/dir2/file.go
func Transformation(outputPath, suffix string) gotransform.FileTransformation {
	return &writeOut{outputPath: outputPath, suffix: suffix}
}

func (wo *writeOut) Apply(context gotransform.FileContext) error {
	var buf bytes.Buffer
	if err := format.Node(&buf, context.FileSet, context.File); err != nil {
		return errors.Wrap(err, "WriteOut: Failed to format file")
	}
	path := filepath.Join(wo.outputPath, addSuffix(context.RelativePath, wo.suffix))
	if err := gotransform.WriteGoFile(path, &buf); err != nil {
		return errors.Wrap(err, "Write out: Failed to write file")
	}
	return nil
}

func (wo *writeOut) Prepare() error {
	return errors.Wrap(PrepareDir(wo.outputPath, wo.suffix), "Write out: Failed to prepare directory")
}

func (wo *writeOut) Finalize() error { return nil }

// addSuffix adds a suffix to a file name, right before its extension
func addSuffix(filename, suffix string) string {
	extension := filepath.Ext(filename)
	n := len(filename)
	filename = filename[:n-len(extension)]
	return filename + suffix + extension
}

// PrepareDir ensures that the given directory exists and removes all files with
// the specified suffix from it.
func PrepareDir(path, suffix string) error {
	suffix = suffix + ".go"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(path, os.ModePerm)
		} else {
			return err
		}
	} else {
		err := DeleteFiles(path, func(p string) bool { return strings.HasSuffix(p, suffix) })
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFiles recursively walks the file system from the given starting part
// and deletes all files whose file names match a predicate.
func DeleteFiles(path string, predicate func(string) bool) error {
	deleter := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if predicate(path) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	}
	return filepath.Walk(path, deleter)
}
