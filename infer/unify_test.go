package infer

import (
	"github.com/weberc2/gallium/ast"
	"testing"

	"github.com/kr/pretty"
)

func TestUnifyOne(t *testing.T) {
	testCases := []struct {
		Name      string
		Input     Constraint
		Wanted    []Substitution
		WantedErr bool
	}{
		{
			Name:   "two-matching-primitives",
			Input:  Constraint{ast.Primitive("int"), ast.Primitive("int")},
			Wanted: nil,
		},
		{
			Name: "mismatched-primitives",
			Input: Constraint{
				ast.Primitive("int"),
				ast.Primitive("string"),
			},
			WantedErr: true,
		},
		{
			Name:  "primitive-and-typevar",
			Input: Constraint{ast.TypeVar("a"), ast.Primitive("int")},
			Wanted: []Substitution{{
				Var:  ast.TypeVar("a"),
				Type: ast.Primitive("int"),
			}},
		},
		{
			Name:   "matching-typevars",
			Input:  Constraint{ast.TypeVar("a"), ast.TypeVar("a")},
			Wanted: nil,
		},
		{
			Name:  "typevar-and-primitive",
			Input: Constraint{ast.Primitive("int"), ast.TypeVar("a")},
			Wanted: []Substitution{{
				Var:  ast.TypeVar("a"),
				Type: ast.Primitive("int"),
			}},
		},
		{
			Name: "identical-concrete-fns",
			Input: Constraint{
				ast.FuncSpec{ast.Primitive("int"), ast.Primitive("string")},
				ast.FuncSpec{ast.Primitive("int"), ast.Primitive("string")},
			},
			Wanted: nil,
		},
		{
			Name: "identical-generic-fns",
			Input: Constraint{
				ast.FuncSpec{ast.TypeVar("a"), ast.Primitive("string")},
				ast.FuncSpec{ast.TypeVar("a"), ast.Primitive("string")},
			},
			Wanted: nil,
		},
		{
			Name: "one-generic-fn-and-one-concrete-fn",
			Input: Constraint{
				ast.FuncSpec{ast.TypeVar("a"), ast.TypeVar("b")},
				ast.FuncSpec{ast.Primitive("int"), ast.Primitive("int")},
			},
			Wanted: []Substitution{
				{ast.TypeVar("a"), ast.Primitive("int")},
				{ast.TypeVar("b"), ast.Primitive("int")},
			},
		},
		{
			Name: "one-concrete-fn-and-one-generic-fn",
			Input: Constraint{
				ast.FuncSpec{ast.Primitive("int"), ast.Primitive("int")},
				ast.FuncSpec{ast.TypeVar("a"), ast.TypeVar("b")},
			},
			Wanted: []Substitution{
				{ast.TypeVar("a"), ast.Primitive("int")},
				{ast.TypeVar("b"), ast.Primitive("int")},
			},
		},
		{
			Name: "identical-tuple-specs",
			Input: Constraint{
				ast.TupleSpec{ast.Primitive("int"), ast.Primitive("int")},
				ast.TupleSpec{ast.Primitive("int"), ast.Primitive("int")},
			},
			Wanted: nil,
		},
		{
			Name: "tuple-specs-equal-length-mismatched-types",
			Input: Constraint{
				ast.TupleSpec{ast.Primitive("string")},
				ast.TupleSpec{ast.Primitive("int")},
			},
			WantedErr: true,
		},
		{
			Name: "tuple-specs-mismatched-length",
			Input: Constraint{
				ast.TupleSpec{ast.Primitive("string"), ast.Primitive("int")},
				ast.TupleSpec{ast.Primitive("string")},
			},
			WantedErr: true,
		},
		{
			Name: "tuple-specs-identical-generic",
			Input: Constraint{
				ast.TupleSpec{ast.TypeVar("a")},
				ast.TupleSpec{ast.TypeVar("a")},
			},
			Wanted: nil,
		},
		{
			Name: "tuple-specs-one-generic-one-concrete",
			Input: Constraint{
				ast.TupleSpec{ast.TypeVar("a")},
				ast.TupleSpec{ast.Primitive("int")},
			},
			Wanted: []Substitution{{ast.TypeVar("a"), ast.Primitive("int")}},
		},
		{
			Name: "tuple-specs-one-concrete-one-generic",
			Input: Constraint{
				ast.TupleSpec{ast.Primitive("int")},
				ast.TupleSpec{ast.TypeVar("a")},
			},
			Wanted: []Substitution{{ast.TypeVar("a"), ast.Primitive("int")}},
		},
		{
			Name:   "one-typevar-and-one-tuple-spec",
			Input:  Constraint{ast.TypeVar("a"), ast.TupleSpec{}},
			Wanted: []Substitution{{ast.TypeVar("a"), ast.TupleSpec{}}},
		},
		{
			Name:   "one-tuple-spec-and-one-typevar",
			Input:  Constraint{ast.TupleSpec{}, ast.TypeVar("a")},
			Wanted: []Substitution{{ast.TypeVar("a"), ast.TupleSpec{}}},
		},
		{
			Name: "one-fn-and-one-typevar",
			Input: Constraint{
				ast.FuncSpec{ast.TupleSpec{}, ast.TupleSpec{}},
				ast.TypeVar("a"),
			},
			Wanted: []Substitution{{
				ast.TypeVar("a"),
				ast.FuncSpec{ast.TupleSpec{}, ast.TupleSpec{}},
			}},
		},
		{
			Name: "one-typevar-and-one-fn",
			Input: Constraint{
				ast.TypeVar("a"),
				ast.FuncSpec{ast.TupleSpec{}, ast.TupleSpec{}},
			},
			Wanted: []Substitution{{
				ast.TypeVar("a"),
				ast.FuncSpec{ast.TupleSpec{}, ast.TupleSpec{}},
			}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got, err := UnifyOne(testCase.Input.L, testCase.Input.R)
			if err != nil {
				if testCase.WantedErr {
					return
				}
				t.Fatal("Unexpected error:", err)
			}

			if len(testCase.Wanted) != len(got) {
				t.Fatalf(
					"WANTED:\n%# v\n\nGOT:\n%# v",
					pretty.Formatter(testCase.Wanted),
					pretty.Formatter(got),
				)
			}

			for i := range got {
				if !got[i].Equal(testCase.Wanted[i]) {
					t.Fatalf(
						"WANTED:\n%# v\n\nGOT:\n%# v",
						pretty.Formatter(testCase.Wanted[i]),
						pretty.Formatter(got[i]),
					)
				}
			}
		})
	}
}
