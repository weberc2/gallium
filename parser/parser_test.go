package parser

import (
	"github.com/weberc2/gallium/ast"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		Name        string
		Input       Input
		WantedRest  Input
		WantedValue interface{}
		WantedErr   bool
		Parser      Parser
	}{
		{
			Name:        "ws",
			Input:       " \t\n",
			WantedValue: " \t\n",
			Parser:      WS,
		},
		{
			Name:        "can-ws",
			Input:       " \t\n",
			WantedValue: " \t\n",
			Parser:      CanWS,
		},
		{
			Name:        "func-spec-simple",
			Input:       "fn()",
			WantedValue: ast.FuncSpec{},
			Parser:      FuncSpec,
		},
		{
			Name:        "func-spec-no-args-w-ret",
			Input:       "fn() -> int",
			WantedValue: ast.FuncSpec{Ret: ast.TypeRef{Name: "int"}},
			Parser:      FuncSpec,
		},
		{
			Name:  "func-spec-one-arg-no-ret",
			Input: "fn(i int)",
			WantedValue: ast.FuncSpec{Args: []ast.ArgSpec{
				{"i", ast.TypeRef{Name: "int"}},
			}},
			Parser: FuncSpec,
		},
		{
			Name:  "func-spec-multi-args-no-ret",
			Input: "fn(i int, j int)",
			WantedValue: ast.FuncSpec{Args: []ast.ArgSpec{
				{"i", ast.TypeRef{Name: "int"}},
				{"j", ast.TypeRef{Name: "int"}},
			}},
			Parser: FuncSpec,
		},
		{
			Name:  "func-spec-multi-args-w-ret",
			Input: "fn(i int, j int) -> bool",
			WantedValue: ast.FuncSpec{
				Args: []ast.ArgSpec{
					{"i", ast.TypeRef{Name: "int"}},
					{"j", ast.TypeRef{Name: "int"}},
				},
				Ret: ast.TypeRef{Name: "bool"},
			},
			Parser: FuncSpec,
		},
		{
			Name:  "type-decl-simple",
			Input: "type foo = int;",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TypeRef{Name: "int"},
			},
			Parser: TypeDecl,
		},
		{
			Name:  "type-decl-generic",
			Input: "type foo a b = bar a b;",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TypeRef{
					Name: "bar",
					Arg:  ast.TypeRef{Name: "a", Arg: ast.TypeRef{Name: "b"}},
				},
				Args: []ast.TypeVar{"a", "b"},
			},
			Parser: TypeDecl,
		},
		{
			Name:  "type-decl-generic-tuple",
			Input: "type foo a b = (int, a, b);",
			WantedValue: ast.TypeDecl{
				Name: "foo",
				Type: ast.TupleSpec{
					ast.TypeRef{Name: "int"},
					ast.TypeRef{Name: "a"},
					ast.TypeRef{Name: "b"},
				},
				Args: []ast.TypeVar{"a", "b"},
			},
			Parser: TypeDecl,
		},
		{
			Name:  "type-decl-simple-function",
			Input: "type Parser = fn(input Input) -> Result;",
			WantedValue: ast.TypeDecl{
				Name: "Parser",
				Type: ast.FuncSpec{
					Args: []ast.ArgSpec{
						{Name: "input", Type: ast.TypeRef{Name: "Input"}},
					},
					Ret: ast.TypeRef{Name: "Result"},
				},
			},
			Parser: TypeDecl,
		},
		{
			Name:        "file-empty",
			Input:       "package main",
			WantedValue: ast.File{Package: "main"},
			Parser:      File,
		},
		{
			Name:        "file-empty-trailing-whitespace",
			Input:       "package main\n",
			WantedValue: ast.File{Package: "main"},
			Parser:      File,
		},
		{
			Name: "file-w-lone-type-decl",
			Input: `package main

					type foo = int;`,
			WantedValue: ast.File{
				Package: "main",
				Decls: []ast.Decl{ast.TypeDecl{
					Name: "foo",
					Type: ast.TypeRef{Name: "int"},
				}},
			},
			Parser: File,
		},
		{
			Name: "file-w-many-type-decls",
			Input: `package main

					type x = foo;

					type y = bar;`,
			WantedValue: ast.File{
				Package: "main",
				Decls: []ast.Decl{
					ast.TypeDecl{Name: "x", Type: ast.TypeRef{Name: "foo"}},
					ast.TypeDecl{Name: "y", Type: ast.TypeRef{Name: "bar"}},
				},
			},
			Parser: File,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := testCase.Parser(testCase.Input)
			if testCase.WantedErr && result.Err == nil {
				t.Fatal("Wanted an error but didn't get any")
			}
			if !testCase.WantedErr && result.Err != nil {
				t.Fatal("Unexpected error:", result)
			}

			if testCase.WantedRest != result.Rest {
				t.Fatalf(
					"Wanted REST: %#v; got REST: %#v",
					testCase.WantedRest,
					result.Rest,
				)
			}

			if wanted, ok := testCase.WantedValue.(ast.Node); ok {
				if got, ok := result.Value.(ast.Node); ok {
					if wanted.EqualNode(got) {
						return
					}
				}
			} else {
				if reflect.DeepEqual(result.Value, testCase.WantedValue) {
					return
				}
			}

			t.Fatalf(
				"Wanted:\n%# v\n\nGot:\n%# v\n",
				pretty.Formatter(testCase.WantedValue),
				pretty.Formatter(result.Value),
			)
		})
	}
}
