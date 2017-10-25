package infer

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/weberc2/gallium/ast"
)

func TestApply(t *testing.T) {
	testCases := []struct {
		Name      string
		InputExpr ast.Expr
		InputSubs []Substitution
		Wanted    ast.Expr
	}{
		{
			Name: "one-sub-w-two-type-vars",
			InputExpr: ast.Expr{
				Type: ast.FuncSpec{
					Arg: ast.TypeVar("a"),
					Ret: ast.TypeVar("b"),
				},
				Node: ast.FuncLit{
					Arg: "x",
					Body: ast.Expr{
						Type: ast.TypeVar("a"),
						Node: ast.Ident("x"),
					},
				},
			},
			InputSubs: []Substitution{{"a", ast.TypeVar("b")}},
			Wanted: ast.Expr{
				Type: ast.FuncSpec{
					Arg: ast.TypeVar("b"),
					Ret: ast.TypeVar("b"),
				},
				Node: ast.FuncLit{
					Arg: "x",
					Body: ast.Expr{
						Type: ast.TypeVar("b"),
						Node: ast.Ident("x"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := ApplyExpr(testCase.InputSubs, testCase.InputExpr)
			if !got.Equal(testCase.Wanted) {
				t.Fatalf(
					"WANTED:\n%# v\n\nGOT:\n%# v",
					pretty.Formatter(testCase.Wanted),
					pretty.Formatter(got),
				)
			}
		})
	}
}
