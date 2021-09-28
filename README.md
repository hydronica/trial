# Trial - Prove the Innocence of your code

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/hydronica/trial)
![Build Status](https://github.com/hydronica/trial/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbsmith7741/trial)](https://goreportcard.com/report/github.com/hydronica/trial)
[![codecov](https://codecov.io/gh/jbsmith7741/trial/branch/master/graph/badge.svg)](https://codecov.io/gh/hydronica/trial)

Framework to make tests easier to create, maintain and debug.


See [wiki](https://github.com/hydronica/trial/wiki) for tips and guides
## Philosophy

- Tests should be written as [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests) with defined inputs and outputs
- Each case must have a unique description
- Tests should be easy to read and change 
- Test shouldn't take too long to complete
- Each case should be fully isolated
  - doesn't depend on previous cases running
  - order of cases shouldn't matter

## Features 
 - test that a function functions as expected. 
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

 Provide a TestFunc method and test Cases to trial.New and call Test with the *testing.T passed in.

``` go
 trial.New(fn testFunc, cases map[string]trial.Case).Test(t)
```

Alternatively to run as each case as a subtest

``` go
 trial.New(fn testFunc, cases trial.Cases).SubTest(t)
 ```

### Case

- **Input struct** - a convenience structure to handle input more dynamically 
  - Smart type conversion were possible ("12" can be converted to int)
- **Expected interface{}** - the expected output of the method being tested.
  - This is compared with the result from the TestFunc
- **ShouldErr bool** - indicates the method should return an error
- **ExpectedErr error** - verifies the method returns the same error as provided.
  - uses strings.Contains to check
  - also implies that the method should error so setting ShouldErr to true is not required
- **ShouldPanic bool** - indicates the method should panic

### TestFunc

``` go
  TestFunc  func(in Input) (result interface{}, err error)
```

- **in Input** - the arguments to be passed as parameters to the method.
  - convert to the Input type. (String(), Int(), Bool(), Map(), Slice(), Interface())
- **output interface{}** - the result from the test that is compared with Case.Expected
- **err error** - any errors that occur during test, return nil if no errors occur.

### Examples

# Compare Functions
used to compare two values to determine equality and displayed a detailed string describing any differences.

``` go
func(actual, expected interface{}) (equal bool, differences string)
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
