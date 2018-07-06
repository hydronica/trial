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
			r := isInSlice(valX, child...)
			return r == "", r
		}
		r := isInSlice(valX, y)
		return r == "", r
	case reflect.Map:
		if valY.Kind() != reflect.Map {
			return false, fmt.Sprintf("- %T\n+%T", x, y)
		}
		r := isInMap(valX, valY)
		return r == "", r
	}
	return Equal(x, y)
}

func isInMap(parent reflect.Value, child reflect.Value) string {
	result := "-"
	for _, key := range child.MapKeys() {
		if !cmp.Equal(parent.MapIndex(key).Interface(), child.MapIndex(key).Interface()) {
			result += fmt.Sprintf("%v \n", child.MapIndex(key).Interface())
		}
	}
	if result == "-" {
		return ""
	}
	return result
}

func isInSlice(parent reflect.Value, child ...interface{}) string {
	result := "-"
	for _, v := range child {
		found := false
		for i := 0; i < parent.Len(); i++ {
			p := parent.Index(i)
			// always use strings.Contains for string value
			// todo: should this be optional?
			if s, ok := v.(string); ok && p.Kind() == reflect.String && strings.Contains(p.Interface().(string), s) {
				found = true
				break
			}
			if cmp.Equal(p.Interface(), v) {
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
