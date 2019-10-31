# Trial - Prove the Innocents of your code

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/hydronica/trial)
[![Build Status](https://travis-ci.com/jbsmith7741/trial.svg?branch=master)](https://travis-ci.com/hydronica/trial)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbsmith7741/trial)](https://goreportcard.com/report/github.com/hydronica/trial)
[![codecov](https://codecov.io/gh/jbsmith7741/trial/branch/master/graph/badge.svg)](https://codecov.io/gh/hydronica/trial)

Framework to make tests easier to create, maintain and debug.

## Philosophy

- Tests should be written as [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests) with defined inputs and outputs
- Each case must have a unique description
- Each case should be fully isolated
  - doesn't depend on previous cases running
  - order of cases shouldn't matter

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

- **Input interface{}** - the input to the method being tested.
  - If the method has multiple parameters either embed the values in a struct or use trial.Args(args ...interface{}) to pass in multiple parameters
- **Expected interface{}** - the expected output of the method being tested.
  - This is compared with the result from the TestFunc
- **ShouldErr bool** - indicates the method should return an error
- **ExpectedErr error** - verifies the method returns the same error as provided.
  - uses strings.Contains to check
  - also implies that the method should error so setting ShouldErr to true is not required
- **ShouldPanic bool** - indicates the method should panic

### TestFunc

``` go
  TestFunc  func(args ...interface{}) (result interface{}, err error)
```

- **args []interface{}** - the arguments to be passed as parameters to the method.
  - case to the expected type, eg args[0].(string), args[1].(int)
- **output interface{}** - the result from the test that is compared with Case.Expected
- **err error** - any errors that occur during test, return nil if no errors occur.

### Examples

#### Test a simple add method

``` go
func Add(i1, i2 int) int {
  return i1 + i2
}

func TestAdd(t *testing.T) {
  testFn := func(args ...interface{}) (interface{}, error) {
    return Add(args[0].(int), args[1].(int), nil
  }
  cases := trial.Cases{
    "Add two numbers":{
      Input: trial.Args(1,2),
      Expected: 3,
  }
  trial.New(fn, cases).Test(t)
  }

  // Output: PASS: "Add two numbers"
```

#### Test string to int conversion

``` go
func TestStrconv_Itoa(t *testing.T)
testFn := func(args ...interface{}) (interface{}, error) {
    return strconv.Itoa(args[0].(int))
}
cases :=trial.Cases{
  "valid int":{
    Input: "12",
    Expected: 12,
  },
  "invalid int": {
    Input: "1abe",
    ShouldErr: true,
  },
}
trial.New(fn, cases).Test(t)
}

// Output: PASS: "valid int"
// PASS: "invalid int"
```

#### Test divide method

``` go
func Divide(i1, i2 int) int {
  return i1/i2
}
func TestDivide(t *testing.T) {
  fn := func(args ...interface) (interface{}, error) {
    return Divide(args[0].(int), args[1].(int)), nil
  }
  cases := trial.Cases{
    "1/1":{
      Input: trial.Args(1,1),
      Expected: 1,
    },
    "6/2": {
      Input: trial.Args(6,2),
      Expected: 1,
    },
    "divide by zero": {
      Input: trial.Args(1,0),
      ShouldPanic: true,
    }
  }
  trial.New(fn, cases).Test(t)
}
// Output: PASS: "1/1"
// FAIL: "6/2"
// PASS: "divide by zero"
```

## Compare Functions
used to compare two values to determine if they are considered equal and displayed a detailed string describing the differences found.

``` go
func(actual, expected interface{}) (equal bool, differences string)
```

override the default

``` go
trial.New(fn, cases).EqualFn(myComparer).Test(t)
```

### Equal
This is the default comparer used, it is a wrapping for cmp.Equal with the AllowUnexported option set for all structs. This causes all fields (public and private) in a struct to be compared. (see https://github.com/google/go-cmp)

### Contains ⊇

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

## Helper Functions
The helper functions are convince methods for either ignoring errors on test setup or for capturing output for testing.

### Output Capturing
  Capture output written to log, stdout or stderr.
  Call *ReadAll* to get captured data as a single string.
  Call *ReadLines* to get captured data as a []string split by newline. Calling either method closes and reset the output redirection.

#### CaptureLog

``` go
  c := CaptureLog()
  // logic that writes to logs
  log.Print("hello")
  c.ReadAll() // -> returns hello
```

Note: log is reset to write to stderr

#### CaptureStdErr

``` go
  c := CaptureStdErr()
  // write to stderr
  fmt.Fprint(os.stderr, "hello\n")
  fmt.Fprint(os.stderr, "world")
  c.ReadLines() // []string{"hello","world"}
```

#### CaptureStdOut

``` go
  c := CaptureStdOut()
  // write to stdout
  fmt.Println("hello")
  fmt.Print("world")
  c.ReadLines() // []string{"hello","world"}
```

### Time Parsing

convenience functions for getting a time value to test, methods panic instead of error

- **TimeHour(s string)** - uses format "2006-01-02T15"
- **TimeDay(s string)** - uses format "2006-01-02"
- **Times(layout string, values ...string)**
- **TimeP(layout string, s string)** returns a *time.Time

### Pointer init

convenience functions for initializing a pointer to a basic type

- int pointer
  - IntP, Int8P, Int16P, Int32P, Int64P
- uint pointer
  - UintP, Uint8P, Uint16P, Uint32P, Uint64P
- bool pointer - BoolP
- float pointer
  - Float32P, Float64P
- string pointer - StringP