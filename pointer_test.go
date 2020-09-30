package ptr_test

import (
	"encoding/json"
	"fmt"
	ptr "github.com/oas3/json-pointer"
	"reflect"
	"testing"
)

// RFC: https://tools.ietf.org/html/rfc6901#section-5
func TestExample(t *testing.T) {
	const JSON = `
{
  "foo": ["bar", "baz"],
  "": 0,
  "a/b": 1,
  "c%d": 2,
  "e^f": 3,
  "g|h": 4,
  "i\\j": 5,
  "k\"l": 6,
  " ": 7,
  "m~n": 8
}
`
	var jsonDocument map[string]interface{}
	if err := json.Unmarshal([]byte(JSON), &jsonDocument); err != nil {
		t.Error(err)
	}

	var (
		pointers = []string{
			"",
			"/foo",
			"/foo/0",
			"/",
			"/a~1b",
			"/c%d",
			"/e^f",
			"/g|h",
			"/i\\j",
			"/k\"l",
			"/ ",
			"/m~0n",
		}
		values = []interface{}{
			jsonDocument,
			[]interface{}{"bar", "baz"},
			"bar",
			// json converts integers to float64
			0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0,
		}
	)

	for i, p := range pointers {
		jsonPtr, err := ptr.New(p)
		if err != nil {
			t.Error(err)
		}

		v, err := jsonPtr.Get(jsonDocument)
		if !reflect.DeepEqual(values[i], v) {
			t.Errorf("expected %v, got %v", values[i], v)
		}
	}
}

func ExampleJSONPointer_Delete() {
	doc := map[string]interface{}{
		"foo": []interface{}{"bar", "baz"},
	}
	p, _ := ptr.New("/foo/0")
	r, _ := p.Delete(doc)
	fmt.Println(doc)
	fmt.Println(r)

	// Output:
	// map[foo:[baz]]
	// bar
}

func ExampleJSONPointer_Delete_error() {
	doc := []interface{}{"bar", "baz"}
	p, _ := ptr.New("/0")
	_, err := p.Delete(doc)
	fmt.Println(err)

	// Output:
	// can not delete from an array at root level
}

func ExampleJSONPointer_Get() {
	doc := map[string]interface{}{
		"foo": []interface{}{"bar", "baz"},
	}

	var p ptr.JSONPointer

	p, _ = ptr.New("")
	r0, _ := p.Get(doc)
	fmt.Println(r0)

	p, _ = ptr.New("/foo")
	r1, _ := p.Get(doc)
	fmt.Println(r1)

	p, _ = ptr.New("/foo/0")
	r2, _ := p.Get(doc)
	fmt.Println(r2)

	// Output:
	// map[foo:[bar baz]]
	// [bar baz]
	// bar
}
