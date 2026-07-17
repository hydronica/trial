package trial

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Contains determines if y is a subset of x.
// x is a string -> y is a string that is equal to or a subset of x (string.Contains)
// x is a slice or array -> y is contained in x
// x is a map -> y is a map and is contained in x
func Contains(x, y any) (bool, string) {
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

// const (
// 	SubStrings = iota
// 	SubSlices
// 	SubMaps
// )

// ContainsOpt allow configurable options to the contains method
// 1. Check for sub-strings ("abc" -> "abcdefg")
// 2. Check for sub-slices (["a"] -> ["a","b","c"])
// 3. Check for sub-maps
// 4. use regex match as a sub-string check
/*
func ContainsOpt(o any) CompareFunc {
	return Contains
}
*/

func contains(x, y any) differ {
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
				arrI := make([]any, len(arr))
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
	case reflect.Array, reflect.Slice:
		if valY.Kind() == reflect.Slice || valY.Kind() == reflect.Array {
			child := make([]any, valY.Len())
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
	d := &mapDiff{values: make(map[any][]string, 0)}
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

func isInSlice(parent reflect.Value, child ...any) differ {
	c := &collection{
		found:   make([]any, 0),
		missing: make([]any, 0),
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
func Equal(actual, expected any) (bool, string) {
	fn := EqualOpt(AllowAllUnexported, EquateEmpty)
	return fn(actual, expected)
}

// JSONEqual compares actual and expected semantically as JSON.
// Strings and []byte are unmarshaled; other values are marshaled then unmarshaled
// so both sides share the same JSON DOM shape (map, slice, float64 numbers).
// Use with embedded fixture strings or struct outputs that differ only in key order or formatting.
func JSONEqual(actual, expected any) (bool, string) {
	actualNorm, err := normalizeToJSON(actual)
	if err != nil {
		return false, fmt.Sprintf("actual: invalid JSON: %v", err)
	}
	expectedNorm, err := normalizeToJSON(expected)
	if err != nil {
		return false, fmt.Sprintf("expected: invalid JSON: %v", err)
	}
	r := cmp.Diff(actualNorm, expectedNorm)
	return r == "", r
}

func normalizeToJSON(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch x := v.(type) {
	case string:
		var out any
		if err := json.Unmarshal([]byte(x), &out); err != nil {
			return nil, err
		}
		return out, nil
	case []byte:
		var out any
		if err := json.Unmarshal(x, &out); err != nil {
			return nil, err
		}
		return out, nil
	case json.RawMessage:
		var out any
		if err := json.Unmarshal(x, &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var out any
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
}

// EqualOpt allow easy customization of the cmp.Equal method.
// see below for a list of supported options
func EqualOpt(optFns ...func(i any) cmp.Option) func(actual, expected any) (bool, string) {
	return func(actual, expected any) (bool, string) {
		opts := make([]cmp.Option, 0)
		for _, fn := range optFns {
			opts = append(opts, fn(actual))
		}

		r := cmp.Diff(actual, expected, opts...)
		return r == "", r
	}
}

// AllowAllUnexported sets cmp.Diff to allow all unexported (private) variables
func AllowAllUnexported(i any) cmp.Option {
	return cmp.AllowUnexported(findAllStructs(i, nil)...)
}

// IgnoreAllUnexported sets cmp.Diff to ignore all unexported (private) variables
func IgnoreAllUnexported(i any) cmp.Option {
	return cmpopts.IgnoreUnexported(findAllStructs(i, nil)...)
}

// IgnoreFields is a wrapper around the cmpopts.IgnoreFields
// syntax: IgnoreFields(package.struct.Field)
func IgnoreFields(f ...string) func(any) cmp.Option {
	return func(i any) cmp.Option {
		t := reflect.TypeOf(i)
		if t.Kind() == reflect.Ptr { // dereference pointers
			i = reflect.New(t.Elem()).Elem().Interface()
		}
		// get the type of element of a slice/array
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			i = reflect.New(t.Elem()).Elem().Interface()
		}
		return cmpopts.IgnoreFields(i, f...)
	}
}

// IgnoreFieldsOf ignores specific fields on a given struct type.
// Use this when ignoring fields in embedded structs where IgnoreFields
// cannot infer the correct type.
//
//	trial.EqualOpt(trial.IgnoreFieldsOf(Metadata{}, "CreatedAt", "UpdatedAt"))
func IgnoreFieldsOf(structType any, fields ...string) func(any) cmp.Option {
	return func(_ any) cmp.Option {
		return cmpopts.IgnoreFields(structType, fields...)
	}
}

// IgnoreTypes is a wrapper around the cmpopts.IgnoreTypes
// it allows ignore the type of the values passed in
// int32(0), int(0), string(0), time.Duration(0), etc
func IgnoreTypes(types ...any) func(any) cmp.Option {
	return func(_ any) cmp.Option {
		return cmpopts.IgnoreTypes(types...)
	}
}

// ApproxTime is a wrapper around the cmpopts.EquateApproxTime
// it will consider time.Time values equal if there difference is
// less than the defined duration
func ApproxTime(d time.Duration) func(any) cmp.Option {
	return func(_ any) cmp.Option {
		return cmpopts.EquateApproxTime(d)
	}
}

/*
func IgnoreInterfaces(i ...any) func(any) cmp.Option {
	return func(i any) cmp.Option {
		return cmpopts.IgnoreInterfaces(i)
	}
}
*/

// EquateEmpty is a wrapper around cmpopts.EquateEmpty
// it determines all maps and slices with a length of zero to be equal,
// regardless of whether they are nil
func EquateEmpty(i any) cmp.Option {
	return cmpopts.EquateEmpty()
}

type structMap map[string]any // [Name]struct

func (s structMap) Add(vals ...any) {
	for _, v := range vals {
		s[reflect.TypeOf(v).Name()] = v
	}
}

func (s structMap) List() []any {
	vals := make([]any, len(s))
	i := 0
	for _, v := range s {
		vals[i] = v
		i++
	}
	sort.Slice(vals, func(i, j int) bool {
		return reflect.TypeOf(vals[i]).Name() < reflect.TypeOf(vals[j]).Name()

	})
	return vals
}

func findAllStructs(i any, structs structMap) []any {
	if structs == nil {
		structs = make(structMap)
	}
	for _, v := range structs {
		if reflect.TypeOf(v) == reflect.TypeOf(i) {
			return []any{}
		}
	}

	t := reflect.TypeOf(i)
	// skip invalid types
	if t == nil {
		return []any{}
	}
	// add struct and pointers to struct
	switch t.Kind() {
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return []any{}
		}
		if reflect.ValueOf(i).IsNil() {
			return []any{}
		}
		i = reflect.ValueOf(i).Elem().Interface()
		fallthrough
	case reflect.Struct:
		structs.Add(i)

		rStruct := reflect.ValueOf(i)

		// look through all fields of a struct for embedded structs
		for index := 0; index < rStruct.NumField(); index++ {
			v := rStruct.Field(index)
			if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
				// to support unexported (private) fields we need to create a copy
				// of the field and then dereference the pointer to that struct
				i = reflect.New(v.Elem().Type()).Elem().Interface()
				structs.Add(findAllStructs(i, structs)...)
				continue
			}
			if !v.CanInterface() {
				// if the field is unexported (private) we wouldn't be able
				// to get the any so instead create a copy of that field
				v = reflect.New(v.Type()).Elem()
			}
			structs.Add(findAllStructs(v.Interface(), structs)...)
		}
	case reflect.Map:
		// since it is possible that we have an empty map
		// create a copy of the map's value Type and check if it's a struct
		valType := reflect.TypeOf(i).Elem()
		// if the map's value Type is a pointer, deference it to avoid pointers to pointers.
		if valType.Kind() == reflect.Pointer {
			valType = valType.Elem()
		}
		v := reflect.New(valType).Elem()
		structs.Add(findAllStructs(v.Interface(), structs)...)
	case reflect.Array, reflect.Slice:

		arrType := reflect.TypeOf(i).Elem()
		if arrType.Kind() == reflect.Interface {
			s := reflect.ValueOf(i)
			for j := 0; j < s.Len(); j++ {
				v := s.Index(j)
				structs.Add(findAllStructs(v.Interface(), structs)...)
			}
		} else {
			if arrType.Kind() == reflect.Pointer {
				arrType = arrType.Elem()
			}
			v := reflect.New(arrType).Elem()
			structs.Add(findAllStructs(v.Interface(), structs)...)
		}
	default:
		return []any{}
	}

	return structs.List()
}

// CmpFuncs tries to determine if x is the same function as y.
func CmpFuncs(x, y any) (b bool, s string) {
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
func newMessagef(s string, args ...any) message {
	return message(fmt.Sprintf(s, args...))
}

// collection is a differ used for slices to show what items match and which don't
type collection struct {
	found   []any
	missing []any
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
	x   any
	y   any
	msg string
}

func newDiff(x, y any) *diff {
	return &diff{
		x:   x,
		y:   y,
		msg: fmt.Sprintf(" + %v\n - %v", x, y)}
}

func newDiffMsg(x, y any, s string) *diff {
	return &diff{x, y, s}
}

func (d *diff) String() string {
	return fmt.Sprintf("%T ⊇ %T\n%s", d.x, d.y, d.msg)
}

// mapDiff is a differ for maps
type mapDiff struct {
	values map[any][]string
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
