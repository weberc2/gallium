package infer

import (
	"github.com/weberc2/gallium/ast"
)

type typeSet map[ast.Type]struct{}{}

func (ts typeSet) add(t ast.Type) {
	ts[t] = struct{}{}
}

type visitor struct {
	bindings map[ast.Ident]typeSet
}

func (v visitor) VisitPrimitive(p Primitive) {
}

func Infer(expr ast.Expr) ast.Expr {
	switch expr.(type)
}
