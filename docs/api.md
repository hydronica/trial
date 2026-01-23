# Trial API Reference

Quick reference for the `hydronica/trial` testing framework. See linked docs for detailed explanations and examples.

## Core Types

```go
// Main test runner
type Trial[In any, Out any] struct { /* ... */ }

// Map of named test cases
type Cases[In any, Out any] map[string]Case[In, Out]

// Single test case
type Case[In any, Out any] struct {
    Input       In
    Expected    Out
    ShouldErr   bool   // expect any error
    ExpectedErr error  // expect error containing this string (or use ErrType for type check)
    ShouldPanic bool   // expect panic
}

// Dynamic input wrapper (use with Args)
type Input struct { /* ... */ }

// Custom comparer signature
type CompareFunc func(actual, expected interface{}) (equal bool, differences string)
```

## Core Functions

| Function | Description |
|----------|-------------|
| `New(fn, cases)` | Create Trial instance |
| `ErrType(err)` | Wrap error for type-based comparison |

## Trial Methods

| Method | Description |
|--------|-------------|
| `.Test(t)` | Run all cases in single test |
| `.SubTest(t)` | Run each case as subtest (recommended) |
| `.Comparer(fn)` | Set custom comparison function |
| `.Timeout(d)` | Set max duration per case |

## Case Fields

| Field | Type | Description |
|-------|------|-------------|
| `Input` | `In` | Value passed to test function |
| `Expected` | `Out` | Expected return value |
| `ShouldErr` | `bool` | Expect any error |
| `ExpectedErr` | `error` | Expect error containing string (uses `strings.Contains`) |
| `ShouldPanic` | `bool` | Expect panic |

**Notes:**
- `ExpectedErr` implies `ShouldErr`, no need to set both
- Use `ErrType(MyError{})` with `ExpectedErr` to check error types

## Input Methods

For use with `trial.Args()`. See [helpers.md](helpers.md#input-helpers) for details.

| Method | Return | Description |
|--------|--------|-------------|
| `String()` | `string` | Get as string |
| `Int()` | `int` | Get as int (parses strings) |
| `Uint()` | `uint` | Get as uint |
| `Float64()` | `float64` | Get as float64 |
| `Bool()` | `bool` | Get as bool |
| `Slice(i)` | `Input` | Get element at index |
| `Map(key)` | `Input` | Get value for key |
| `Interface()` | `interface{}` | Get raw value |

## Comparers

See [comparers.md](comparers.md) for detailed documentation.

| Comparer | Description |
|----------|-------------|
| `Equal` | Default. Strict equality via `cmp.Equal` |
| `Contains` | Subset/substring matching |
| `EqualOpt(opts...)` | Customizable equality |
| `CmpFuncs` | Compare function pointers |

### EqualOpt Options

| Option | Description |
|--------|-------------|
| `AllowAllUnexported` | Compare private fields (default in `Equal`) |
| `IgnoreAllUnexported` | Skip private fields |
| `IgnoreFields(names...)` | Skip specific fields |
| `IgnoreTypes(values...)` | Skip specific types |
| `ApproxTime(d)` | Fuzzy time comparison |
| `EquateEmpty` | nil == empty slice/map (default in `Equal`) |

## Helpers

See [helpers.md](helpers.md) for detailed documentation.

| Function | Description |
|----------|-------------|
| `Args(values...)` | Create Input from multiple args |
| `Pointer[T](v)` | Create pointer to primitive |
| `Time(layout, value)` | Parse time (panics on error) |
| `Day(value)` | Parse `2006-01-02` |
| `Hour(value)` | Parse `2006-01-02T15` |
| `Times(layout, values...)` | Parse multiple times |
| `TimeP(layout, value)` | Parse time, return pointer |
| `CaptureLog()` | Capture log output |
| `CaptureStdOut()` | Capture stdout |
| `CaptureStdErr()` | Capture stderr |
