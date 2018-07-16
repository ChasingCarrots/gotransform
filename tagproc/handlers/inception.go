package handlers

import (
	"bytes"
	"fmt"
	"go/ast"
	"text/template"

	"github.com/pkg/errors"

	"github.com/chasingcarrots/gotransform"
	"github.com/chasingcarrots/gotransform/tagproc"

	"os/exec"
)

// NewInceptionTemplater creates a handler that applies the inception principle:
// c.f. https://journal.paul.querna.org/articles/2014/03/31/ffjson-faster-json-in-go/
// This means: It will apply the template to all collected tags like the collection
// templater handler. When it has done that, it will proceed by running the generated
// file as a Go file, allowing you to use reflection on types from your own package.
func NewInceptionTemplater(outputPath string, template *template.Template, parameters ...string) *InceptionTemplater {
	return &InceptionTemplater{
		outputPath: outputPath,
		template:   template,
		parameters: parameters,
	}
}

type InceptionTemplater struct {
	templateCollection
	template   *template.Template
	outputPath string
	parameters []string
	delay      bool
}

// Delay prevents any generation of files in the Finalize function and instead requires you to call
// PerformInception at another time that is more suitable for your uses. This might be required when
// using an inception-based approach.
func (ct *InceptionTemplater) Delay() {
	ct.delay = true
}

func (ct *InceptionTemplater) BeginFile(context tagproc.TagContext) error  { return nil }
func (ct *InceptionTemplater) FinishFile(context tagproc.TagContext) error { return nil }

func (ct *InceptionTemplater) HandleTag(context tagproc.TagContext, obj *ast.Object, tagLiteral string) error {
	return ct.addEntry(context, obj, tagLiteral)
}

func (ct *InceptionTemplater) Finalize() error {
	if !ct.delay {
		return ct.PerformInception()
	}
	return nil
}

func (ct *InceptionTemplater) PerformInception() error {
	if err := gotransform.WriteGoTemplate(ct.outputPath, ct.template, ct.entries); err != nil {
		return err
	}

	params := []string{"run", ct.outputPath}
	params = append(params, ct.parameters...)

	cmd := exec.Command("go", params...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "Inception cmd.Run() failed with stderr: %s\nstdout:", stderr.String(), stdout.String())
	}
	fmt.Println(stdout.String())
	return nil
}
