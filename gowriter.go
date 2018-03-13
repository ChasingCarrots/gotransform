package gotransform

import (
	"bytes"
	"io"
	"os"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/tools/imports"
)

// WriteGoFile writes the contents of a reader to the given path, formatting it and
// running go imports on the output.
func WriteGoFile(path string, reader io.Reader) error {
	buf, ok := reader.(*bytes.Buffer)
	if !ok {
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
	}

	options := &imports.Options{
		TabWidth:   8,
		TabIndent:  true,
		Comments:   true,
		Fragment:   true,
		AllErrors:  true,
		FormatOnly: false,
	}
	formattedCode, formatErr := imports.Process(path, buf.Bytes(), options)
	if formatErr != nil {
		formattedCode = buf.Bytes()
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "WriteGoFile: File creation failed for %s", path)
	}
	defer file.Close()

	_, err = file.Write(formattedCode)
	if err != nil {
		return errors.Wrapf(err, "WriteGoFile: Write failed for %s", path)
	}
	return errors.Wrapf(formatErr, "WriteGoFile: Formatting failed for", path)
}

// WriteGoTemplate applies the given template to the value and writes it out as a Go
// file, formatting it and running go imports on it.
func WriteGoTemplate(path string, tmpl *template.Template, value interface{}) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, value); err != nil {
		return errors.Wrapf(err, "WriteGoTemplate: Failed to write template to %s", path)
	}
	return WriteGoFile(path, &buf)
}
