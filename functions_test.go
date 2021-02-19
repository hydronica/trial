package trial

import (
	"errors"
	"testing"
	"time"
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
		t      time.Time
	}
	fn := func(in Input) (interface{}, error) {
		r, _ := Equal(in.Slice(0).Interface(), in.Slice(1).Interface())
		return r, nil
	}
	cases := Cases{"strings are equal": {
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
		"private members as part of map type": {
			Input: Args(
				map[string]test{"": {Public: 1, private: "a"}},
				map[string]test{"": {Public: 1, private: "a"}},
			),
			Expected: true,
		},
		"empty private map type": {
			Input: Args(
				map[string]test{},
				map[string]test{},
			),
			Expected: true,
		},
		"private key in map": {
			Input: Args(
				map[test]string{test{Public: 1, private: "a"}: "apple"},
				map[test]string{test{Public: 1, private: "a"}: "apple"},
			),
			Expected: true,
		},
		"map with *pointer struct": {
			Input: Args(
				map[string]grandparent{"a": {t: TimeHour("2018-01-01T00")}},
				map[string]grandparent{"a": {t: TimeHour("2018-01-01T00")}},
			),
			Expected: true,
		},
		"private slice": {
			Input: Args(
				[]test{{Public: 1, private: "a"}},
				[]test{{Public: 1, private: "a"}}),
			Expected: true,
		},
		"interface slice with private methods": {
			Input: Args(
				[]interface{}{test{Public: 1, private: "a"}, parent{}},
				[]interface{}{test{Public: 1, private: "a"}, parent{}},
			),
			Expected: true,
		},
		"private array": {
			Input: Args(
				[1]test{{Public: 1, private: "a"}},
				[1]test{{Public: 1, private: "a"}}),
			Expected: true,
		},
	}
	New(fn, cases).Test(t)
}

func TestContainsFn(t *testing.T) {
	New(func(in Input) (interface{}, error) {
		b, s := Contains(in.Slice(0).Interface(), in.Slice(1).Interface())
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
			Input:       Args("Hello World", "hello"),
			Expected:    false,
			ExpectedErr: errors.New("string ⊇ string\n + Hello World\n - hello"),
		},
		"type mismatch (string)": {
			Input:       Args("hello", 1),
			Expected:    false,
			ExpectedErr: errors.New("type mismatch string int"),
		},
		"type mismatch (map)": {
			Input:       Args(map[int]int{1: 1}, 1),
			Expected:    false,
			ExpectedErr: errors.New("type mismatch map[int]int int"),
		},
		"stringer type": {
			Input:    Args("hello", newMessagef("llo")),
			Expected: true,
		},
		"alias string": {
			Input:    Args(newMessagef("hello"), "ello"),
			Expected: true,
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
		"string with []string": {
			Input:    Args("the quick brown fox jumps over the lazy dog", []string{"quick", "fox", "dog"}),
			Expected: true,
		},
		"slice of ints": {
			Input:    Args([]int{12, 3, 5}, 3),
			Expected: true,
		},
		"[]int format check": {
			Input:       Args([]int{1, 2, 3}, []int{2, 3, 4, 5}),
			Expected:    false,
			ExpectedErr: errors.New("[]int ⊇ []int\n ∈ 2, 3\n - 4, 5"),
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
			Input: Args(
				map[string]interface{}{
					"int":     10,
					"float64": 1.1,
					"name":    "hello",
				},
				map[string]interface{}{"int": 10},
			),
			Expected: true,
		},
		"map[int]string": {
			Input:     Args(map[int]string{1: "a", 2: "b"}, map[int]string{1: "a", 2: "c", 3: "b"}),
			Expected:  false,
			ShouldErr: true,
		},
		"map[int][]string": {
			Input:     Args(map[int][]string{1: {"a", "b", "c"}, 2: {"d", "e", "f"}}, map[int][]string{3: {}, 1: {"b", "c", "d"}, 2: {"f"}}),
			Expected:  false,
			ShouldErr: true,
		},
		"map parent missing key": {
			Input:     Args(map[string]string{}, map[string]string{"test": "a"}),
			ShouldErr: true,
		},
	}).SubTest(t)
}

func TestCmpFuncs(t *testing.T) {
	fn := func(in Input) (interface{}, error) {
		_, s := CmpFuncs(in.Slice(0).Interface(), in.Slice(1).Interface())
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
			Input:    Args(Equal, Contains),
			Expected: "funcs not equal",
		},
	}).EqualFn(Contains).Test(t)
}
