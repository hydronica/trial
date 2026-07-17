# Helper Functions

Trial provides helper functions to simplify test setup.

## Table of Contents

- [Input Helpers](#input-helpers)
- [Pointer Helpers](#pointer-helpers)
- [Time Helpers](#time-helpers)
- [Error Helpers](#error-helpers)
- [Capture Utilities](#capture-utilities)

---

## Input Helpers

### Args

Converts multiple arguments into an `Input` value.

```go
func Args(args ...any) Input
```

**Example:**
```go
fn := func(in trial.Input) (int, error) {
    a := in.Slice(0).Int()
    b := in.Slice(1).Int()
    return a + b, nil
}
cases := trial.Cases[trial.Input, int]{
    "add": {
        Input:    trial.Args(5, 3),
        Expected: 8,
    },
}
```

### Input Methods

| Method | Return | Description |
|--------|--------|-------------|
| `String()` | `string` | Get as string (panics on struct/ptr/slice/map/array/chan) |
| `Int()` | `int` | Get as int (parses strings) |
| `Uint()` | `uint` | Get as uint (parses strings) |
| `Float64()` | `float64` | Get as float64 (parses strings) |
| `Bool()` | `bool` | Get as bool (parses strings) |
| `Slice(i int)` | `Input` | Get element at index |
| `Map(key any)` | `Input` | Get value for key |
| `Interface()` | `any` | Get raw value |

---

## Pointer Helpers

### Pointer[T]

Creates a pointer to a primitive value.

```go
func Pointer[T primitives](v T) *T
```

**Supported types:** `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `string`, `bool`

**Example:**
```go
type Config struct {
    Host  *string
    Port  *int
    Debug *bool
}

cases := trial.Cases[Config, bool]{
    "debug enabled": {
        Input: Config{
            Host:  trial.Pointer("localhost"),
            Port:  trial.Pointer(8080),
            Debug: trial.Pointer(true),
        },
        Expected: true,
    },
}
```

### Deprecated Pointer Helpers

Use `Pointer[T]()` instead of: `IntP`, `Int8P`, `Int16P`, `Int32P`, `Int64P`, `UintP`, `Uint8P`, `Uint16P`, `Uint32P`, `Uint64P`, `Float32P`, `Float64P`, `BoolP`, `StringP`

---

## Time Helpers

Time helpers parse time strings and **panic on error** (fail fast in test setup).

| Function | Format | Example |
|----------|--------|---------|
| `Day(value)` | `2006-01-02` | `Day("2024-01-15")` |
| `Hour(value)` | `2006-01-02T15` | `Hour("2024-01-15T14")` |
| `Time(layout, value)` | Custom | `Time(time.RFC3339, "2024-01-15T14:30:00Z")` |
| `Times(layout, values...)` | Custom (slice) | `Times(time.RFC3339, "...", "...")` |
| `TimeP(layout, value)` | Custom (pointer) | `TimeP(time.RFC3339, "...")` |

**Example:**
```go
cases := trial.Cases[input, int]{
    "one day difference": {
        Input: input{
            start: trial.Day("2024-01-15"),
            end:   trial.Day("2024-01-16"),
        },
        Expected: 24,  // hours
    },
}
```

### Deprecated Time Helpers

Use `Hour()` instead of `TimeHour()`, use `Day()` instead of `TimeDay()`.

---

## Error Helpers

### ErrType

Wraps an error for **type-based comparison**. When used with `ExpectedErr`, trial checks that the returned error is the same type (not message).

```go
func ErrType(err error) error
```

**Example:**
```go
type ValidationError struct{ Field string }
func (e ValidationError) Error() string { return "invalid " + e.Field }

cases := trial.Cases[string, string]{
    "returns validation error": {
        Input:       "",
        ExpectedErr: trial.ErrType(ValidationError{}),
    },
}
```

---

## Capture Utilities

Capture log output, stdout, or stderr during tests.

### Functions

| Function | Captures |
|----------|----------|
| `CaptureLog()` | Standard `log` package output |
| `CaptureStdOut()` | `os.Stdout` |
| `CaptureStdErr()` | `os.Stderr` |

### Methods

| Method | Return | Description |
|--------|--------|-------------|
| `ReadAll()` | `string` | Stop capturing, return all output |
| `ReadLines()` | `[]string` | Stop capturing, return lines |

**Example:**
```go
func TestLogging(t *testing.T) {
    c := trial.CaptureLog()
    
    log.Println("Hello, World!")
    
    output := c.ReadAll()
    if !strings.Contains(output, "Hello, World!") {
        t.Error("expected log message not found")
    }
}
```

**Important:**
- Always call `ReadAll()` or `ReadLines()` to restore normal output
- Only one capture at a time
