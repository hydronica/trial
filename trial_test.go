package trial

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestMain(t *testing.M) {
	localTest = true
	t.Run()
	localTest = false
}
func TestTrial_TestCase(t *testing.T) {
	divideFn := func(in Input) (interface{}, error) {
		return func(a, b int) (int, error) {
			if b == 0 {
				return 0, errors.New("divide by zero")
			}
			return a / b, nil
		}(in.Slice(0).Int(), in.Slice(1).Int())
	}

	panicFn := func(in Input) (interface{}, error) {
		return func(s string) string {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				panic(err)
			}
			return t.Format("2006-01-02")
		}(in.String()), nil
	}

	cases := map[string]struct {
		trial     *Trial
		Case      Case
		expResult result
	}{
		"1/1 - pass case": {
			trial: New(divideFn, nil),
			Case: Case{
				Input:    []interface{}{1, 1},
				Expected: 1,
			},
			expResult: result{true, `PASS: "1/1 - pass case"`},
		},
		"1/0 - error check": {
			trial: New(divideFn, nil),
			Case: Case{
				Input:     []interface{}{1, 0},
				ShouldErr: true,
			},
			expResult: result{true, `PASS: "1/0 - error check"`},
		},
		"1/0 - unexpected error": {
			trial: New(divideFn, nil),
			Case: Case{
				Input: []interface{}{1, 0},
			},
			expResult: result{false, `FAIL: "1/0 - unexpected error" unexpected error 'divide by zero'`},
		},
		"10/2 - unexpected result": {
			trial: New(divideFn, nil),
			Case: Case{
				Input:    []interface{}{10, 2},
				Expected: 10,
			},
			expResult: result{false, "FAIL: \"10/2 - unexpected result\" \n{int}:\n\t-: 5\n\t+: 10\n"},
		},
		"parse time": {
			trial: New(panicFn, nil),
			Case: Case{
				Input:    "2018-01-02T00:00:00Z",
				Expected: "2018-01-02",
			},
			expResult: result{true, `PASS: "parse time"`},
		},
		"parse time with panic": {
			trial: New(panicFn, nil),
			Case: Case{
				Input:       "invalid",
				ShouldPanic: true,
			},
			expResult: result{true, `PASS: "parse time with panic"`},
		},
		"parse time with unexpected panic": {
			trial: New(panicFn, nil),
			Case: Case{
				Input: "invalid",
			},
			expResult: result{false, `PANIC: "parse time with unexpected panic" parsing time "invalid" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid" as "2006"`},
		},
		"expected panic did not occur": {
			trial: New(func(Input) (interface{}, error) {
				return nil, nil
			}, nil),
			Case: Case{
				ShouldPanic: true,
			},
			expResult: result{false, `FAIL: "expected panic did not occur" did not panic`},
		},
		"test should error but no error occurred": {
			trial: New(func(Input) (interface{}, error) {
				return nil, nil
			}, nil),
			Case: Case{
				ShouldErr: true,
			},
			expResult: result{false, `FAIL: "test should error but no error occurred" should error`},
		},
		"expected error string match": {
			trial: New(func(Input) (interface{}, error) {
				return nil, errors.New("test error")
			}, nil),
			Case: Case{
				ExpectedErr: errors.New("test error"),
			},
			expResult: result{true, `PASS: "expected error string match"`},
		},
		"expected error string does not match": {
			trial: New(divideFn, nil),
			Case: Case{
				Input:       Args(10, 0),
				ExpectedErr: errors.New("test error"),
			},
			expResult: result{false, `FAIL: "expected error string does not match" error "divide by zero" does not match expected "test error"`},
		},
		"expected error of type testErr": {
			trial: New(func(Input) (interface{}, error) {
				return nil, testErr{}
			}, nil),
			Case: Case{
				ExpectedErr: ErrType(testErr{}),
			},
			expResult: result{true, `PASS: "expected error of type testErr"`},
		},
		"error type testErr with nil response": {
			trial: New(func(Input) (interface{}, error) {
				return nil, nil
			}, nil),
			Case: Case{
				ExpectedErr: ErrType(testErr{}),
			},
			expResult: result{false, `FAIL: "error type testErr with nil response"`},
		},
		"error type testErr with mismatch response": {
			trial: New(func(Input) (interface{}, error) {
				return nil, errors.New("some error")
			}, nil),
			Case: Case{
				ExpectedErr: ErrType(testErr{}),
			},
			expResult: result{false, `FAIL: "error type testErr with mismatch response"`},
		},
	}
	for msg, test := range cases {
		r := test.trial.testCase(msg, test.Case)
		if r.Success != test.expResult.Success || !strings.Contains(r.Message, test.expResult.Message) {
			t.Errorf("FAIL: %q %v", msg, cmp.Diff(r, test.expResult))
		} else {
			t.Logf("PASS: %q", msg)
		}
	}
}

type testErr struct{}

func (e testErr) Error() string {
	return ""
}

func TestInput(t *testing.T) {
	type tester struct {
		shouldPanic bool
		fn          func() interface{}
		expected    interface{}
	}
	cases := map[string]tester{
		"string": tester{
			fn:       func() interface{} { return newInput("hello world").String() },
			expected: "hello world",
		},
		"string panic": tester{
			fn:          func() interface{} { return newInput(12).String() },
			shouldPanic: true,
		},
		"bool": {
			fn:       func() interface{} { return newInput(true).Bool() },
			expected: true,
		},
		"int": {
			fn:       func() interface{} { return newInput(12).Int() },
			expected: 12,
		},
		"uint": {
			fn:       func() interface{} { return newInput(12).Uint() },
			expected: uint(12),
		},
		"float64": {
			fn:       func() interface{} { return newInput(12.4).Float64() },
			expected: 12.4,
		},
		"float64 (float32)": {
			fn:       func() interface{} { return newInput(float32(12.4)).Float64() },
			expected: 12.399999618530273,
		},
		"float64 (int)": {
			fn:       func() interface{} { return newInput(12).Float64() },
			expected: float64(12),
		},
		"map[string]string": {
			fn:       func() interface{} { return newInput(map[string]string{"abc": "def"}).Map("abc").String() },
			expected: "def",
		},
		"map[int]string": {
			fn:       func() interface{} { return newInput(map[int]string{12: "def"}).Map(12).String() },
			expected: "def",
		},
		"map[interface]interface": {
			fn:       func() interface{} { return newInput(map[interface{}]interface{}{12: "def"}).Map(12).String() },
			expected: "def",
		},
		"[]string": {
			fn:       func() interface{} { return newInput([]string{"ab", "cd", "ef", "g"}).Slice(2).String() },
			expected: "ef",
		},
		"slice out of bounds": {
			fn:          func() interface{} { return newInput([]string{}).Slice(2).String() },
			shouldPanic: true,
		},
		"invalid type": {
			fn:          func() interface{} { return newInput([]string{"ab", "cd", "ef", "g"}).Map(2).String() },
			shouldPanic: true,
		},
	}
	for name, in := range cases {
		// panic wrapper
		func() {
			var result interface{}
			defer func() {
				rec := recover()
				if rec == nil && in.shouldPanic {
					t.Errorf("FAIL: %q should panic", name)
				} else if rec != nil && !in.shouldPanic {
					t.Errorf("PANIC: %q %v", name, rec)
				} else if b, s := Equal(result, in.expected); !b {
					t.Errorf("FAIL: %q %s", name, s)
				} else {
					t.Logf("Pass: %q", name)
				}
			}()
			result = in.fn()
		}()
	}
}
