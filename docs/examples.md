# Trial Examples and Patterns

Practical examples for using the `hydronica/trial` testing framework.

## Table of Contents

- [Principles](#principles)
- [Basic Template](#basic-template)
- [Testing with Input Structs](#testing-with-input-structs)
- [Testing Functions with Multiple Arguments](#testing-functions-with-multiple-arguments)
- [Error Handling](#error-handling)
- [Panic Testing](#panic-testing)
- [Test() vs SubTest()](#test-vs-subtest)
- [Using Helpers](#using-helpers)
- [Custom Comparers](#custom-comparers)
- [Timeout](#timeout)
- [Complete Example](#complete-example)

---

## Principles

1. **Keep it simple** - Use primitives when possible for input and output
2. **Return errors** - Test functions must return `(result, error)`
3. **Self-contained cases** - Each case should not depend on others
4. **Limit scope** - Define custom structs within the test function

---

## Basic Template

```go
func TestBasic(t *testing.T) {
    fn := func(in string) (int, error) {
        return len(in), nil
    }
    cases := trial.Cases[string, int]{
        "empty string": {
            Input:    "",
            Expected: 0,
        },
        "hello": {
            Input:    "hello",
            Expected: 5,
        },
    }
    trial.New(fn, cases).Test(t)
}
```

---

## Testing with Input Structs

When testing functions with multiple parameters, define an input struct:

```go
func Add(i1, i2 int) int {
    return i1 + i2
}

func TestAdd(t *testing.T) {
    type input struct {
        i1 int
        i2 int
    }
    fn := func(in input) (int, error) {
        return Add(in.i1, in.i2), nil
    }
    cases := trial.Cases[input, int]{
        "add two positive numbers": {
            Input:    input{i1: 1, i2: 2},
            Expected: 3,
        },
        "add negative numbers": {
            Input:    input{i1: -5, i2: -3},
            Expected: -8,
        },
    }
    trial.New(fn, cases).Test(t)
}
```

---

## Testing Functions with Multiple Arguments

Alternative: use `trial.Args()` with `trial.Input` for dynamic argument handling:

```go
func TestAddWithArgs(t *testing.T) {
    fn := func(in trial.Input) (int, error) {
        return Add(in.Slice(0).Int(), in.Slice(1).Int()), nil
    }
    cases := trial.Cases[trial.Input, int]{
        "add two numbers": {
            Input:    trial.Args(1, 2),
            Expected: 3,
        },
    }
    trial.New(fn, cases).Test(t)
}
```

---

## Error Handling

### ShouldErr - Expect Any Error

```go
cases := trial.Cases[[]int, int]{
    "divide by zero": {
        Input:     []int{1, 0},
        ShouldErr: true,
    },
}
```

### ExpectedErr - Expect Specific Error Message

Uses `strings.Contains` to check the error message:

```go
cases := trial.Cases[string, output]{
    "invalid character": {
        Input:       `{"Value", "abc"}`,
        ExpectedErr: errors.New("invalid character"),
    },
}
```

### ErrType - Expect Specific Error Type

```go
type ValidationError struct {
    Field string
}

func (e ValidationError) Error() string {
    return "validation failed: " + e.Field
}

cases := trial.Cases[string, string]{
    "validation error": {
        Input:       "",
        ExpectedErr: trial.ErrType(ValidationError{}),
    },
}
```

---

## Panic Testing

Use `ShouldPanic: true` to test functions that should panic. Each case is isolated, so a panic won't stop other cases.

```go
func TestDivideWithPanic(t *testing.T) {
    fn := func(in []int) (int, error) {
        return in[0] / in[1], nil  // panics if in[1] is 0
    }
    cases := trial.Cases[[]int, int]{
        "normal division": {
            Input:    []int{10, 2},
            Expected: 5,
        },
        "divide by zero panics": {
            Input:       []int{1, 0},
            ShouldPanic: true,
        },
    }
    trial.New(fn, cases).Test(t)
}
```

---

## Test() vs SubTest()

### Test()

Runs all cases in a single test function:

```go
trial.New(fn, cases).Test(t)
// PASS: "case 1"
// PASS: "case 2"
// FAIL: "case 3"
```

### SubTest()

Runs each case as a separate Go subtest using `t.Run()`:

```go
trial.New(fn, cases).SubTest(t)
// === RUN   TestFunction/case_1
// --- PASS: TestFunction/case_1
// === RUN   TestFunction/case_2
// --- PASS: TestFunction/case_2
```

**Use `SubTest()` when you need:**
- CI/CD systems that track individual test results
- Run specific cases: `go test -run "TestName/case_name"`
- Better test isolation and reporting

---

## Using Helpers

See [helpers.md](helpers.md) for complete documentation.

### Time Helpers

```go
cases := trial.Cases[time.Time, string]{
    "parse day": {
        Input:    trial.Day("2024-01-15"),
        Expected: "January 15, 2024",
    },
    "parse hour": {
        Input:    trial.Hour("2024-01-15T14"),
        Expected: "January 15, 2024",
    },
    "custom format": {
        Input:    trial.Time(time.RFC3339, "2024-01-15T14:30:00Z"),
        Expected: "January 15, 2024",
    },
}
```

### Pointer Helpers

```go
type Config struct {
    Name    *string
    Count   *int
    Enabled *bool
}

cases := trial.Cases[Config, string]{
    "with all fields": {
        Input: Config{
            Name:    trial.Pointer("app"),
            Count:   trial.Pointer(42),
            Enabled: trial.Pointer(true),
        },
        Expected: "app",
    },
}
```

---

## Custom Comparers

See [comparers.md](comparers.md) for complete documentation.

### Ignoring Fields

```go
type Example struct {
    Field1     int
    Field2     int
    Timestamp  time.Time
    LastUpdate time.Time
}

trial.New(fn, cases).Comparer(
    trial.EqualOpt(
        trial.IgnoreAllUnexported,
        trial.IgnoreFields("Timestamp", "LastUpdate"),
    ),
).SubTest(t)
```

### Using Contains

For substring/subset matching:

```go
fn := func(in string) (string, error) {
    return "Hello, " + in + "! Welcome.", nil
}
cases := trial.Cases[string, string]{
    "greeting contains name": {
        Input:    "World",
        Expected: "World",  // actual contains "World"
    },
}
trial.New(fn, cases).Comparer(trial.Contains).Test(t)
```

### Approximate Time Comparison

```go
trial.New(fn, cases).Comparer(
    trial.EqualOpt(trial.ApproxTime(time.Second)),
).Test(t)
```

---

## Timeout

Fail cases that take too long:

```go
trial.New(fn, cases).Timeout(time.Second).Test(t)
```

---

## Complete Example

Copy-paste-ready test file:

```go
package mypackage

import (
    "errors"
    "testing"
    "github.com/hydronica/trial"
)

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}

func TestDivide(t *testing.T) {
    type input struct {
        a int
        b int
    }
    fn := func(in input) (int, error) {
        return Divide(in.a, in.b)
    }
    cases := trial.Cases[input, int]{
        "10 / 2 = 5": {
            Input:    input{a: 10, b: 2},
            Expected: 5,
        },
        "9 / 3 = 3": {
            Input:    input{a: 9, b: 3},
            Expected: 3,
        },
        "divide by zero": {
            Input:       input{a: 1, b: 0},
            ExpectedErr: errors.New("cannot divide by zero"),
        },
        "negative division": {
            Input:    input{a: -10, b: 2},
            Expected: -5,
        },
    }
    trial.New(fn, cases).SubTest(t)
}
```
