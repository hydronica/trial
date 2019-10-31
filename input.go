package trial

import "reflect"

// Input the input value given to the trial test function
type Input struct {
	value reflect.Value
}

func newInput(i interface{}) Input {
	return Input{value: reflect.ValueOf(i)}
}

// String value of input, panics on on non string value
func (in Input) String() string {
	// todo: should all types be cast to their string value?
	return in.value.Interface().(string)
}

// Bool value of input, panics on non bool value
func (in Input) Bool() bool {
	return in.value.Bool()
}

// Int value of input, panics on non int value
func (in Input) Int() int {
	return int(in.value.Int())
}

// Uint value of input, panics on non uint value
func (in Input) Uint() uint {
	switch in.value.Kind() {
	case reflect.Int:
		return uint(in.value.Int())
	default:
		return uint(in.value.Uint())
	}
}

// Interface returns the current value of input
func (in Input) Interface() interface{} {
	return in.value.Interface()
}

// Float64 value of input, panics on non float64 value
func (in Input) Float64() float64 {
	switch in.value.Kind() {
	case reflect.Int:
		return float64(in.value.Int())
	default:
		return in.value.Float()
	}
}

// Slice returns the input value of the index of a slice/array. panics if non slice value
func (in Input) Slice(i int) Input {
	// use reflect to access any slice type []int, etc
	return Input{value: in.value.Index(i)}
}

// Map returns the value for the provided key, panics on non map value
func (in Input) Map(key interface{}) Input {
	// use reflection to access any map type map[string]string, etc
	return Input{value: in.value.MapIndex(reflect.ValueOf(key))}
}
