package tagproc

import (
	"go/ast"
	"strings"
)

// importMap produces a map of package scopes to their full paths. This works under
// the assumption that the packages are defined in folders with identical names.
// For example, the imports
//   "github.com/foo/bar"
//   test "github.com/foo/baz"
//
// will produce a map imports satisfying
//   imports["bar"] == "github.com/foo/bar"
//   imports["test"] == "github.com/foo/baz"
// independently of whether the package defined in github.com/foo/bar is actually
// called bar.
func importMap(f *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, i := range f.Imports {
		path, name := extractImport(i)
		imports[name] = path
	}
	imports[""] = f.Name.Name
	return imports
}

// extractImport divides an import into its path and the expected name of the
// imported package.
func extractImport(is *ast.ImportSpec) (path, name string) {
	path = is.Path.Value
	path = path[1 : len(path)-1] // remove wrapping ""
	if is.Name != nil {
		name = is.Name.Name
	} else {
		idx := strings.LastIndexAny(path, "/")
		if idx == -1 {
			name = path
		} else {
			name = path[(idx + 1):]
		}
	}
	return
}
