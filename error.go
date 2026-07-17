package trial

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type matchMode int

const (
	matchExact matchMode = iota
	matchContains
	matchRegex
)

type expectErr struct {
	err  error
	text string
	mode matchMode
}

// Error returns an expected error matcher. Use builder methods to constrain the
// match: IsType, Exact, Contains, or Regex. With no methods, any error matches.
func Error() expectErr {
	return expectErr{}
}

// IsType requires the actual error to be the same type as err.
func (e expectErr) IsType(err error) expectErr {
	e.err = err
	return e
}

// Exact requires the error message to match exactly.
func (e expectErr) Exact(msg string) error {
	e.text = msg
	e.mode = matchExact
	return e
}

// Contains requires the error message to contain substr.
func (e expectErr) Contains(substr string) error {
	e.text = substr
	e.mode = matchContains
	return e
}

// Regex requires the error message to match pattern. Invalid patterns fail at
// compare time with a clear error message.
func (e expectErr) Regex(pattern string) error {
	e.text = pattern
	e.mode = matchRegex
	return e
}

// ErrType is deprecated. Use Error().IsType(err) instead.
func ErrType(err error) error {
	return Error().IsType(err)
}

func (e expectErr) modeWord() string {
	switch e.mode {
	case matchContains:
		return "containing"
	case matchRegex:
		return "matching pattern"
	default:
		return "exactly"
	}
}

func (e expectErr) Error() string {
	if e.text == "" {
		if e.err == nil {
			return "any error"
		}
		if msg := e.err.Error(); msg != "" {
			return msg
		}
		return fmt.Sprintf("error of type %T", e.err)
	}
	desc := fmt.Sprintf("%s %q", e.modeWord(), e.text)
	if e.err != nil {
		return fmt.Sprintf("error of type %T %s", e.err, desc)
	}
	return "error " + desc
}

func (e expectErr) matches(actual error) (matched bool, failMsg string) {
	if e.err != nil {
		if reflect.TypeOf(actual) != reflect.TypeOf(e.err) {
			if e.text == "" {
				return false, fmt.Sprintf("error %q is not %T", actual, e.err)
			}
			switch e.mode {
			case matchContains:
				return false, fmt.Sprintf("error %q is not %T (expected to contain %q)", actual, e.err, e.text)
			case matchRegex:
				return false, fmt.Sprintf("error %q is not %T (expected to match pattern %q)", actual, e.err, e.text)
			default:
				return false, fmt.Sprintf("error %q is not %T (expected exactly %q)", actual, e.err, e.text)
			}
		}
	}

	if e.text == "" {
		return true, ""
	}

	actualMsg := actual.Error()
	switch e.mode {
	case matchContains:
		if strings.Contains(actualMsg, e.text) {
			return true, ""
		}
		return false, fmt.Sprintf("error %q does not contain %q", actual, e.text)
	case matchRegex:
		re, err := regexp.Compile(e.text)
		if err != nil {
			return false, fmt.Sprintf("invalid regex pattern %q: %v", e.text, err)
		}
		if re.MatchString(actualMsg) {
			return true, ""
		}
		return false, fmt.Sprintf("error %q does not match pattern %q", actual, e.text)
	default:
		if actualMsg == e.text {
			return true, ""
		}
		return false, fmt.Sprintf("error %q does not match exactly %q", actual, e.text)
	}
}

func isExpectedError(actual, expected error) (matched bool, failMsg string) {
	if e, ok := expected.(expectErr); ok {
		return e.matches(actual)
	}
	if strings.Contains(actual.Error(), expected.Error()) {
		return true, ""
	}
	return false, ""
}
