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

// Error returns an expected error matcher that compares the full error message
// by default. Use Contains or Regex to change the comparison mode.
func Error(text string) expectErr {
	return expectErr{text: text, mode: matchExact}
}

// ErrType can be used with ExpectedErr to check that the expected err is of a
// certain type. Use Contains or Regex to also match the error message.
func ErrType(err error) expectErr {
	return expectErr{err: err}
}

// Contains sets substring matching on the error message. When called on
// Error(text), the text argument is reused. When called on ErrType, a substring
// argument is required.
func (e expectErr) Contains(substr ...string) error {
	if len(substr) > 0 {
		e.text = substr[0]
	}
	e.mode = matchContains
	return e
}

// Regex sets regex matching on the error message. When called on Error(text),
// the text argument is reused as the pattern. When called on ErrType, a pattern
// argument is required. Invalid patterns fail at compare time with a clear
// error message.
func (e expectErr) Regex(pattern ...string) error {
	if len(pattern) > 0 {
		e.text = pattern[0]
	}
	e.mode = matchRegex
	return e
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
			return ""
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
