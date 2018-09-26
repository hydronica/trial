# Trial - Prove the Innocents of your code

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/jbsmith7741/trial)
[![Build Status](https://travis-ci.com/jbsmith7741/trial.svg?branch=master)](https://travis-ci.com/jbsmith7741/trial)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbsmith7741/trial)](https://goreportcard.com/report/github.com/jbsmith7741/trial)
[![codecov](https://codecov.io/gh/jbsmith7741/trial/branch/master/graph/badge.svg)](https://codecov.io/gh/jbsmith7741/trial)

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

There is no way in golang to pass a generic function as an interface{} and call the function, so we need to use some functional programming. We need to embedded the function we are testing in another function.

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