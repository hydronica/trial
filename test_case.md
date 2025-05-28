# Table-Driven Testing with hydronica/trial

The `hydronica/trial` package makes table-driven testing in Go simple and expressive. Below are practical examples and patterns for using this library, including time helper functions. For more, see the [trial wiki](https://github.com/hydronica/trial/wiki).

---

## Principles 
1. Keep it simple
   1. Use primitives when possible for input and output. 
   2. Functions with multiple args can be passed in by creating an input struct with each arg
2. Return errors in the test function. Use nil if needed  
3. Test cases should be self-contained
4. limit scope of custom structs and helper function by embedding them without the Test Function

## 

```go
import (
    "testing"
    "github.com/hydronica/trial"
)

func Add(i1, i2 int) int {
    return i1 + i2
}

func TestAdd(t *testing.T) {
    type input struct {
        i1 int 
        i2 int
    }
    testFn := func(in input) (int, error) {
        return Add(in.i1, in.i2), nil
    }
    cases := trial.Cases[input, int]{
        "Add two numbers": {
            Input:    input{i1:1, i2:2},
            Expected: 3,
        },
    }
    trial.New(testFn, cases).Test(t)
}
```

---

## Testing with error and using output struct

```go
import (
    "encoding/json"
    "errors"
    "testing"
    "github.com/hydronica/trial"
)

type output struct {
    Name  string
    Count int
    Value float64
}

func TestMarshal(t *testing.T) {
    fn := func(s string) (output, error) {
        v := output{}
        err := json.Unmarshal([]byte(s), &v)
        return v, err
    }
    cases := trial.Cases[string, output]{
        "valid": {
            Input:    `{"Name":"abc", "Count":10, "Value": 12.34}`,
            Expected: output{Name: "abc", Count: 10, Value: 12.34},
        },
        "invalid type": {
            Input:       `{"Value", "abc"}`,
            ExpectedErr: errors.New("invalid character"),
        },
        "invalid json": {
            Input:     `{`,
            ShouldErr: true,
        },
    }
    trial.New(fn, cases).SubTest(t)
}
```

---

## Using Time Helper Functions

The package provides helpers for time parsing in tests:
- `TimeHour(s string)` — format: `"2006-01-02T15"`
- `TimeDay(s string)` — format: `"2006-01-02"`
- `Time(layout, value string)`
- `Times(layout string, values ...string)`
- `TimeP(layout, s string) *time.Time`

## Using Pointer Helper Functions

The `Pointer[T]` helper function makes it easy to create pointers to primitive values in your test cases. This is particularly useful when testing functions that expect pointer parameters.

Supported primitive types:
- Integers: `int`, `int8`, `int16`, `int32`, `int64`
- Unsigned integers: `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- Floats: `float32`, `float64`
- `string`
- `bool`


---

## Using Custom Comparers

You can customize how trial compares expected and actual values using comparers. For example, you may want to ignore unexported fields or certain struct fields when comparing results.

```go
import (
    "testing"
    "time"
    "github.com/hydronica/trial"
)

type Example struct {
    Field1     int
    Field2     int
    ignoreMe   string    // unexported, will be ignored
    Timestamp  time.Time // will be ignored
    LastUpdate time.Time // will be ignored
}

func TestWithComparer(t *testing.T) {
    fn := func(i int) (Example, error) {
        return Example{
            Field1: i,
            Field2: i * 2,
            ignoreMe: "should be ignored",
            Timestamp: time.Now(),
            LastUpdate: time.Now(),
        }, nil
    }
    cases := trial.Cases[int, Example]{
        "basic": {
            Input:    2,
            Expected: Example{Field1: 2, Field2: 4},
        },
    }
    trial.New(fn, cases).Comparer(
        trial.EqualOpt(
            trial.IgnoreAllUnexported,
            trial.IgnoreFields("Timestamp", "LastUpdate"),
        ),
    ).SubTest(t)
}
```

This will compare only the exported fields `Field1` and `Field2`, ignoring the unexported and specified fields.

---
