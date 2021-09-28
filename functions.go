package trial

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Contains determines if y is a subset of x.
// x is a string -> y is a string that is equal to or a subset of x (string.Contains)
// x is a slice or array -> y is contained in x
// x is a map -> y is a map and is contained in x
func Contains(x, y interface{}) (bool, string) {
	// if nothing is expected we have a match
	if y == nil {
		return true, ""
	}
	r := contains(x, y)
	if r == nil {
		return true, ""
	}
	return false, r.String()
}

const (
	SubStrings = iota
	SubSlices
	SubMaps
)

// ContainsOpt allow configurable options to the contains method
// 1. Check for sub-strings ("abc" -> "abcdefg")
// 2. Check for sub-slices (["a"] -> ["a","b","c"])
// 3. Check for sub-maps
// 4. use regex match as a sub-string check
/*
func ContainsOpt(o interface{}) CompareFunc {
	return Contains
}
*/

func contains(x, y interface{}) differ {
	valX := reflect.ValueOf(x)
	valY := reflect.ValueOf(y)
	switch valX.Kind() {
	case reflect.String:
		s, ok := y.(string)
		if !ok {
			if v, ok := y.(fmt.Stringer); ok {
				s = v.String()
			} else {
				arr, ok := y.([]string)
				if !ok {
					return newMessagef("type mismatch %T %T", x, y)
				}
				v := []string{valX.String()}
				arrI := make([]interface{}, len(arr))
				for i, v := range arr {
					arrI[i] = v
				}
				return isInSlice(reflect.ValueOf(v), arrI...)

			}
		}
		if strings.Contains(valX.String(), s) {
			return nil
		}
		return newDiff(x, s)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if valY.Kind() == reflect.Slice || valY.Kind() == reflect.Array {
			child := make([]interface{}, valY.Len())
			for i := 0; i < valY.Len(); i++ {
				child[i] = valY.Index(i).Interface()
			}
			if d := isInSlice(valX, child...); d != nil {
				return newDiffMsg(x, y, d.String())
			}
			return nil
		}
		if d := isInSlice(valX, y); d != nil {
			return newDiffMsg(x, y, d.String())
		}
		return nil
	case reflect.Map:
		if valY.Kind() != reflect.Map {
			return newMessagef("type mismatch %T %T", x, y)

		}
		if d := isInMap(valX, valY); d != nil {
			return newDiffMsg(x, y, d.String())
		}
		return nil
	}
	isEqual, s := Equal(x, y)
	if isEqual {
		return nil
	}
	return newMessagef(s)
}

func isInMap(parent reflect.Value, child reflect.Value) differ {
	d := &mapDiff{values: make(map[interface{}][]string, 0)}
	for _, key := range child.MapKeys() {
		p := parent.MapIndex(key)
		if !p.IsValid() {
			d.values[key] = make([]string, 0)
			continue
		}
		c := child.MapIndex(key)
		if ok := contains(p.Interface(), c.Interface()); ok != nil {
			d.values[key] = append(d.values[key], ok.String())
		}
	}
	return d.diffOrNil()
}

func isInSlice(parent reflect.Value, child ...interface{}) differ {
	c := &collection{
		found:   make([]interface{}, 0),
		missing: make([]interface{}, 0),
	}
	for _, v := range child {
		found := false
		for i := 0; i < parent.Len(); i++ {
			p := parent.Index(i)
			if contains(p.Interface(), v) == nil {
				found = true
				c.found = append(c.found, v)
				break
			}
		}
		if !found {
			c.missing = append(c.missing, v)
		}
	}
	if len(c.missing) > 0 {
		return c
	}
	return nil
}

// Equal use the cmp.Diff method to check equality and display differences.
// This method checks all unexpected values
func Equal(actual, expected interface{}) (bool, string) {
	fn := EqualOpt(AllowAllUnexported, EquateEmpty)
	return fn(actual, expected)
}

// EqualOpt allow easy customization of the cmp.Equal method.
// see below for a list of supported options
func EqualOpt(optFns ...func(i interface{}) cmp.Option) func(actual, expected interface{}) (bool, string) {
	return func(actual, expected interface{}) (bool, string) {
		opts := make([]cmp.Option, 0)
		for _, fn := range optFns {
			opts = append(opts, fn(actual))
		}

		r := cmp.Diff(actual, expected, opts...)
		return r == "", r
	}
}

// AllowAllUnexported sets cmp.Diff to allow all unexported (private) variables
func AllowAllUnexported(i interface{}) cmp.Option {
	return cmp.AllowUnexported(findAllStructs(i)...)
}

// IgnoreAllUnexported sets cmp.Diff to ignore all unexported (private) variables
func IgnoreAllUnexported(i interface{}) cmp.Option {
	return cmpopts.IgnoreUnexported(findAllStructs(i)...)
}

// IgnoreFields is a wrapper around the cmpopts.IgnoreFields
func IgnoreFields(f ...string) func(interface{}) cmp.Option {
	return func(i interface{}) cmp.Option {
		return cmpopts.IgnoreFields(i, f...)
	}
}

// IgnoreTypes is a wrapper around the cmpopts.IgnoreTypes
// it allows ignore the type of the values passed in
// int32(0), int(0), string(0), time.Duration(0), etc
func IgnoreTypes(types ...interface{}) func(interface{}) cmp.Option {
	return func(_ interface{}) cmp.Option {
		return cmpopts.IgnoreTypes(types...)
	}
}

// ApproxTime is a wrapper around the cmpopts.EquateApproxTime
// it will consider time.Time values equal if there difference is
// less than the defined duration
func ApproxTime(d time.Duration) func(interface{}) cmp.Option {
	return func(_ interface{}) cmp.Option {
		return cmpopts.EquateApproxTime(d)
	}
}

/*
func IgnoreInterfaces(i ...interface{}) func(interface{}) cmp.Option {
	return func(i interface{}) cmp.Option {
		return cmpopts.IgnoreInterfaces(i)
	}
}
*/

// EquateEmpty is a wrapper around cmpopts.EquateEmpty
// it determines all maps and slices with a length of zero to be equal,
// regardless of whether they are nil
func EquateEmpty(i interface{}) cmp.Option {
	return cmpopts.EquateEmpty()
}

func findAllStructs(i interface{}) []interface{} {
	structs := make([]interface{}, 0)
	t := reflect.TypeOf(i)
	// skip invalid types
	if t == nil {
		return structs
	}
	// add struct and pointers to struct
	switch t.Kind() {
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return structs
		}
		if reflect.ValueOf(i).IsNil() {
			return structs
		}
		i = reflect.ValueOf(i).Elem().Interface()
		fallthrough
	case reflect.Struct:
		structs = append(structs, i)

		rStruct := reflect.ValueOf(i)

		// look through all fields of a struct for embedded structs
		for index := 0; index < rStruct.NumField(); index++ {
			v := rStruct.Field(index)
			if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
				// to support unexported (private) fields we need to create a copy
				// of the field and then dereference the pointer to that struct
				i = reflect.New(v.Elem().Type()).Elem().Interface()
				structs = append(structs, findAllStructs(i)...)
				continue
			}
			if !v.CanInterface() {
				// if the field is unexported (private) we wouldn't be able
				// get the interface{} so instead create a copy of that field
				v = reflect.New(v.Type()).Elem()
			}

			structs = append(structs, findAllStructs(v.Interface())...)
		}
	case reflect.Map:
		// since it is possible that we have an empty map
		// create a copy of the map's value Type and check if its a struct
		v := reflect.New(reflect.TypeOf(i).Elem()).Elem()
		structs = append(structs, findAllStructs(v.Interface())...)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		s := reflect.ValueOf(i)
		for i := 0; i < s.Len(); i++ {
			v := s.Index(i)
			structs = append(structs, findAllStructs(v.Interface())...)
		}
	default:
		return structs
	}

	return structs
}

// CmpFuncs tries to determine if x is the same function as y.
func CmpFuncs(x, y interface{}) (b bool, s string) {
	if x == nil || y == nil {
		if x == y {
			return true, ""
		}
		return false, fmt.Sprintf("%v != %v", x, y)
	}

	valX := reflect.ValueOf(x)
	valY := reflect.ValueOf(y)

	if valX.Kind() != reflect.Func || valY.Kind() != reflect.Func {
		return false, fmt.Sprintf("can only compare functions x=%v(%v) y=%v(%v) ", valX.Type(), x, valY.Type(), y)
	}

	if valY.Pointer() == valX.Pointer() {
		return true, ""
	}
	return false, fmt.Sprintf("funcs not equal 0x%x != 0x%x", valY.Pointer(), valX.Pointer())
}

type differ interface {
	String() string
}

// message is a differ for display a custom message
type message string

func (m message) String() string { return string(m) }
func newMessagef(s string, args ...interface{}) message {
	return message(fmt.Sprintf(s, args...))
}

//collection is a differ used for slices to show what items match and which don't
type collection struct {
	found   []interface{}
	missing []interface{}
}

func (c *collection) String() (s string) {
	s = " ∈"
	for _, v := range c.found {
		s += fmt.Sprintf(" %v,", v)
	}
	s = strings.TrimRight(s, " ∈,")
	s += "\n -"
	for _, v := range c.missing {
		s += fmt.Sprintf(" %v,", v)
	}
	return strings.Trim(s, ",\n")
}

type diff struct {
	x   interface{}
	y   interface{}
	msg string
}

func newDiff(x, y interface{}) *diff {
	return &diff{
		x:   x,
		y:   y,
		msg: fmt.Sprintf(" + %v\n - %v", x, y)}
}

func newDiffMsg(x, y interface{}, s string) *diff {
	return &diff{x, y, s}
}

func (d *diff) String() string {
	return fmt.Sprintf("%T ⊇ %T\n%s", d.x, d.y, d.msg)
}

// mapDiff is a differ for maps
type mapDiff struct {
	values map[interface{}][]string
}

func (d *mapDiff) String() (s string) {
	for key, args := range d.values {
		s += fmt.Sprintf(" [%v]", key)
		if len(args) == 0 {
			s += ": missing key\n"
			continue
		}
		var sub string
		for _, v := range args {
			sub += fmt.Sprintf(" %v", v)
		}
		s += ":" + strings.Replace(sub, "\n", "\n    ", -1) + "\n"
	}

	return strings.TrimRight(s, "\n")
}

func (d *mapDiff) diffOrNil() differ {
	if len(d.values) > 0 {
		return d
	}
	return nil
}
