# gotransform
A Go package to help with Go code generation.

## Example
The main idea of `gotransform` is to define a pipeline of transformations that are uniformly applied to a set of Go-files to yield new Go files. One such pipeline might look like this: 

```golang
// define a pipeline of transformation on the Go-files
transformations := []gotransform.FileTransformation{
    // first change the package name to components
    gotransform.ChangePackageName("components"),
    // then drop any +build ignore comments
    gotransform.DropBuildIgnore(),
    // add an import to the file
    gotransform.AddImport("github.com/chasingcarrots/gotransform"),
    // finally, write out all transformed files to the output directory with
    // their relative paths, adding a "_gen" suffix to each file to mark it
    // as automatically generated.
    // In this process, all go files are formatted and goimports is applied to
    // them.
    // writeout is a subpackage of gotransform with some helper function to
    // write and delete generated files.
    writeout.Transformation(outputPath, "_gen"),
}
// apply the transformations to all Go files in the input directory (recursively)
err := gotransform.Apply(inputPath, transformations)
```

The transformations are applied in the order they are defined in the transformation list.

## Custom Transformations
It is easy to specify custom transformations, just implement the `FileTransformation` interface found in `filetransform.go`:

```golang
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
```

The `FileContext` struct allows you to operate on the AST of the current file, applying your transformations as necessary.

## Tag-based processing
A key component of `gotransform` is the `tagproc` subpackage that allows you to do tag-based processing on structs and interfaces. A *tag* is an embedded member in a struct or interface that lives in any kind of tag namespace. For example, consider the following definition of a tag:

```golang
package tags
// Because the package is called 'tags', the tag processor will recognize it.
// For the sake of discussion, assume that the tag package is defined in
//   github.com/chasingcarrots/gotransform/tags
// The last part of the path ('tags') is important, since the tool uses the
// import-path to determine the package name.

// CollectMe is a tag to be used on structs to mark them as interesting for collection
type CollectMe interface {}
```

This tag can now be used to mark structs and interface of interest in your files:

```golang
package otherpkg

import (
    // you can rename your imports, the tool will still be able to identify the
    // tags package
    mytags "github.com/chasingcarrots/gotransform/tags"
)

type ImportantStruct struct {
    // Mark ImportantStruct with a CollectMe tag. You can use Go's field tags to embed
    // meta data that is specific to you tag.
    mytags.CollectMe `YourMetaData:"Here"`
}

type OtherImportantStruct struct {
    mytags.CollectMe
}
```

Finally, we need to run the tag processor as a file transformation to see any effects. This
tag processor can specify *tag handlers* for each tag.

```golang
// this handler collects the names of all tagged structs and interfaces and executes a Go
// text template on the result.
collector := handlers.NewCollectionTemplater(collectedData, template)

transformations := []gotransform.FileTransformation{
    // The tag processor transformation goes through all files and looks for structs and interfaces
    // with specific tags. It then removes all tags from structs and interfaces and calls the tag
    // handlers specified here for each tag.
    &tagproc.TagProcessor{
			tagproc.HandlerMap{
				"github.com/chasingcarrots/gotransform/tags/CollectMe": {
					&collector
				},
			},
		},
    writeout.Transformation(outputPath, "_gen"),
}
// apply the transformations to all Go files in the input directory (recursively)
err := gotransform.Apply(inputPath, transformations)
```

Assume that the Go text-template used above looked like this:
```
const (
{{ range . }}
    {{.}} int = iota
{{- end }}
)
```

This will produce a file at the `collectedData` path that looks like this:
```golang
const (
    ImportantStruct int = iota
    OtherImportantStruct int = iota
)
```
Additionally, it will write out the input files with the tags stripped away to `outputPath`.

### Writing New Tag Handlers
`gotransform` comes with a few tag handlers that we found helpful (see `tagproc/handlers`), but it is easy to implement your own. Simply implement the `TagHandler` interface from the `tagproc` subpackage.