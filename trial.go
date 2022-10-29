package trial

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

var localTest = false

type (
	// TestFunc a wrapper function used to setup the method being tested.
	TestFunc func(in Input) (result interface{}, err error)

	// CompareFunc compares actual and expected to determine equality. It should return
	// a human readable string representing the differences between actual and
	// expected.
	// Symbols with meaning:
	// "-" elements missing from actual
	// "+" elements missing from expected
	CompareFunc               func(actual, expected interface{}) (equal bool, differences string)
	testFunc[In any, Out any] func(in In) (result Out, err error)
)

// Comparer interface is implemented by an object to check for equality
// and show any differences found
type Comparer interface {
	Equals(interface{}) (bool, string)
}

// Trial framework used to test different logical states
type Trial[In any, Out any] struct {
	cases   map[string]Case[In, Out]
	testFn  testFunc[In, Out]
	equalFn CompareFunc
	timeout time.Duration
}

// Cases made during the trial
type Cases[In any, Out any] map[string]Case[In, Out]

// Case made during the trial of your code
type Case[In any, Out any] struct {
	Input    In
	Expected Out

	// testing conditions
	ShouldErr   bool  // is an error expected
	ExpectedErr error // the error that was expected (nil is no error expected)
	ShouldPanic bool  // is a panic expected
}

func New[In any, Out any](fn func(In) (Out, error), cases map[string]Case[In, Out]) *Trial[In, Out] {
	if cases == nil {
		cases = make(map[string]Case[In, Out])
	}

	return &Trial[In, Out]{
		cases:   cases,
		testFn:  fn,
		equalFn: Equal,
	}
}

// EqualFn override the default comparison method used.
// see ContainsFn(x, y interface{}) (bool, string)
// deprecated
func (t *Trial[In, Out]) EqualFn(fn CompareFunc) *Trial[In, Out] {
	return t.Comparer(fn)
}

// Comparer override the default comparison function.
// see Contains(x, y interface{}) (bool, string)
// see Equals(x, y interface{}) (bool, string)
func (t *Trial[In, Out]) Comparer(fn CompareFunc) *Trial[In, Out] {
	t.equalFn = fn
	return t
}

// SubTest runs all cases as individual subtests
func (t *Trial[In, Out]) SubTest(tst testing.TB) {
	if h, ok := tst.(tHelper); ok {
		h.Helper()
	}

	for msg, test := range t.cases {
		tst.(*testing.T).Run(msg, func(tb *testing.T) {
			tb.Helper()
			r := t.testCase(msg, test)
			if !r.Success {
				s := strings.Replace(r.Message, "\""+msg+"\"", "", 1)
				s = strings.Replace(s, "FAIL:", "", 1)
				tb.Error("\033[31m" + strings.TrimLeft(s, " \n") + "\033[39m")
			}
		})
	}
}

// Timeout will make sure that a test case has finished
// within the timeout or the test will fail.
func (t *Trial[In, Out]) Timeout(d time.Duration) *Trial[In, Out] {
	t.timeout = d
	return t
}

// Test all cases provided
func (t *Trial[In, Out]) Test(tst testing.TB) {
	if h, ok := tst.(tHelper); ok {
		h.Helper()
	}
	for msg, test := range t.cases {
		r := t.testCase(msg, test)
		if r.Success {
			tst.Log(r.Message)
		} else {
			tst.Error("\033[31m" + r.Message + "\033[39m")
		}
	}
}

func (t *Trial[In, Out]) testCase(msg string, test Case[In, Out]) result {
	// setup
	done := make(chan *result)
	ctx := context.Background()
	if t.timeout > time.Nanosecond {
		ctx, _ = context.WithTimeout(context.Background(), t.timeout)
	}
	// run the test function
	go func() {
		r := &result{}
		defer func() { // panic recovery and check
			rec := recover()
			r.panicCheck = rec != nil
			if rec == nil && test.ShouldPanic {
				r.fail("FAIL: %q did not panic", msg)
				r.panicCheck = true
			} else if rec != nil && !test.ShouldPanic {
				r.fail("PANIC: %q %v\n%s", msg, rec, cleanStack())
			} else {
				r.pass("PASS: %q", msg)
			}
			done <- r // send result to channel
		}()
		r.value, r.err = t.testFn(test.Input)
	}()
	result := &result{}
	select {
	case result = <-done:
		if result.panicCheck {
			return *result
		}
	case <-ctx.Done():
		result.fail("FAIL: %q timeout after %s", msg, t.timeout.String())
		return *result
	}

	if (test.ShouldErr && result.err == nil) || (test.ExpectedErr != nil && result.err == nil) {
		result.fail("FAIL: %q should error", msg)
	} else if !test.ShouldErr && result.err != nil && test.ExpectedErr == nil {
		result.fail("FAIL: %q unexpected error '%s'", msg, result.err.Error())
	} else if test.ExpectedErr != nil && !isExpectedError(result.err, test.ExpectedErr) {
		result.fail("FAIL: %q error %q does not match expected %q", msg, result.err, test.ExpectedErr)
	} else if !test.ShouldErr && test.ExpectedErr == nil {
		if equal, diff := t.equalFn(result.value, test.Expected); !equal {
			result.fail("FAIL: %q \n%s", msg, diff)
		} else {
			result.pass("PASS: %q", msg)
		}
	}
	return *result
}

// cleanStack removes unhelpful lines from a panic stack track
func cleanStack() (s string) {
	for _, ln := range strings.Split(string(debug.Stack()), "\n") {
		if !localTest && strings.Contains(ln, "/hydronica/trial") {
			continue
		}
		if strings.Contains(ln, "go/src/runtime/debug/stack.go") {
			continue
		}
		if strings.Contains(ln, "go/src/runtime/panic.go") {
			continue
		}
		s += ln + "\n"
	}
	return s
}

func isExpectedError(actual, expected error) bool {
	if err, ok := expected.(errCheck); ok {
		return reflect.TypeOf(actual) == reflect.TypeOf(err.err)
	}
	return strings.Contains(actual.Error(), expected.Error())
}

type errCheck struct {
	err error
}

func (e errCheck) Error() string {
	return e.err.Error()
}

// ErrType can be used with ExpectedErr to check
// that the expected err is of a certain type
func ErrType(err error) error {
	return errCheck{err}
}

type result struct {
	Success    bool
	Message    string
	value      interface{}
	err        error
	panicCheck bool
}

func (r *result) pass(format string, args ...interface{}) {
	r.Success = true
	r.Message = fmt.Sprintf(format, args...)
}

func (r *result) fail(format string, args ...interface{}) {
	r.Success = false
	r.Message = fmt.Sprintf(format, args...)
}

func (r result) string() string {
	return fmt.Sprintf("{Success: %v, Message: %s, value: %v, err: %v, paniced: %v}",
		r.Success, r.Message, r.value, r.err, r.panicCheck)

}

type tHelper interface {
	Helper()
}
