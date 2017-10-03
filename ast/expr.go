package ast

import (
	"strconv"
	"strings"
)

const indent = "    "

type StringLit string

type IntLit int

type FuncLit struct {
	Body Expr
}

type TupleLit []Expr

func (fl FuncLit) exprNode()   {}
func (il IntLit) exprNode()    {}
func (sl StringLit) exprNode() {}
func (tl TupleLit) exprNode()  {}

func (fl FuncLit) RenderGo(t Type) string {
	return t.RenderGo() + " { return " + fl.Body.RenderGo() + " }"
}

func (fl FuncLit) TypeRefs() []TypeRef {
	return fl.Body.TypeRefs()
}

func (il IntLit) RenderGo(t Type) string {
	return strconv.Itoa(int(il))
}

func (il IntLit) TypeRefs() []TypeRef {
	return nil
}

func (sl StringLit) RenderGo(t Type) string {
	return "\"" + string(sl) + "\""
}

func (sl StringLit) TypeRefs() []TypeRef {
	return nil
}

func (tl TupleLit) RenderGo(t Type) string {
	args := make([]string, len(tl))
	for i, expr := range tl {
		args[i] = "_" + strconv.Itoa(i) + ": " + expr.RenderGo()
	}
	return t.RenderGo() + " {" + strings.Join(args, ", ") + "}"
}

func (tl TupleLit) TypeRefs() []TypeRef {
	var refs []TypeRef
	for _, expr := range tl {
		refs = append(refs, expr.TypeRefs()...)
	}
	return refs
}

type Ident string

type Call struct {
	Fn   Expr
	Args []Expr
}

func (c Call) exprNode()  {}
func (i Ident) exprNode() {}

func (c Call) RenderGo(t Type) string {
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.RenderGo()
	}
	return c.Fn.RenderGo() + "(" + strings.Join(args, ", ") + ")"
}

func (c Call) TypeRefs() []TypeRef {
	refs := c.Fn.TypeRefs()
	for _, arg := range c.Args {
		refs = append(refs, arg.TypeRefs()...)
	}
	return refs
}

func (i Ident) RenderGo(t Type) string {
	return string(i)
}

func (i Ident) TypeRefs() []TypeRef {
	return nil
}

type Expr struct {
	Type Type
	m    interface {
		RenderGo(t Type) string
		TypeRefs() []TypeRef
		exprNode()
	}
}

func (expr Expr) TypeRefs() []TypeRef {
	return append(expr.Type.TypeRefs(), expr.m.TypeRefs()...)
}

func (expr Expr) RenderGo() string {
	return expr.m.RenderGo(expr.Type)
}
