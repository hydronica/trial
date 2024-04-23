package trial

import (
	"time"
)

// Args converts any number of parameters to an interface.
// generally used with Case's Input for multiple params
func Args(args ...interface{}) Input {
	if len(args) == 1 {
		return newInput(args[0])
	}
	return newInput(args)
}

type primitives interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		string | bool | float32 | float64
}

func Pointer[T primitives](v T) *T {
	return &v
}

// Time is a panic wrapper for the time.Parse method
// it returns a time.Time for the given layout and value
func Time(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// Hour is a helper method for parsing times with hour precision.
// format 2006-01-02T15
func Hour(value string) time.Time {
	return Time("2006-01-02T15", value)
}

// Day is a helper method for parsing times with day precision.
// format 2006-01-02
func Day(value string) time.Time {
	// time.DateOnly was added in go1.20
	return Time("2006-01-02", value)
}

// Times is a panic wrapper for the time.Parse method
// it returns a time.Time slice for the given layout and values
func Times(layout string, values ...string) []time.Time {
	times := make([]time.Time, len(values))
	for i, v := range values {
		times[i] = Time(layout, v)
	}
	return times
}

// TimeP return a pointers to a time.Time for the given layout and value.
// it panics on error
func TimeP(layout, value string) *time.Time {
	t := Time(layout, value)
	return &t
}

// ***** DEPRECATED functions

// Deprecated: TimeHour is a helper method for parsing times with hour precision.
// format 2006-01-02T15
func TimeHour(value string) time.Time {
	return Time("2006-01-02T15", value)
}

// Deprecated: TimeDay is a helper method for parsing times with day precision.
// format 2006-01-02
func TimeDay(value string) time.Time {
	return Time("2006-01-02", value)
}

// Deprecated: IntP returns a pointer to a defined int
func IntP(i int) *int {
	return &i
}

// Deprecated: Int8P returns a pointer to a defined int8
func Int8P(i int8) *int8 {
	return &i
}

// Deprecated:  Int16P returns a pointer to a defined int16
func Int16P(i int16) *int16 {
	return &i
}

// Deprecated:  Int32P returns a pointer to a defined int32
func Int32P(i int32) *int32 {
	return &i
}

// Deprecated:  Int64P returns a pointer to a defined int64
func Int64P(i int64) *int64 {
	return &i
}

// Deprecated:  UintP returns a pointer to a defined uint
func UintP(i uint) *uint {
	return &i
}

// Deprecated:  Uint8P returns a pointer to a defined uint8
func Uint8P(i uint8) *uint8 {
	return &i
}

// Deprecated:  Uint16P returns a pointer to a defined uint16
func Uint16P(i uint16) *uint16 {
	return &i
}

// Deprecated:  Uint32P returns a pointer to a defined uint32
func Uint32P(i uint32) *uint32 {
	return &i
}

// Deprecated:  Uint64P returns a pointer to a defined uint64
func Uint64P(i uint64) *uint64 {
	return &i
}

// Deprecated:  Float32P returns a pointer to a defined float32
func Float32P(f float32) *float32 {
	return &f
}

// Deprecated:  Float64P returns a pointer to a defined float64
func Float64P(f float64) *float64 {
	return &f
}

// Deprecated:  BoolP returns a pointer to a defined bool
func BoolP(b bool) *bool {
	return &b
}

// Deprecated:  StringP returns a pointer to a defined string
func StringP(s string) *string {
	return &s
}
