# Trial - Prove the Innocence of your code

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/hydronica/trial)
![Build Status](https://github.com/hydronica/trial/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbsmith7741/trial)](https://goreportcard.com/report/github.com/hydronica/trial)
[![codecov](https://codecov.io/gh/jbsmith7741/trial/branch/master/graph/badge.svg)](https://codecov.io/gh/hydronica/trial)

Go testing framework to make tests easier to create, maintain and debug.


See [wiki](https://github.com/hydronica/trial/wiki) for tips and guides

| [Examples](https://github.com/hydronica/trial/wiki) | [Compare Functions](https://github.com/hydronica/trial/wiki/Comparers) | [Helper Functions](https://github.com/hydronica/trial/wiki/Helpers) |

## Philosophy

- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- *Descriptive*
  - tests should describe the expected behavior of a function
  - case needs unique description
  - easy to read and change 
- *Fast* - the test suite should run in under 5 minutes
- *Independent*
  - Each case should be fully isolated
  - doesn't depend on previous cases running
  - order of cases shouldn't matter
- *Repeatable*: non-flaky, runs regardless of the time of year
- Test for observable behavior

## Features 
 - function verification. 
    - Check results for exact matches including private values (default behavior) 
    - Check results for values contained in others (see Contains function) 
    - Allows for custom compare functions 
  - Catch and test for panics 
    - each test is self isolated so a panic won't stop other cases from running 
    - check for expected panic cases with `ShouldPanic`
  - Test error cases 
    - Check that a function returns an error: `ShouldErr`
    - Check that an error strings contains expected string: `ExpectedErr`
    - Check that an error is of expected type: `ExpectedErr: ErrType(err)`
  - Fail tests that take too long to complete
    - `trial.New(fn,cases).Timeout(time.Second)`



## Getting Starting

``` go 
go get github.com/hydronica/trial
```

Each test has 3 parts to it. A **function** to test against, a set of test **cases** to validate with and **trial** that sets up the table driven tests

### **Test Function**
``` go 
testFunc[In any, Out any] func(in In) (result Out, err error)
```
a generic function that has a single input and returns a result and an error. Wrap your function to test more complex functions or methods.

*Example* 
``` go 
  fn := func(i int) (string, error) {
    return strconv.ItoA(i), nil 
  }

```


### **Cases**
a collection (map) of test cases that have a unique title and defined *input* to be passed to the test function. The expected behavior is described by provided an *output*, *ExpectedErr* or specified a generic error with *ShouldErr*. *ShouldPanic* can be used in the rare cases of function that need to panic. 
``` go 
type Case[In any, Out any] struct {
	Input    In
	Expected Out


	ShouldErr   bool  // is an error expected
	ExpectedErr error // the error that was expected (nil is no error expected)
	ShouldPanic bool  // is a panic expected
}
```
Each 

- **Input** *generic* - a convenience structure to handle input more dynamically 
  - Smart type conversion were possible ("12" can be converted to int)
- **Expected** *generic* - the expected output of the method being tested.
  - This is compared with the result from the TestFunc
- **ShouldErr** *bool* - indicates the function should return an error
- **ExpectedErr** *error* - verifies the function error string matches the result
  - uses strings.Contains to check
  - also implies that the method should error so setting ShouldErr to true is not required
  - use *ErrType* to test that the error is the same type as expected. 
- **ShouldPanic** *bool* - indicates the method should panic


### Trial Setup

Run the test cases either within a single test function or as subtests. The *input* and *output* values must match between the test function and cases. 

``` go
trial.New(fn,cases).Test(t)
// or 
trial.New(fn,cases).SubTest(t)
```

By default trial uses a strict matching values and uses cmp.Equal to compare values. *Compare* Functions can be customized to ignore certain fields or are contained withing maps, slices or strings. See **Compare Functions** for more details. A timeout can be added onto the trial builder with `.Timeout(time.Second)` 

### Getting Started Template 
``` go  
fn := func(in any) (any, error) {
    // TODO: setup test case, call routine and return result
    return nil, nil
}
cases := trail.Cases{
    "default": {
        Input:    123,
        Expected: 1,
    },
}
trial.New(fn,cases).Test(t)
```

For more examples see the [wiki](https://github.com/hydronica/trial/wiki)

# Compare Functions
used to compare two values to determine equality and displayed a detailed string describing any differences.

``` go
func(actual, expected any) (equal bool, differences string)
```

override the default

``` go
trial.New(fn, cases).Comparer(myComparer).Test(t)
```

## Equal
The default comparer used, it is a wrapping for cmp.Equal with the AllowUnexported option set for all structs. This causes all fields (public and private) in a struct to be compared. (see https://github.com/google/go-cmp)

### EqualOpt

Customize the use of cmp.Equal with the following supported options: 
  - `AllowAllUnexported` - **[default: Equal]** compare all unexported (private) variables within a struct. This is useful when testing a struct inside its own package. 
  - `IgnoreAllUnexported` - ignore all unexported (private) variables within a struct. This is useful when dealing with a struct outside the project. 
  - `IgnoreFields(fields ...string)` - define a list of variables to exclude for the comparer, the field name are case sensitize and can be dot-delimited ("Field", "Parent.child")
  - `EquateEmpty`- **[default: Equal]** a nil map or slice is equal to an empty one (len is zero)
  - `IgnoreTypes(values ...interface{})` - ignore all types of the values passed in. Ex: IgnoreTypes(int64(0), float32(0.0)) ignore int64 and float32
  - `ApproxTime(d time.Duration)` - approximates time values to to the nearest duration. 

  
## Contains ⊇

Checks if the expected value is *contained* in the actual value. The symbol ⊇ is used to donate a subset. ∈ is used to show that a value exists in a slice. Contains checks the following relationships

- **string ⊇ string**
  - is the expected string contained in the actual string (strings.Contains)
- **string ⊇ []string**
  - are the expected substrings contained in the actual string
- **[]interface{} ⊇ interface{}**
  - is the expected value found in the slice or array
- **[]interface{} ⊇ []interface{}**
  - is the expected slice a subset of the actual slice. all values in expected exist and are contained in actual.
- **map[key]interface{} ⊇ map[key]interface{}**
  - is the expected map a subset of the actual map. all keys in expected are in actual and all values under that key are contained in actual
