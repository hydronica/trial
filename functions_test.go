package trial

import (
	"errors"
	"testing"
)

func TestEqualFn(t *testing.T) {
	type test struct {
		Public  int
		private string
	}
	type parent struct {
		child test
	}
	type grandparent struct {
		parent *parent
	}
	New(func(args ...interface{}) (interface{}, error) {
		r, _ := Equal(args[0], args[1])
		return r, nil
	}, map[string]Case{
		"strings are equal": {
			Input:    Args("hello", "hello"),
			Expected: true,
		},
		"ints not equal": {
			Input:    Args(1, 2),
			Expected: false,
		},
		"compare private methods": {
			Input:    Args(test{Public: 1, private: "a"}, test{Public: 1, private: "a"}),
			Expected: true,
		},
		"compare private structs with private methods": {
			Input:    Args(parent{child: test{Public: 1, private: "a"}}, parent{child: test{Public: 1, private: "a"}}),
			Expected: true,
		},
		"multi-depth private struct with pointer": {
			Input: Args(grandparent{
				parent: &parent{
					child: test{Public: 1, private: "a"},
				},
			}, grandparent{
				parent: &parent{
					child: test{Public: 1, private: "a"},
				},
			}),
			Expected: true,
		},
		"private method pointer": {
			Input:    Args(&test{Public: 1, private: "a"}, &test{Public: 1, private: "a"}),
			Expected: true,
		},
		"nils don't panic": {
			Input:    Args(nil, nil),
			Expected: true,
		},
	}).Test(t)
}

func TestContainsFn(t *testing.T) {
	New(func(args ...interface{}) (interface{}, error) {
		b, s := ContainsFn(args[0], args[1])
		var err error
		if s != "" {
			err = errors.New(s)
		}
		return b, err
	}, map[string]Case{
		"blank string matches anything": {
			Input:    Args("Hello world", ""),
			Expected: true,
		},
		"case sensitive": {
			Input:     Args("Hello World", "hello"),
			Expected:  false,
			ShouldErr: true,
		},
		"nil matches everything": {
			Input:    Args("hello world", nil),
			Expected: true,
		},
		"match substring": {
			Input:    Args("abcdefghijklmnopqrstuvwxyz", "lmnop"),
			Expected: true,
		},
		"match full string": {
			Input:    Args("abcdefghijklmnopqrstuvwxyz", "abcdefghijklmnopqrstuvwxyz"),
			Expected: true,
		},
		"slice of strings": {
			Input:    Args([]string{"hello", "world"}, "world"),
			Expected: true,
		},
		"array of strings": {
			Input:    Args([2]string{"hello", "world"}, "world"),
			Expected: true,
		},
		"slice of ints": {
			Input:    Args([]int{12, 3, 5}, 3),
			Expected: true,
		},
		"slice of different types": {
			Input:     Args([]int{1, 2, 3}, []float32{1.1}),
			Expected:  false,
			ShouldErr: true,
		},
		"empty slice": {
			Input:     Args([]int{}, 1),
			Expected:  false,
			ShouldErr: true,
		},
		"array of floats": {
			Input:    Args([3]float64{1.1, 2.2, 3.3}, 2.2),
			Expected: true,
		},
		"array of different type": {
			Input:     Args([2]int{1, 2}, "hello"),
			Expected:  false,
			ShouldErr: true,
		},
		"[]interface{}": {
			Input:    Args([]interface{}{1, 2, 3, "abc", 4.5}, []interface{}{2, "abc"}),
			Expected: true,
		},
		"[]interface{} with int slice": {
			Input:    Args([]interface{}{1, 2, 3, "abc", 4.5}, []int{2, 1}),
			Expected: true,
		},
		"expected is slice subset of actual": {
			Input:    Args([]string{"the", "quick", "brown", "fox"}, []string{"fox", "quick"}),
			Expected: true,
		},
		"expected is array subset of actual": {
			Input:    Args([4]string{"the", "quick", "brown", "fox"}, [2]string{"fox", "quick"}),
			Expected: true,
		},
		"partial match of string slices": {
			Input:    Args([]string{"abcdefghijklmnop", "qrstuvwxyz"}, []string{"abc", "def"}),
			Expected: true,
		},
		"map[string]interface{}": {
			Input: Args(map[string]interface{}{
				"int":     10,
				"float64": 1.1,
				"name":    "hello",
			},
				map[string]interface{}{"int": 10},
			),
			Expected: true,
		},
	}).Test(t)
}

func TestCmpFuncs(t *testing.T) {
	fn := func(args ...interface{}) (interface{}, error) {
		_, s := CmpFuncs(args[0], args[1])
		return s, nil
	}
	New(fn, Cases{
		"x & y are nil": {
			Input:    Args(nil, nil),
			Expected: "",
		},
		"only y is nil": {
			Input:    Args(10, nil),
			Expected: "10 != <nil>",
		},
		"handle non-function input": {
			Input:    Args(1, 2),
			Expected: "can only compare functions",
		},
		"identical functions": {
			Input:    Args(Equal, Equal),
			Expected: "",
		},
		"non-identical functions": {
			Input:    Args(Equal, ContainsFn),
			Expected: "funcs not equal",
		},
	}).EqualFn(ContainsFn).Test(t)
}
