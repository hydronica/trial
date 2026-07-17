# Trial - Prove the Innocence of your code

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/hydronica/trial)
![Build Status](https://github.com/hydronica/trial/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/hydronica/trial)](https://goreportcard.com/report/github.com/hydronica/trial)
[![codecov](https://codecov.io/gh/hydronica/trial/branch/master/graph/badge.svg)](https://codecov.io/gh/hydronica/trial)

Go testing framework to make tests easier to create, maintain and debug.

## Documentation

| [Examples & Patterns](docs/examples.md) | [API Reference](docs/api.md) | [Comparers](docs/comparers.md) | [Helpers](docs/helpers.md) |
|-|-|-|-|

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
 - Function verification
    - Check results for exact matches including private values (default behavior) 
    - Check results for values contained in others (see Contains function) 
    - Allows for custom compare functions 
  - Catch and test for panics 
    - each test is self isolated so a panic won't stop other cases from running 
    - check for expected panic cases with `ShouldPanic`
  - Test error cases 
    - Check that a function returns an error: `ShouldErr` or `ExpectedErr: trial.Error()`
    - Match error message: `trial.Error().Exact()`, `.Contains()`, or `.Regex()`
    - Legacy substring match: `ExpectedErr: errors.New("fragment")`
    - Check error type: `ExpectedErr: trial.Error().IsType(err)`
  - Fail tests that take too long to complete
    - `trial.New(fn,cases).Timeout(time.Second)`
  - Run subtests in parallel for faster execution
    - `trial.New(fn,cases).Parallel().SubTest(t)`

## Getting Started

```go 
go get github.com/hydronica/trial
```

Each test has 3 parts to it. A **function** to test against, a set of test **cases** to validate with and **trial** that sets up the table driven tests

### Test Function

```go 
testFunc[In any, Out any] func(in In) (result Out, err error)
```

A generic function that has a single input and returns a result and an error. Wrap your function to test more complex functions or methods.

*Example* 
```go 
fn := func(i int) (string, error) {
    return strconv.Itoa(i), nil 
}
```

### Cases

A collection (map) of test cases that have a unique title and defined *input* to be passed to the test function. The expected behavior is described by providing an *output*, setting an *ExpectedErr* or saying it *ShouldErr*. *ShouldPanic* can be used to test when function panic condition.  

```go 
type Case[In any, Out any] struct {
    Input    In
    Expected Out

    ShouldErr   bool  // is an error expected
    ExpectedErr error // the error that was expected (nil is no error expected)
    ShouldPanic bool  // is a panic expected
}
```

Each case field is described below:

- **Input** *generic* - a convenience structure to handle input more dynamically 
- **Expected** *generic* - the expected output of the method being tested.
  - This is compared with the result from the TestFunc
- **ShouldErr** *bool* - indicates the function should return an error
- **ExpectedErr** *error* - verifies the function returns an expected error
  - `trial.Error()` — any error
  - `trial.Error().Exact("msg")` — exact message; `.Contains("msg")` for substring; `.Regex(pat)` for pattern
  - `trial.Error().IsType(err)` — type match; chain `.Contains()` / `.Regex()` for message checks
  - `errors.New("fragment")` — legacy substring match via `strings.Contains`
  - also implies that the method should error so setting ShouldErr to true is not required
- **ShouldPanic** *bool* - indicates the method should panic

### Trial Setup

Run the test cases either within a single test function or as subtests. The *input* and *output* values must match between the test function and cases. 

```go
trial.New(fn,cases).Test(t)
// or 
trial.New(fn,cases).SubTest(t)
```

By default trial uses strict matching values and uses cmp.Equal to compare values. *Compare* functions can be customized to ignore certain fields or are contained within maps, slices or strings. See [Comparers](docs/comparers.md) for more details. A timeout can be added onto the trial builder with `.Timeout(time.Second)`. For faster test execution, subtests can run in parallel with `.Parallel()` (test cases must be thread-safe) 

### Getting Started Template 

```go  
fn := func(in any) (any, error) {
    // TODO: setup test case, call routine and return result
    return nil, nil
}
cases := trial.Cases[any, any]{
    "default": {
        Input:    123,
        Expected: 1,
    },
}
trial.New(fn,cases).Test(t)
```

## Compare Functions

Used to compare two values to determine equality and display a detailed string describing any differences.

```go
func(actual, expected any) (equal bool, differences string)
```

Override the default:

```go
trial.New(fn, cases).Comparer(myComparer).Test(t)
```

### Equal

The default comparer used, it is a wrapper for cmp.Equal with the AllowUnexported option set for all structs. This causes all fields (public and private) in a struct to be compared. (see https://github.com/google/go-cmp)

### EqualOpt

Customize the use of cmp.Equal with the following supported options: 
  - `AllowAllUnexported` - **[default: Equal]** compare all unexported (private) variables within a struct. This is useful when testing a struct inside its own package. 
  - `IgnoreAllUnexported` - ignore all unexported (private) variables within a struct. This is useful when dealing with a struct outside the project. 
  - `IgnoreFields(fields ...string)` - define a list of variables to exclude for the comparer, the field names are case sensitive and can be dot-delimited ("Field", "Parent.child")
  - `EquateEmpty`- **[default: Equal]** a nil map or slice is equal to an empty one (len is zero)
  - `IgnoreTypes(values ...interface{})` - ignore all types of the values passed in. Ex: IgnoreTypes(int64(0), float32(0.0)) ignore int64 and float32
  - `ApproxTime(d time.Duration)` - approximates time values to the nearest duration. 

```go
// example struct that is compared to expected
type Example struct {
    Field1     int
    Field2     int
    ignoreMe   string     // ignored unexported
    Timestamp  time.Time  // ignored
    LastUpdate time.Time  // ignored
}

trial.New(fn, cases).Comparer(
    trial.EqualOpt(
        trial.IgnoreAllUnexported,
        trial.IgnoreFields("Timestamp", "LastUpdate")),
).SubTest(t)
```
  
### Contains ⊇

Checks if the expected value is *contained* in the actual value. The symbol ⊇ is used to denote a subset. ∈ is used to show that a value exists in a slice. Contains checks the following relationships:

- **string ⊇ string** - is the expected string contained in the actual string (strings.Contains)
- **string ⊇ []string** - are the expected substrings contained in the actual string
- **[]interface{} ⊇ interface{}** - is the expected value found in the slice or array
- **[]interface{} ⊇ []interface{}** - is the expected slice a subset of the actual slice. all values in expected exist and are contained in actual.
- **map[key]interface{} ⊇ map[key]interface{}** - is the expected map a subset of the actual map. all keys in expected are in actual and all values under that key are contained in actual
