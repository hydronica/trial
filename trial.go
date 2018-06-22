package trial

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type (
	TestFunc  func(args ...interface{}) (result interface{}, err error)
	DiffFunc  func(actual interface{}, expected interface{}) string
	EqualFunc func(actual interface{}, expected interface{}) bool
)

type Trial struct {
	cases   map[string]Case
	testFn  TestFunc
	diffFn  DiffFunc
	equalFn EqualFunc
}
type Cases map[string]Case

type Case struct {
	Input    interface{}
	Expected interface{}

	// testing conditions
	ShouldErr   bool  // is an error expected
	ExpectedErr error // the error that was expected (nil is no error expected)
	ShouldPanic bool  // is a panic expected
}

func New(fn TestFunc, cases map[string]Case) *Trial {
	if cases == nil {
		cases = make(map[string]Case)
	}
	return &Trial{
		cases:   cases,
		testFn:  fn,
		diffFn:  Diff,
		equalFn: Equal,
	}
}

func (t *Trial) EqualFn(fn EqualFunc) *Trial {
	t.equalFn = fn
	return t
}

func (t *Trial) DiffFn(fn DiffFunc) *Trial {
	t.diffFn = fn
	return t
}

// Test all cases provided to trial
func (trial *Trial) Test(t testing.TB) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	for msg, test := range trial.cases {
		r := trial.testCase(msg, test)
		if r.Success {
			t.Log(r.Message)
		} else {
			t.Error("\033[31m" + r.Message + "\033[39m")
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
	} else if !test.ShouldErr && test.ExpectedErr == nil && !t.equalFn(result, test.Expected) {
		finished = true
		return fail("FAIL: %q differences %v", msg, t.diffFn(result, test.Expected))
	} else {
		finished = true
		return pass("PASS: %q", msg)
	}
}

// ContainsFn uses the strings.Contain method to compare two interfaces.
// both interfaces need to be strings or slice of strings.
func ContainsFn(actual, expected interface{}) bool {
	// if nothing is expected we have a match
	if expected == nil {
		return true
	}
	s1 := asStrings(actual)
	s2 := asStrings(expected)
	if s1 == nil {
		panic(fmt.Errorf("%s is not a string", reflect.TypeOf(actual)))
	}
	if s2 == nil {
		panic(fmt.Errorf("%s is not a string", reflect.TypeOf(expected)))
	}

	return containsSlice(s1, s2...)
}

func asStrings(i interface{}) []string {
	if s, ok := i.([]string); ok {
		return s
	}
	if s, ok := i.(string); ok {
		return []string{s}
	}
	return nil
}

func containsSlice(actual []string, expected ...string) bool {
	for _, s := range expected {
		found := false
		for _, act := range actual {
			if strings.Contains(act, s) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Equal uses the cmp.Equal method to compare two interfaces including unexported fields
func Equal(actual, expected interface{}) bool {
	t := reflect.TypeOf(actual)
	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		return false
	}

	if t == nil {
	} else if t.Kind() == reflect.Struct {
		return cmp.Equal(actual, expected, cmp.AllowUnexported(actual))
	} else if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		v := reflect.ValueOf(actual).Elem().Interface()
		return cmp.Equal(actual, expected, cmp.AllowUnexported(v))
	}
	return cmp.Equal(actual, expected)
}

// Diff use the cmp.Diff method to display differences between two interfaces
func Diff(actual, expected interface{}) string {
	var opts []cmp.Option
	if t := reflect.TypeOf(actual); t.Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(actual))
	} else if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(reflect.ValueOf(actual).Elem().Interface()))
	}
	if t := reflect.TypeOf(expected); t.Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(expected))
	} else if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		opts = append(opts, cmp.AllowUnexported(reflect.ValueOf(expected).Elem().Interface()))
	}
	return cmp.Diff(actual, expected, opts...)
}

// cleanStack removes unhelpful lines from a panic stack track
func cleanStack() (s string) {
	for _, ln := range strings.Split(string(debug.Stack()), "\n") {
		if !strings.Contains(ln, "/go-tools/trial") {
			s += ln + "\n"
		}
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
