package trial

import (
	"errors"
	"strings"
	"testing"
)

func TestError_matching(t *testing.T) {
	errFn := func(msg string) func(Input) (any, error) {
		return func(Input) (any, error) {
			return nil, errors.New(msg)
		}
	}

	cases := map[string]struct {
		fn        func(Input) (any, error)
		expected  error
		wantPass  bool
		wantInMsg string
	}{
		"exact pass": {
			fn:        errFn("test error"),
			expected:  Error("test error"),
			wantPass:  true,
			wantInMsg: `PASS: "exact pass"`,
		},
		"exact fail extra text": {
			fn:        errFn("test error: details"),
			expected:  Error("test error"),
			wantPass:  false,
			wantInMsg: `does not match exactly`,
		},
		"contains pass from Error": {
			fn:        errFn("request timeout after 5s"),
			expected:  Error("timeout").Contains(),
			wantPass:  true,
			wantInMsg: `PASS: "contains pass from Error"`,
		},
		"contains fail": {
			fn:        errFn("not found"),
			expected:  Error("timeout").Contains(),
			wantPass:  false,
			wantInMsg: `does not contain`,
		},
		"regex pass from Error": {
			fn:        errFn("invalid xyz format"),
			expected:  Error(`invalid.*format`).Regex(),
			wantPass:  true,
			wantInMsg: `PASS: "regex pass from Error"`,
		},
		"regex fail": {
			fn:        errFn("bad format"),
			expected:  Error(`invalid.*format`).Regex(),
			wantPass:  false,
			wantInMsg: `does not match pattern`,
		},
		"invalid regex pattern": {
			fn:        errFn("any error"),
			expected:  Error(`[invalid`).Regex(),
			wantPass:  false,
			wantInMsg: `invalid regex pattern`,
		},
		"legacy errors.New substring": {
			fn:        errFn("test error"),
			expected:  errors.New("test error"),
			wantPass:  true,
			wantInMsg: `PASS: "legacy errors.New substring"`,
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

type typedErr struct{ msg string }

func (e typedErr) Error() string { return e.msg }

func TestErrType_matching(t *testing.T) {
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
			expected:  ErrType(typedErr{}),
			wantPass:  true,
			wantInMsg: `PASS: "type only pass"`,
		},
		"type only fail wrong type": {
			fn: func(Input) (any, error) {
				return nil, errors.New("plain")
			},
			expected:  ErrType(typedErr{}),
			wantPass:  false,
			wantInMsg: `is not trial.typedErr`,
		},
		"type and contains pass": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "validation failed: field required"}
			},
			expected:  ErrType(typedErr{}).Contains("field required"),
			wantPass:  true,
			wantInMsg: `PASS: "type and contains pass"`,
		},
		"type and contains fail wrong message": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "validation failed"}
			},
			expected:  ErrType(typedErr{}).Contains("field required"),
			wantPass:  false,
			wantInMsg: `does not contain`,
		},
		"type and contains fail wrong type": {
			fn: func(Input) (any, error) {
				return nil, errors.New("field required")
			},
			expected:  ErrType(typedErr{}).Contains("field required"),
			wantPass:  false,
			wantInMsg: `is not trial.typedErr`,
		},
		"type and regex pass": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "field abc required"}
			},
			expected:  ErrType(typedErr{}).Regex(`field .+ required`),
			wantPass:  true,
			wantInMsg: `PASS: "type and regex pass"`,
		},
		"type and regex fail": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "field missing"}
			},
			expected:  ErrType(typedErr{}).Regex(`field .+ required`),
			wantPass:  false,
			wantInMsg: `does not match pattern`,
		},
		"type and invalid regex": {
			fn: func(Input) (any, error) {
				return nil, typedErr{msg: "field required"}
			},
			expected:  ErrType(typedErr{}).Regex(`[invalid`),
			wantPass:  false,
			wantInMsg: `invalid regex pattern`,
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
			name: "exact",
			err:  Error("timeout"),
			want: `error exactly "timeout"`,
		},
		{
			name: "contains",
			err:  Error("timeout").Contains(),
			want: `error containing "timeout"`,
		},
		{
			name: "regex",
			err:  Error(`time.*out`).Regex(),
			want: `error matching pattern "time.*out"`,
		},
		{
			name: "type only",
			err:  ErrType(typedErr{}),
			want: "error of type trial.typedErr",
		},
		{
			name: "type with error message",
			err:  ErrType(errors.New("not found")),
			want: "not found",
		},
		{
			name: "type contains",
			err:  ErrType(typedErr{}).Contains("required"),
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
