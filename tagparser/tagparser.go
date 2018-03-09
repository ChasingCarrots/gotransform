package tagparser

import (
	"fmt"
	"strings"
)

// Parse takes a struct field tag in Go's standard format,
// that is key:"value" pairs separated by spaces, and turns it into a map
// from keys to values, removing all "" around values in the process.
// For example, a field tag such as
//     `protobuf:"1"`
// would be turned into the map
//      { "protobuf": [ "1" ] }
func Parse(fieldTag string) (result map[string][]string, err error) {
	fieldTag = strings.TrimSpace(fieldTag)
	result = make(map[string][]string)
	for len(fieldTag) > 0 {
		delimPos := strings.IndexRune(fieldTag, ':')
		if delimPos < 0 {
			return result, fmt.Errorf("Invalid format in struct tag: %s", fieldTag)
		}
		key := fieldTag[:delimPos]
		fieldTag = strings.TrimSpace(fieldTag[delimPos+1:])
		valueStartPos := strings.IndexRune(fieldTag, '"')
		if valueStartPos != 0 {
			return result, fmt.Errorf("Invalid format in struct tag: %s", fieldTag)
		}
		fieldTag = fieldTag[1:]
		valueEndPos := strings.IndexRune(fieldTag, '"')
		if valueEndPos < 0 {
			return result, fmt.Errorf("Invalid format in struct tag: %s", fieldTag)
		}
		value := fieldTag[:valueEndPos]
		existingValues, ok := result[key]
		if !ok {
			result[key] = []string{value}
		} else {
			result[key] = append(existingValues, value)
		}
		fieldTag = strings.TrimSpace(fieldTag[valueEndPos+1:])
	}
	return result, nil
}

// Unique maps each value in a list-valued map to its first element.
// This is useful in combination with Parse whenever you are sure that
// there is only a single value associated to each key.
func Unique(values map[string][]string) map[string]string {
	keyValuePairs := make(map[string]string, len(values))
	for k, v := range values {
		if len(v) > 0 {
			keyValuePairs[k] = v[0]
		}
	}
	return keyValuePairs
}
