// tagproc provides utilities to quickly process declarations in a Go codebase.
// The idea is as follows: Both struct and interface declarations may have embedded
// tag member that are processed and removed by the TagProcessor.
// As an example, consider
//     type X struct {
//	       tags.ImportantStruct
//     }
// The TagProcessor allows you to register handlers for the tags.ImportantStruct type
// that are called for all declarations that have an embedded (i.e. anonymous) member
// of this kind. A type may be used as a tag whenever it is defined in a package named
// tags.
package tagproc
