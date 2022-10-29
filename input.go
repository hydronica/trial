package trial

import (
	"fmt"
	"reflect"
	"strconv"
)

// Input the input value given to the trial test function
type Input struct { // TODO: try type Input interface{}
	value reflect.Value
}

func newInput(i interface{}) Input {
	return Input{value: reflect.ValueOf(i)}
}

// String value of input, panics on on non string value
func (in Input) String() string {
	switch in.value.Kind() {
	case reflect.Struct, reflect.Ptr, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		panic("unsupported string conversion " + in.value.Kind().String())
	default:
		return fmt.Sprintf("%v", in.Interface())
	}
}

// Bool value of input, panics on non bool value
func (in Input) Bool() bool {
	if in.value.Kind() == reflect.String {
		b, err := strconv.ParseBool(in.value.String())
		if err != nil {
			panic("invalid bool " + in.value.Interface().(string))
		}
		return b
	}
	return in.value.Bool()
}

// Int value of input, panics on non int value
func (in Input) Int() int {
	switch in.value.Kind() {
	case reflect.String:
		i, err := strconv.Atoi(in.value.String())
		if err != nil {
			panic("invalid int " + in.value.Interface().(string))
		}
		return i
	}
	return int(in.value.Int())
}

// Uint value of input, panics on non uint value
func (in Input) Uint() uint {
	switch in.value.Kind() {
	case reflect.Int:
		return uint(in.value.Int())
	case reflect.String:
		u, err := strconv.ParseUint(in.value.String(), 10, 64)
		if err != nil {
			panic("invalid uint " + in.value.Interface().(string))
		}
		return uint(u)
	default:
		return uint(in.value.Uint())
	}
}

// Interface returns the current value of input
func (in Input) Interface() interface{} {
	//TODO: check for nil
	if in.value.Kind() == reflect.Invalid {
		return nil
	}
	return in.value.Interface()
}

// Float64 value of input, panics on non float64 value
func (in Input) Float64() float64 {
	switch in.value.Kind() {
	case reflect.Int:
		return float64(in.value.Int())
	case reflect.String:
		f, err := strconv.ParseFloat(in.value.String(), 64)
		if err != nil {
			panic("invalid float64 " + in.value.Interface().(string))
		}
		return f
	default:
		return in.value.Float()
	}
}

// Slice returns the input value of the index of a slice/array. panics if non slice value
func (in Input) Slice(i int) Input {
	// use reflect to access any slice type []int, etc
	v := in.value.Index(i)
	if v.Kind() == reflect.Interface {
		return Input{value: reflect.ValueOf(v.Interface())}
	}
	return Input{value: v}
}

// Map returns the value for the provided key, panics on non map value
func (in Input) Map(key interface{}) Input {
	// use reflection to access any map type map[string]string, etc
	return Input{value: in.value.MapIndex(reflect.ValueOf(key))}
}
