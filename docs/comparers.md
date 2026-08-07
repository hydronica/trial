# Comparers

Comparers determine how trial checks equality between actual and expected values. This guide covers all available comparers and their options.

## Table of Contents

- [Overview](#overview)
- [Equal (Default)](#equal-default)
- [EqualOpt](#equalopt)
- [JSONEqual](#jsonequal)
- [JSONOpt](#jsonopt)
- [Contains](#contains)
- [CmpFuncs](#cmpfuncs)
- [Custom Comparers](#custom-comparers)

---

## Overview

A comparer is a function with this signature:

```go
type CompareFunc func(actual, expected any) (equal bool, differences string)
```

- `equal` - Returns `true` if values match
- `differences` - Human-readable string describing differences (used in failure messages)

Set a custom comparer using the `Comparer()` method:

```go
trial.New(fn, cases).Comparer(myComparer).Test(t)
```

---

## Equal (Default)

The default comparer. Wraps `cmp.Equal` from [google/go-cmp](https://github.com/google/go-cmp) with these defaults:

- **AllowAllUnexported** - Compares all fields, including private (unexported) ones
- **EquateEmpty** - Treats `nil` and empty slices/maps as equal

```go
func Equal(actual, expected any) (bool, string)
```

**When to use:** Most cases. Strict equality checking.

**Example:**
```go
// Default behavior - no need to specify
trial.New(fn, cases).Test(t)

// Or explicitly:
trial.New(fn, cases).Comparer(trial.Equal).Test(t)
```

---

## EqualOpt

Customizable equality comparer. Build your own comparison logic by combining options.

```go
func EqualOpt(optFns ...func(i any) cmp.Option) func(actual, expected any) (bool, string)
```

### Available Options

#### AllowAllUnexported

Compare all unexported (private) fields in structs. This is the default behavior in `Equal`.

```go
trial.EqualOpt(trial.AllowAllUnexported)
```

**Use case:** Testing structs within the same package where you have access to private fields.

#### IgnoreAllUnexported

Ignore all unexported (private) fields in structs.

```go
trial.EqualOpt(trial.IgnoreAllUnexported)
```

**Use case:** Testing against structs from external packages where private fields may change.

#### IgnoreFields

Exclude specific fields from comparison by name.

```go
func IgnoreFields(f ...string) func(any) cmp.Option
```

- Field names are **case-sensitive**
- Supports dot notation for nested fields: `"Parent.child"`

```go
trial.EqualOpt(
    trial.IgnoreFields("ID", "CreatedAt", "UpdatedAt"),
)
```

**Use case:** Ignoring auto-generated fields like timestamps or IDs.

#### IgnoreFieldsOf (Embedded Structs)

Ignore specific fields on a given struct type. Use this when ignoring fields in embedded structs where `IgnoreFields` cannot infer the correct type.

```go
func IgnoreFieldsOf(structType any, fields ...string) func(any) cmp.Option
```

```go
type Metadata struct {
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int
}

type User struct {
    ID       int
    Name     string
    Metadata // embedded struct
}

// To ignore fields in the embedded Metadata struct:
trial.New(fn, cases).Comparer(
    trial.EqualOpt(
        trial.IgnoreFieldsOf(Metadata{}, "CreatedAt", "UpdatedAt"),
    ),
).SubTest(t)
```

**Use case:** Ignoring fields in embedded structs where `trial.IgnoreFields` doesn't reach the nested type.

#### IgnoreTypes

Ignore all values of specified types.

```go
func IgnoreTypes(types ...any) func(any) cmp.Option
```

```go
trial.EqualOpt(
    trial.IgnoreTypes(time.Time{}, uuid.UUID{}),
)
```

**Use case:** Ignoring all time fields without listing each one.

#### ApproxTime

Consider time values equal if they differ by less than the specified duration.

```go
func ApproxTime(d time.Duration) func(any) cmp.Option
```

```go
trial.EqualOpt(
    trial.ApproxTime(time.Second),
)
```

**Use case:** Comparing times that may differ by small amounts due to execution time.

#### EquateEmpty

Treat `nil` and empty slices/maps as equal. This is the default behavior in `Equal`.

```go
trial.EqualOpt(trial.EquateEmpty)
```

**Use case:** When `nil` vs empty slice distinction doesn't matter.

### Combining Options

Options can be combined:

```go
type User struct {
    ID         int
    Name       string
    email      string    // private
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

trial.New(fn, cases).Comparer(
    trial.EqualOpt(
        trial.IgnoreAllUnexported,
        trial.IgnoreFields("ID", "CreatedAt", "UpdatedAt"),
    ),
).SubTest(t)
```

---

## JSONEqual

Semantic JSON comparison. Normalizes both sides to JSON DOM shape before comparing with `cmp.Diff`. Ignores key ordering and whitespace differences.

```go
func JSONEqual(actual, expected any) (bool, string)
```

**When to use:** Comparing struct output to embedded JSON fixtures, API response bodies, or JSON strings that differ only in formatting.

**Accepted shapes on either side:**

| Input | Normalization |
|-------|---------------|
| `string` | `json.Unmarshal` (inline, embedded, or API body) |
| `[]byte`, `json.RawMessage` | `json.Unmarshal` |
| struct, map, slice, pointer | `json.Marshal` → `json.Unmarshal` |

**Recommended fixture pattern** — load golden JSON with `//go:embed`, pass as `Expected`:

```go
//go:embed testdata/expected.json
var expectedJSON string

cases := trial.Cases[Input, string]{
    "valid response": {Input: someInput, Expected: expectedJSON},
}
trial.New(fn, cases).Comparer(trial.JSONEqual).SubTest(t)
```

**Struct vs JSON string:**

```go
type Response struct {
    Name  string `json:"name"`
    Count int    `json:"count"`
}

cases := trial.Cases[Input, string]{
    "matches golden": {
        Input:    in,
        Expected: `{"count":5,"name":"foo"}`, // key order differs from struct
    },
}
// fn returns Response; JSONEqual compares semantically
trial.New(fn, cases).Comparer(trial.JSONEqual).SubTest(t)
```

**Limitations:**

- JSON numbers become `float64` after normalization (standard Go JSON behavior)
- Unexported struct fields are not compared (JSON round-trip uses exported fields only)
- Invalid JSON returns an error naming the side with parse detail (e.g. `actual: invalid JSON: invalid character 'n' ...`)

For subset matching, ignoring dynamic JSON fields, or exact large integers, see [JSONOpt](#jsonopt).

---

## JSONOpt

Configurable JSON comparison. `JSONEqual` is equivalent to `JSONOpt()` with no options.

```go
func JSONOpt(opts ...JSONOption) CompareFunc
```

**Sugar comparer:**

```go
var JSONContains CompareFunc // JSONOpt(JSONSubset())
```

### Available Options

#### JSONSubset

After normalization, check that expected is contained in actual (extra keys in actual are OK). Array semantics match `Contains`: each expected element must appear somewhere in the actual array.

```go
trial.JSONOpt(trial.JSONSubset())
// or
trial.JSONContains
```

**Example — partial API response:**

```go
cases := trial.Cases[Input, string]{
    "has required fields": {
        Input:    in,
        Expected: `{"name":"foo","count":5}`,
    },
}
// fn returns full JSON body with id, timestamps, etc.
trial.New(fn, cases).Comparer(trial.JSONContains).SubTest(t)
```

#### JSONIgnorePaths

Remove JSON keys from both sides before comparison. Uses dot notation for nested keys. Paths refer to **JSON keys** after normalization, not Go struct field names.

```go
trial.JSONOpt(trial.JSONIgnorePaths("id", "meta.created_at"))
```

**Example — ignore dynamic fields on full equality:**

```go
trial.New(fn, cases).Comparer(
    trial.JSONOpt(trial.JSONIgnorePaths("id", "created_at")),
).SubTest(t)
```

#### JSONUseNumber

Unmarshal numbers as `json.Number` instead of `float64`. Opt-in when large integers must compare exactly.

```go
trial.JSONOpt(trial.JSONUseNumber())
```

Default `JSONEqual` keeps `float64` normalization (standard Go JSON behavior).

### Combining Options

```go
trial.New(fn, cases).Comparer(
    trial.JSONOpt(trial.JSONSubset(), trial.JSONIgnorePaths("request_id")),
).SubTest(t)
```

**Note:** `JSONOption` is separate from `EqualOpt`. Use `EqualOpt` for struct/cmp rules; use `JSONOpt` for JSON fixtures and API bodies.

---

## Contains

Subset matching comparer. Checks if the expected value is **contained in** the actual value.

```go
func Contains(x, y any) (bool, string)
```

### Symbols

- `⊇` - Superset (actual contains expected)
- `∈` - Element of (value exists in slice)

### Supported Comparisons

#### String Contains String

Checks if expected string is a substring of actual.

```go
actual := "Hello, World!"
expected := "World"
// Passes: "Hello, World!" contains "World"
```

#### String Contains []String

Checks if all expected substrings are in the actual string.

```go
actual := "The quick brown fox"
expected := []string{"quick", "fox"}
// Passes: actual contains both "quick" and "fox"
```

#### Slice Contains Value

Checks if expected value exists in the actual slice.

```go
actual := []int{1, 2, 3, 4, 5}
expected := 3
// Passes: 3 is in the slice
```

#### Slice Contains Slice

Checks if all expected values exist in the actual slice (subset).

```go
actual := []string{"a", "b", "c", "d"}
expected := []string{"b", "d"}
// Passes: both "b" and "d" are in actual
```

#### Map Contains Map

Checks if all expected keys exist in actual and their values match.

```go
actual := map[string]int{"a": 1, "b": 2, "c": 3}
expected := map[string]int{"a": 1, "c": 3}
// Passes: keys "a" and "c" exist with matching values
```

### Example

```go
func TestContains(t *testing.T) {
    fn := func(in string) (string, error) {
        return fmt.Sprintf("User %s logged in at %s", in, time.Now()), nil
    }
    cases := trial.Cases[string, string]{
        "message contains username": {
            Input:    "alice",
            Expected: "alice",
        },
    }
    trial.New(fn, cases).Comparer(trial.Contains).Test(t)
}
```

---

## CmpFuncs

Compares two function values to determine if they point to the same function.

```go
func CmpFuncs(x, y any) (b bool, s string)
```

**Note:** Due to Go's function comparison limitations, this compares function pointers, not function behavior.

```go
func TestFunctionEquality(t *testing.T) {
    myHandler := func(w http.ResponseWriter, r *http.Request) {}
    
    fn := func(in string) (func(http.ResponseWriter, *http.Request), error) {
        return myHandler, nil
    }
    cases := trial.Cases[string, func(http.ResponseWriter, *http.Request)]{
        "returns correct handler": {
            Input:    "test",
            Expected: myHandler,
        },
    }
    trial.New(fn, cases).Comparer(trial.CmpFuncs).Test(t)
}
```

---

## Custom Comparers

Create your own comparer for specialized comparison logic.

### Signature

```go
func MyComparer(actual, expected any) (bool, string)
```

### Example: Fuzzy String Matching

```go
func FuzzyMatch(actual, expected any) (bool, string) {
    a, ok1 := actual.(string)
    e, ok2 := expected.(string)
    if !ok1 || !ok2 {
        return false, "both values must be strings"
    }
    
    // Normalize: lowercase and trim
    a = strings.TrimSpace(strings.ToLower(a))
    e = strings.TrimSpace(strings.ToLower(e))
    
    if a == e {
        return true, ""
    }
    return false, fmt.Sprintf("'%s' != '%s'", a, e)
}

// Usage
trial.New(fn, cases).Comparer(FuzzyMatch).Test(t)
```

### Example: Struct with Tolerance

```go
func WithinTolerance(tolerance float64) trial.CompareFunc {
    return func(actual, expected any) (bool, string) {
        a, ok1 := actual.(float64)
        e, ok2 := expected.(float64)
        if !ok1 || !ok2 {
            return trial.Equal(actual, expected)
        }
        
        diff := math.Abs(a - e)
        if diff <= tolerance {
            return true, ""
        }
        return false, fmt.Sprintf("%f differs from %f by %f (tolerance: %f)", a, e, diff, tolerance)
    }
}

// Usage
trial.New(fn, cases).Comparer(WithinTolerance(0.001)).Test(t)
```

---

## Choosing a Comparer

| Scenario | Recommended Comparer |
|----------|---------------------|
| Exact equality (default) | `Equal` |
| JSON / struct semantic equality | `JSONEqual` or `JSONOpt()` |
| Partial JSON body (extra keys OK) | `JSONContains` or `JSONOpt(JSONSubset())` |
| Ignore dynamic JSON fields | `JSONOpt(JSONIgnorePaths(...))` |
| Exact large JSON integers | `JSONOpt(JSONUseNumber())` |
| Ignore specific fields | `EqualOpt(IgnoreFields(...))` |
| Ignore embedded struct fields | `EqualOpt(IgnoreFieldsOf(EmbeddedType{}, ...))` |
| Ignore private fields | `EqualOpt(IgnoreAllUnexported)` |
| Ignore timestamps | `EqualOpt(IgnoreFields(...))` or `ApproxTime(...)` |
| Substring/subset matching | `Contains` |
| Function comparison | `CmpFuncs` |
| Custom logic | Write your own `CompareFunc` |
