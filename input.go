package trial

import "reflect"

type TestFunc2 func(in Input) (result interface{}, err error)

// Input is the input value given to the test struct
type Input struct {
	value interface{}
}

func (in Input) String() string {
	return in.value.(string)
}

func (in Input) Bool() bool {
	return in.value.(bool)
}

func (in Input) Int() int {
	return in.value.(int)
}

func (in Input) Uint() uint {
	return in.value.(uint)
}

func (in Input) Interface() interface{} {
	return in.value
}

func (in Input) Float64() float64 {
	return in.value.(float64)
}

func (in Input) Slice(i int) Input {
	// use reflect to access any slice type []int, etc
	v := reflect.ValueOf(in.value)
	return Input{value: v.Index(i).Interface()}
}

func (in Input) Map(key interface{}) Input {
	// use reflection to access any map type map[string]string, etc
	v := reflect.ValueOf(in.value)
	return Input{value: v.MapIndex(reflect.ValueOf(key)).Interface()}
}
