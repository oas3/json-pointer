package ptr

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	prefixErr = errors.New("a JSON Pointer is prefixed by a '/' (%x2F) character")
)

// New creates a JSON Pointer based on the given string.
//
// A JSON Pointer can be represented in a JSON string value. All
// instances of quotation mark '"' (%x22), reverse solidus '\' (%x5C),
// and control (%x00-1F) characters MUST be escaped.
func New(ptr string) (JSONPointer, error) {
	if ptr == "" {
		return JSONPointer{}, nil
	}

	if !isPtr(ptr) {
		return JSONPointer{}, prefixErr
	}

	return JSONPointer{
		references: strings.Split(ptr[1:], "/"),
	}, nil
}

// JSONPointer represents a JavaScript Object Notation (JSON) Pointer.
type JSONPointer struct {
	references []string
}

// Delete removes the value corresponding with the JSON Pointer.
//
// If the document is an array and the pointer removes an element at the
// root level, an error will be returned. Since the document can not be
// updated since the array needs te be recreated. (see examples)
func (ptr *JSONPointer) Delete(document interface{}) (interface{}, error) {
	doc, _, err := ptr.traverse(nil, document, true)
	return doc, err
}

// Get returns the value corresponding with the JSON Pointer.
func (ptr *JSONPointer) Get(document interface{}) (interface{}, error) {
	doc, _, err := ptr.traverse(nil, document, false)
	return doc, err
}

// Set assigns a the given value to the JSON Pointers location.
func (ptr *JSONPointer) Set(value, document interface{}) (interface{}, reflect.Kind, error) {
	return ptr.traverse(value, document, false)
}

func (ptr *JSONPointer) String() string {
	if len(ptr.references) == 0 {
		return ""
	}
	return fmt.Sprintf("/%s", strings.Join(ptr.references, "/"))
}

// traverse iterates over the json document based on the JSON Pointer.
//
// value:    the value that needs to be set.
// document: the json document to search in.
// remove:   indicates whether the value needs to be removed.
func (ptr *JSONPointer) traverse(value, document interface{}, remove bool) (interface{}, reflect.Kind, error) {
	kind := reflect.Invalid
	if len(ptr.references) == 0 {
		return document, kind, nil
	}

	// current 'points' at the field the for-loop is currently at.
	current := document

	var (
		// keep track of previous nodes and tokens to be able to remove elements from an
		// arrays and update the corresponding maps with the new array.
		nodes  = make([]interface{}, len(ptr.references))
		tokens = make([]string, len(ptr.references))
	)

	// Evaluation of a JSON Pointer begins with a reference to the root
	// value of a JSON document and completes with a reference to some
	// value within the document.  Each reference token in the JSON
	// Pointer is evaluated sequentially.
	for i, tk := range ptr.references {
		var end bool // indicates if tk is the last token of the pointer
		if i == len(ptr.references)-1 {
			end = true
		}

		// keep track of visited nodes
		nodes[i] = current
		tokens[i] = tk

		switch t := current.(type) {
		case []interface{}:
			if i == 0 && len(ptr.references) == 1 {
				return nil, reflect.Slice, fmt.Errorf("can not delete from an array at root level")
			}

			// Raise an error condition if it fails to resolve a
			// concrete value for any of the JSON pointer's reference
			// tokens.
			idx, err := strconv.Atoi(tk)
			if err != nil {
				return nil, reflect.Slice, fmt.Errorf("invalid array index %q", tk)
			}
			if idx < 0 || len(t) <= idx {
				return nil, reflect.Slice, fmt.Errorf("out of bound [0,%d[, index %q", len(t), idx)
			}

			// The reference token MUST contain either:
			// - Characters comprised of digits (note that leading zeros
			//   are not allowed) that represent an unsigned base-10
			//   integer value, making the new referenced value the
			//   array element with the zero-based index identified by
			//   the token.
			// - Exactly the single character "-", making the new
			//   referenced value the (nonexistent) member after the
			//   last array element.
			current = t[idx]
			if end {
				// replace/set value
				if value != nil {
					t[idx] = value
					break
				}
				// remove value
				if remove {
					t = append(t[:idx], t[idx+1:]...)
					// update previous map that contains this slice
					if 0 < i {
						nodes[i-1].(map[string]interface{})[tokens[i-1]] = t
					}
				}
			}
		case map[string]interface{}:
			// Evaluation of each reference token begins by decoding any
			// escaped character sequence.  This is performed by first
			// transforming any occurrence of the sequence '~1' to '/',
			// and then transforming any occurrence of the sequence '~0'
			// to '~'.
			tk = decode(tk)
			if _, ok := t[tk]; ok {
				// The new referenced value is the object member with
				// the name identified by the reference token.
				current = t[tk]
				if end {
					// replace/set value
					if value != nil {
						t[tk] = value
						break
					}
					// remove value
					if remove {
						delete(t, tk)
					}
				}
			} else if end && value != nil {
				t[tk] = value
			} else {
				return nil, reflect.Map, fmt.Errorf("object does not have the key %q", tk)
			}
		default:
			return nil, reflect.ValueOf(current).Kind(), fmt.Errorf("invalid token reference %q", tk)
		}
	}

	return current, reflect.ValueOf(current).Kind(), nil
}

func isPtr(ptr string) bool {
	return ptr != "" && ptr[0] == '/'
}

// Because the characters '~' (%x7E) and '/' (%x2F) have special meanings in JSON Pointer,
// '~' needs to be encoded as '~0' and '/' needs to be encoded as '~1' when these
// characters appear in a reference token.
//
// By performing the substitutions in this order, an implementation
// avoids the error of turning '~01' first into '~1' and then into '/',
// which would be incorrect (the string '~01' correctly becomes '~1'
// after transformation).
func decode(tk string) string {
	tk = strings.Replace(tk, "~1", "/", -1) // ! order is important
	tk = strings.Replace(tk, "~0", "~", -1)
	return tk
}
