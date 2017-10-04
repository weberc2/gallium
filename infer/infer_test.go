package infer

import (
	"testing"
)

func TestInfer(t *testing.T) {
	testCases := []struct {
		Name       string
		Input      ast.Expr
		WantedType ast.Type
		WantedErr  bool
	}{
		{
			Name:       "primitive-int",
			Input:      Expr{M: ast.IntLit(0)},
			WantedType: ast.Primitive("int"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got, err := Infer(testCase.Input)
			if err != nil {
				if !testCase.WantedErr {
					t.Fatal("Unexpected error:", err)
				}
			}
			if !testCase.WantedType.EqualType(got.Type) {
				t.Fatalf(
					"Wanted:\n%# v\n\nGot:\n%# v",
					pretty.Formatter(testCase.WantedType),
					pretty.Formatter(got.Type),
				)
			}
		})
	}
}
