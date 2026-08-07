package trial

import (
	"errors"
	"strings"
	"testing"
)

type typedErr struct{ msg string }

func (e typedErr) Error() string { return e.msg }

// nilPanicsErr.Error dereferences the receiver; a typed-nil *nilPanicsErr panics on Error().
type nilPanicsErr struct{ msg string }

func (e *nilPanicsErr) Error() string { return "prefix: " + e.msg }

func typedNilPanicsErr() error {
	var err *nilPanicsErr
	return err
}

func TestError_matching(t *testing.T) {
	errFn := func(msg string) func(Input) (any, error) {
		return func(Input) (any, error) {
			return nil, errors.New(msg)
		}
	}
	typedNilFn := func(Input) (any, error) {
		return nil, typedNilPanicsErr()
	}

	cases := map[string]struct {
		fn        func(Input) (any, error)
		expected  error
		wantPass  bool
		wantInMsg string
	}{
		"any error pass": {
			fn:        errFn("anything"),
			expected:  Error(),
			wantPass:  true,
			wantInMsg: `PASS: "any error pass"`,
		},
		"any error fail when no error": {
			fn: func(Input) (any, error) {
				return nil, nil
			},
			expected:  Error(),
			wantPass:  false,
			wantInMsg: `should error`,
		},
		"exact pass": {
			fn:        errFn("test error"),
			expected:  Error().Exact("test error"),
			wantPass:  true,
			wantInMsg: `PASS: "exact pass"`,
		},
		"exact fail extra text": {
			fn:        errFn("test error: details"),
			expected:  Error().Exact("test error"),
			wantPass:  false,
			wantInMsg: `does not match exactly`,
		},
		"contains pass": {
			fn:        errFn("request timeout after 5s"),
			expected:  Error().Contains("timeout"),
			wantPass:  true,
			wantInMsg: `PASS: "contains pass"`,
		},
		"contains fail": {
			fn:        errFn("not found"),
			expected:  Error().Contains("timeout"),
			wantPass:  false,
			wantInMsg: `does not contain`,
		},
		"regex pass": {
			fn:        errFn("invalid xyz format"),
			expected:  Error().Regex(`invalid.*format`),
			wantPass:  true,
			wantInMsg: `PASS: "regex pass"`,
		},
		"regex fail": {
			fn:        errFn("bad format"),
			expected:  Error().Regex(`invalid.*format`),
			wantPass:  false,
			wantInMsg: `does not match pattern`,
		},
		"invalid regex pattern": {
			fn:        errFn("any error"),
			expected:  Error().Regex(`[invalid`),
			wantPass:  false,
			wantInMsg: `invalid regex pattern`,
		},
		"legacy errors.New substring": {
			fn:        errFn("test error"),
			expected:  errors.New("test error"),
			wantPass:  true,
			wantInMsg: `PASS: "legacy errors.New substring"`,
		},
		"typed-nil contains fails safely": {
			fn:        typedNilFn,
			expected:  Error().Contains("x"),
			wantPass:  false,
			wantInMsg: "<nil",
		},
		"typed-nil exact fails safely": {
			fn:        typedNilFn,
			expected:  Error().Exact("x"),
			wantPass:  false,
			wantInMsg: "<nil",
		},
		"typed-nil legacy errors.New fails safely": {
			fn:        typedNilFn,
			expected:  errors.New("x"),
			wantPass:  false,
			wantInMsg: "<nil",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tr := New(tc.fn, nil)
			r := tr.testCase(name, Case[Input, any]{ExpectedErr: tc.expected})
			if r.Success != tc.wantPass {
				t.Fatalf("Success = %v, want %v; message: %s", r.Success, tc.wantPass, r.Message)
			}
			if !strings.Contains(r.Message, tc.wantInMsg) {
				t.Fatalf("message %q does not contain %q", r.Message, tc.wantInMsg)
			}
		})
	}
}

func TestError_IsType_matching(t *testing.T) {
	typedNilFn := func(Input) (any, error) {
		return nil, typedNilPanicsErr()
	}

	cases := map[string]struct {
		fn        func(Input) (any, error)
		expected  error
		wantPass  bool
		wantInMsg string
	}{
		"type only pass": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "anything"}
			},
			expected:  Error().IsType(typedErr{}),
			wantPass:  true,
			wantInMsg: `PASS: "type only pass"`,
		},
		"type only fail wrong type": {
			fn: func(Input) (any, error) {
				return nil, errors.New("plain")
			},
			expected:  Error().IsType(typedErr{}),
			wantPass:  false,
			wantInMsg: `is not trial.typedErr`,
		},
		"typed-nil is type mismatch fails safely": {
			fn:        typedNilFn,
			expected:  Error().IsType(typedErr{}),
			wantPass:  false,
			wantInMsg: "<nil",
		},
		"type and contains pass": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "validation failed: field required"}
			},
			expected:  Error().IsType(typedErr{}).Contains("field required"),
			wantPass:  true,
			wantInMsg: `PASS: "type and contains pass"`,
		},
		"type and contains fail wrong message": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "validation failed"}
			},
			expected:  Error().IsType(typedErr{}).Contains("field required"),
			wantPass:  false,
			wantInMsg: `does not contain`,
		},
		"type and contains fail wrong type": {
			fn: func(Input) (any, error) {
				return nil, errors.New("field required")
			},
			expected:  Error().IsType(typedErr{}).Contains("field required"),
			wantPass:  false,
			wantInMsg: `is not trial.typedErr`,
		},
		"type and regex pass": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "field abc required"}
			},
			expected:  Error().IsType(typedErr{}).Regex(`field .+ required`),
			wantPass:  true,
			wantInMsg: `PASS: "type and regex pass"`,
		},
		"type and regex fail": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "field missing"}
			},
			expected:  Error().IsType(typedErr{}).Regex(`field .+ required`),
			wantPass:  false,
			wantInMsg: `does not match pattern`,
		},
		"type and invalid regex": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "field required"}
			},
			expected:  Error().IsType(typedErr{}).Regex(`[invalid`),
			wantPass:  false,
			wantInMsg: `invalid regex pattern`,
		},
		"deprecated ErrType still works": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "anything"}
			},
			expected:  ErrType(typedErr{}),
			wantPass:  true,
			wantInMsg: `PASS: "deprecated ErrType still works"`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tr := New(tc.fn, nil)
			r := tr.testCase(name, Case[Input, any]{ExpectedErr: tc.expected})
			if r.Success != tc.wantPass {
				t.Fatalf("Success = %v, want %v; message: %s", r.Success, tc.wantPass, r.Message)
			}
			if !strings.Contains(r.Message, tc.wantInMsg) {
				t.Fatalf("message %q does not contain %q", r.Message, tc.wantInMsg)
			}
		})
	}
}

func TestExpectErr_Error_string(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "any error",
			err:  Error(),
			want: "any error",
		},
		{
			name: "exact",
			err:  Error().Exact("timeout"),
			want: `error exactly "timeout"`,
		},
		{
			name: "contains",
			err:  Error().Contains("timeout"),
			want: `error containing "timeout"`,
		},
		{
			name: "regex",
			err:  Error().Regex(`time.*out`),
			want: `error matching pattern "time.*out"`,
		},
		{
			name: "is type only",
			err:  Error().IsType(typedErr{}),
			want: "error of type trial.typedErr",
		},
		{
			name: "is type only with sample error message",
			err:  Error().IsType(errors.New("not found")),
			want: "error of type *errors.errorString",
		},
		{
			name: "is type contains",
			err:  Error().IsType(typedErr{}).Contains("required"),
			want: `error of type trial.typedErr containing "required"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}
