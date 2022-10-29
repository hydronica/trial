package trial

import (
	"errors"
	"strings"
	"testing"
	"time"
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
		trial     *Trial[Input, any]
		Case      Case[Input, any]
		expResult result
	}{
		"1/1 - pass case": {
			trial: New(divideFn, nil),
			Case: Case[Input, any]{
				Input:    Args(1, 1),
				Expected: 1,
			},
			expResult: result{Success: true, Message: `PASS: "1/1 - pass case"`},
		},
		"1/0 - error check": {
			trial: New(divideFn, nil),
			Case: Case[Input, any]{
				Input:     Args(1, 0),
				ShouldErr: true,
			},
			expResult: result{Success: true, Message: `PASS: "1/0 - error check"`},
		},
		"1/0 - unexpected error": {
			trial: New(divideFn, nil),
			Case: Case[Input, any]{
				Input: Args(1, 0),
			},
			expResult: result{Success: false, Message: `FAIL: "1/0 - unexpected error" unexpected error 'divide by zero'`},
		},
		"10/2 - unexpected result": {
			trial: New(divideFn, nil),
			Case: Case[Input, any]{
				Input:    Args(10, 2),
				Expected: 10,
			},
			expResult: result{Success: false, Message: "FAIL: \"10/2 - unexpected result\""},
		},
		"parse time": {
			trial: New(panicFn, nil),
			Case: Case[Input, any]{
				Input:    Args("2018-01-02T00:00:00Z"),
				Expected: "2018-01-02",
			},
			expResult: result{Success: true, Message: `PASS: "parse time"`},
		},
		"parse time with panic": {
			trial: New(panicFn, nil),
			Case: Case[Input, any]{
				Input:       Args("invalid"),
				ShouldPanic: true,
			},
			expResult: result{Success: true, Message: `PASS: "parse time with panic"`},
		},
		"parse time with unexpected panic": {
			trial: New(panicFn, nil),
			Case: Case[Input, any]{
				Input: Args("invalid"),
			},
			expResult: result{Success: false, Message: `PANIC: "parse time with unexpected panic" parsing time "invalid" as "2006-01-02T15:04:05Z07:00": cannot parse "invalid" as "2006"`},
		},
		"expected panic did not occur": {
			trial: New(func(Input) (interface{}, error) {
				return nil, nil
			}, nil),
			Case: Case[Input, any]{
				ShouldPanic: true,
			},
			expResult: result{Success: false, Message: `FAIL: "expected panic did not occur" did not panic`},
		},
		"test should error but no error occurred": {
			trial: New(func(Input) (interface{}, error) {
				return nil, nil
			}, nil),
			Case: Case[Input, any]{
				ShouldErr: true,
			},
			expResult: result{Success: false, Message: `FAIL: "test should error but no error occurred" should error`},
		},
		"expected error string match": {
			trial: New(func(Input) (interface{}, error) {
				return nil, errors.New("test error")
			}, nil),
			Case: Case[Input, any]{
				ExpectedErr: errors.New("test error"),
			},
			expResult: result{Success: true, Message: `PASS: "expected error string match"`},
		},
		"expected error string does not match": {
			trial: New(divideFn, nil),
			Case: Case[Input, any]{
				Input:       Args(10, 0),
				ExpectedErr: errors.New("test error"),
			},
			expResult: result{Success: false, Message: `FAIL: "expected error string does not match" error "divide by zero" does not match expected "test error"`},
		},
		"expected error of type testErr": {
			trial: New(func(Input) (interface{}, error) {
				return nil, testErr{}
			}, nil),
			Case: Case[Input, any]{
				ExpectedErr: ErrType(testErr{}),
			},
			expResult: result{Success: true, Message: `PASS: "expected error of type testErr"`},
		},
		"error type testErr with nil response": {
			trial: New(func(Input) (interface{}, error) {
				return nil, nil
			}, nil),
			Case: Case[Input, any]{
				ExpectedErr: ErrType(testErr{}),
			},
			expResult: result{Success: false, Message: `FAIL: "error type testErr with nil response"`},
		},
		"error type testErr with mismatch response": {
			trial: New(func(Input) (interface{}, error) {
				return nil, errors.New("some error")
			}, nil),
			Case: Case[Input, any]{
				ExpectedErr: ErrType(testErr{}),
			},
			expResult: result{Success: false, Message: `FAIL: "error type testErr with mismatch response"`},
		},
		"timeout error": {
			trial: New(func(Input) (interface{}, error) {
				time.Sleep(time.Second)
				return nil, nil
			}, nil).Timeout(time.Millisecond),
			Case: Case[Input, any]{
				ExpectedErr: errors.New("timeout"),
			},
			expResult: result{Success: false, Message: `FAIL: "timeout error" timeout after 1ms`},
		},
	}
	for msg, test := range cases {
		r := test.trial.testCase(msg, test.Case)
		if r.Success != test.expResult.Success || !strings.Contains(r.Message, test.expResult.Message) {
			t.Errorf("\033[31mFAIL: %q\n%v\033[39m", msg, r.string())
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
		"string (int)": tester{
			fn:       func() interface{} { return newInput(123).String() },
			expected: "123",
		},
		"string (float)": tester{
			fn:       func() interface{} { return newInput(12.8).String() },
			expected: "12.8",
		},
		"string (bool)": tester{
			fn:       func() interface{} { return newInput(true).String() },
			expected: "true",
		},
		"string panic": tester{
			fn:          func() interface{} { return newInput(struct{}{}).String() },
			shouldPanic: true,
		},
		"bool": {
			fn:       func() interface{} { return newInput(true).Bool() },
			expected: true,
		},
		"bool (string)": {
			fn: func() interface{} {
				newInput("false").Bool()
				return newInput("true").Bool()
			},
			expected: true,
		},
		"bool (invalid)": {
			fn:          func() interface{} { return newInput("abc").Bool() },
			shouldPanic: true,
		},
		"int": {
			fn:       func() interface{} { return newInput(12).Int() },
			expected: 12,
		},
		"int (string)": {
			fn:       func() interface{} { return newInput("12").Int() },
			expected: 12,
		},
		"int (invalid)": {
			fn:          func() interface{} { return newInput("abc").Int() },
			shouldPanic: true,
		},
		"uint": {
			fn:       func() interface{} { return newInput(12).Uint() },
			expected: uint(12),
		},
		"uint (string)": {
			fn:       func() interface{} { return newInput("12").Uint() },
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
		"float64 (string)": {
			fn:       func() interface{} { return newInput("12.5").Float64() },
			expected: 12.5,
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
			fn: func() interface{} {
				in := newInput([]string{"ab", "cd", "ef", "g"})
				in.Slice(0).String()
				return in.Slice(2).String()
			},
			expected: "ef",
		},
		"[]int": {
			fn: func() interface{} {
				in := newInput([]interface{}{1, 2, 3, 4})
				in.Slice(0).Int()
				return in.Slice(2).Int()
			},
			expected: 3,
		},
		"slice out of bounds": {
			fn:          func() interface{} { return newInput([]string{}).Slice(2).String() },
			shouldPanic: true,
		},
		"invalid type": {
			fn:          func() interface{} { return newInput([]string{"ab", "cd", "ef", "g"}).Map(2).String() },
			shouldPanic: true,
		},
		"nil": {
			fn:       func() interface{} { return newInput(nil).Interface() },
			expected: nil,
		},
	}
	for name, in := range cases {
		// panic wrapper
		t.Run(name, func(t *testing.T) {
			var result interface{}
			defer func() {
				rec := recover()
				if rec == nil && in.shouldPanic {
					t.Error("FAIL: should panic")
				} else if rec != nil && !in.shouldPanic {
					t.Errorf("PANIC: %v", rec)
				} else if b, s := Equal(result, in.expected); !b {
					t.Errorf("FAIL: %s", s)
				}
			}()
			result = in.fn()
		})
	}
}

/*
// NOTE: UNCOMMENT for verification
// this test is use to verify that the failure cases
// in subTest display the line that SubTest is called rather than trial.go:96
func TestSub(t *testing.T) {

	fn := func(input Input) (interface{}, error) {
		return 1, nil
	}
	cases := Cases{
		"not equal": {Expected: 12},
	}
	New(fn, cases).SubTest(t)

	t.Logf("%q", s)
}
*/
