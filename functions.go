package trial

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// ContainsFn determines if y is a subset of x.
// x is a string -> y is a string that is equal to or a subset of x (string.Contains)
// x is a slice or array -> y is contained in x
// x is a map -> y is a map and is contained in x
func ContainsFn(x, y interface{}) (bool, string) {
	// if nothing is expected we have a match
	if y == nil {
		return true, ""
	}
	valX := reflect.ValueOf(x)
	valY := reflect.ValueOf(y)
	switch valX.Kind() {
	case reflect.String:
		s, ok := y.(string)
		if !ok {
			if v, ok := y.(fmt.Stringer); ok {
				s = v.String()
			} else {
				return false, fmt.Sprintf("- %T\n+%T", x, y)
			}
		}
		return strings.Contains(x.(string), s), ""
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if valY.Kind() == reflect.Slice {
			child := make([]interface{}, valY.Len())
			for i := 0; i < valY.Len(); i++ {
				child[i] = valY.Index(i).Interface()
			}
			r := containsDiffSlice(valX, child...)
			return r == "", r
		}
		r := containsDiffSlice(valX, y)
		return r == "", r
	case reflect.Map:
	}
	return false, ""
}

func containsDiffSlice(parent reflect.Value, child ...interface{}) string {
	result := "-"
	for _, v := range child {
		found := false
		for i := 0; i < parent.Len(); i++ {
			if cmp.Equal(parent.Index(i).Interface(), v) {
				found = true
				break
			}
		}
		if !found {
			result += fmt.Sprintf(" %v\n", v)
		}
	}
	if result == "-" {
		result = ""
	}
	return result
}

// Equal use the cmp.Diff method to display differences between two interfaces
// and check for equality
func Equal(actual, expected interface{}) (bool, string) {
	var opts []cmp.Option
	t := reflect.TypeOf(actual)
	if t == nil {
	} else if t.Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(actual))
	} else if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(reflect.ValueOf(actual).Elem().Interface()))
	}
	t = reflect.TypeOf(expected)
	if t == nil {
	} else if t.Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(expected))
	} else if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(reflect.ValueOf(expected).Elem().Interface()))
	}
	r := cmp.Diff(actual, expected, opts...)
	return r == "", r
}
