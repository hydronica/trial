package trial

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

type (
	// TestFunc a wrapper function used to setup the method being tested.
	TestFunc func(args ...interface{}) (result interface{}, err error)

	// EqualFunc compares actual and expected to determine equality. It should return
	// a human readable string representing the differences between actual and
	// expected.The output string may use the "-" symbol to indicate elements
	// removed from actual, and the "+" symbol to indicate elements
	// added to expected
	EqualFunc func(actual, expected interface{}) (equal bool, differences string)
)

// Trial framework used to test different logical states
type Trial struct {
	cases   map[string]Case
	testFn  TestFunc
	equalFn EqualFunc
}

// Cases made during the trial
type Cases map[string]Case

// Case made during the trial of your code
type Case struct {
	Input    interface{}
	Expected interface{}

	// testing conditions
	ShouldErr   bool  // is an error expected
	ExpectedErr error // the error that was expected (nil is no error expected)
	ShouldPanic bool  // is a panic expected
}

// New trial for your code
func New(fn TestFunc, cases map[string]Case) *Trial {
	if cases == nil {
		cases = make(map[string]Case)
	}
	return &Trial{
		cases:   cases,
		testFn:  fn,
		equalFn: Equal,
	}
}

// EqualFn override the default comparison method used.
// see ContainsFn(x, y interface{}) (bool, string)
func (t *Trial) EqualFn(fn EqualFunc) *Trial {
	t.equalFn = fn
	return t
}

// Test all cases provided
func (t *Trial) Test(tst testing.TB) {
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

func (t *Trial) testCase(msg string, test Case) (r result) {
	var finished bool
	defer func() {
		rec := recover()
		if rec == nil && test.ShouldPanic {
			r = fail("FAIL: %q did not panic", msg)
		} else if rec != nil && !test.ShouldPanic {
			r = fail("PANIC: %q %v\n%s", msg, rec, cleanStack())
		} else if !finished {
			r = pass("PASS: %q", msg)
		}
	}()
	var err error
	var result interface{}
	if inputs, ok := test.Input.([]interface{}); ok {
		result, err = t.testFn(inputs...)
	} else {
		result, err = t.testFn(test.Input)
	}

	if (test.ShouldErr && err == nil) || (test.ExpectedErr != nil && err == nil) {
		finished = true
		return fail("FAIL: %q should error", msg)
	} else if !test.ShouldErr && err != nil && test.ExpectedErr == nil {
		finished = true
		return fail("FAIL: %q unexpected error %s", msg, err.Error())
	} else if test.ExpectedErr != nil && !isExpectedError(err, test.ExpectedErr) {
		finished = true
		return fail("FAIL: %q error %q does not match expected %q", msg, err, test.ExpectedErr)
	} else if !test.ShouldErr && test.ExpectedErr == nil {
		if equal, diff := t.equalFn(result, test.Expected); !equal {
			finished = true
			return fail("FAIL: %q differences %v", msg, diff)
		}
		finished = true
		return pass("PASS: %q", msg)
	}
	return pass("PASS: %q", msg)
}

// cleanStack removes unhelpful lines from a panic stack track
func cleanStack() (s string) {
	for _, ln := range strings.Split(string(debug.Stack()), "\n") {
		if strings.Contains(ln, "/jbsmith7741/trial") {
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
	Success bool
	Message string
}

func pass(format string, args ...interface{}) result {
	return result{
		Success: true,
		Message: fmt.Sprintf(format, args...),
	}
}

func fail(format string, args ...interface{}) result {
	return result{
		Success: false,
		Message: fmt.Sprintf(format, args...),
	}
}

type tHelper interface {
	Helper()
}
