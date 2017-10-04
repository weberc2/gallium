package ast

import (
	"strconv"
	"strings"
)

const indent = "    "

type FuncLit struct {
	Spec FuncSpec
	Body Expr
}

type TupleLit []Expr

func (fl FuncLit) RenderGo(t Type) string {
	return t.RenderGo() + " { return " + fl.Body.RenderGo() + " }"
}

func (fl FuncLit) Equal(other FuncLit) bool {
	return fl.Spec.Equal(other.Spec) && fl.Body.Equal(other.Body)
}

func (fl FuncLit) EqualExprNode(other ExprNode) bool {
	otherFuncLit, ok := other.(FuncLit)
	return ok && fl.Equal(otherFuncLit)
}

// func (fl FuncLit) TypeRefs() []TypeRef {
// 	return fl.Body.TypeRefs()
// }

type IntLit int

func (il IntLit) RenderGo(t Type) string {
	return strconv.Itoa(int(il))
}

func (il IntLit) EqualExprNode(other ExprNode) bool {
	otherIntLit, ok := other.(IntLit)
	return ok && il == otherIntLit
}

// func (il IntLit) TypeRefs() []TypeRef {
// 	return nil
// }

type StringLit string

func (sl StringLit) RenderGo(t Type) string {
	return "\"" + string(sl) + "\""
}

func (sl StringLit) EqualExprNode(other ExprNode) bool {
	otherStringLit, ok := other.(StringLit)
	return ok && sl == otherStringLit
}

// func (sl StringLit) TypeRefs() []TypeRef {
// 	return nil
// }

func (tl TupleLit) RenderGo(t Type) string {
	args := make([]string, len(tl))
	for i, expr := range tl {
		args[i] = "_" + strconv.Itoa(i) + ": " + expr.RenderGo()
	}
	return t.RenderGo() + " {" + strings.Join(args, ", ") + "}"
}

func (tl TupleLit) Equal(other TupleLit) bool {
	if len(tl) != len(other) {
		return false
	}
	for i, expr := range tl {
		if !expr.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (tl TupleLit) EqualExprNode(other ExprNode) bool {
	otherTupleLit, ok := other.(TupleLit)
	return ok && tl.Equal(otherTupleLit)
}

// func (tl TupleLit) TypeRefs() []TypeRef {
// 	var refs []TypeRef
// 	for _, expr := range tl {
// 		refs = append(refs, expr.TypeRefs()...)
// 	}
// 	return refs
// }

type Ident string

func (i Ident) RenderGo(t Type) string {
	return string(i)
}

func (i Ident) EqualExprNode(other ExprNode) bool {
	otherIdent, ok := other.(Ident)
	return ok && i == otherIdent
}

// func (i Ident) TypeRefs() []TypeRef {
// 	return nil
// }

type Call struct {
	Fn  Expr
	Arg Expr
}

func (c Call) RenderGo(t Type) string {
	return c.Fn.RenderGo() + "(" + c.Arg.RenderGo() + ")"
}

func (c Call) Equal(other Call) bool {
	return c.Fn.Equal(other.Fn) && c.Arg.Equal(other.Arg)
}

func (c Call) EqualExprNode(other ExprNode) bool {
	otherCall, ok := other.(Call)
	return ok && c.Equal(otherCall)
}

// func (c Call) TypeRefs() []TypeRef {
// 	return append(c.Fn.TypeRefs(), c.Arg.TypeRefs()...)
// }

type Block struct {
	Stmts []Stmt
	Expr  Expr
}

func (b Block) RenderGo(t Type) string {
	panic("Block.RenderGo() not yet implemented")
}

func (b Block) Visit(env ExprNodeVisitor) {
	env.VisitBlock(b)
}

func (b Block) Equal(other Block) bool {
	if len(b.Stmts) != len(other.Stmts) {
		return false
	}
	for i, stmt := range b.Stmts {
		if !stmt.EqualStmt(other.Stmts[i]) {
			return false
		}
	}
	return b.Expr.Equal(other.Expr)
}

func (b Block) EqualExprNode(other ExprNode) bool {
	otherBlock, ok := other.(Block)
	return ok && b.Equal(otherBlock)
}

type ExprNodeVisitor interface {
	VisitIntLit(IntLit)
	VisitStringLit(StringLit)
	VisitIdent(Ident)
	VisitTupleLit(TupleLit)
	VisitBlock(Block)
	VisitFuncLit(FuncLit)
	VisitCall(Call)
}

type ExprNode interface {
	RenderGo(t Type) string
	// TypeRefs() []TypeRef
	Visit(ExprNodeVisitor)
	EqualExprNode(ExprNode) bool
}

func (il IntLit) Visit(env ExprNodeVisitor) {
	env.VisitIntLit(il)
}

func (sl StringLit) Visit(env ExprNodeVisitor) {
	env.VisitStringLit(sl)
}

func (i Ident) Visit(env ExprNodeVisitor) {
	env.VisitIdent(i)
}

func (tl TupleLit) Visit(env ExprNodeVisitor) {
	env.VisitTupleLit(tl)
}

func (fl FuncLit) Visit(env ExprNodeVisitor) {
	env.VisitFuncLit(fl)
}

func (c Call) Visit(env ExprNodeVisitor) {
	env.VisitCall(c)
}

type Expr struct {
	Type Type
	Node ExprNode
}

// func (expr Expr) TypeRefs() []TypeRef {
// 	return append(expr.Type.TypeRefs(), expr.Node.TypeRefs()...)
// }

func (expr Expr) RenderGo() string {
	return expr.Node.RenderGo(expr.Type)
}

func (expr Expr) Equal(other Expr) bool {
	if expr.Type != nil {
		if !expr.Type.EqualType(other.Type) {
			return false
		}
	} else {
		if other.Type != nil {
			return false
		}
	}

	if expr.Node != nil {
		if !expr.Node.EqualExprNode(other.Node) {
			return false
		}
	} else {
		if other.Node != nil {
			return false
		}
	}

	return true
}

func (expr Expr) EqualNode(other Node) bool {
	otherExpr, ok := other.(Expr)
	return ok && expr.Equal(otherExpr)
}

func (expr Expr) EqualStmt(other Stmt) bool {
	otherExpr, ok := other.(Expr)
	return ok && expr.Equal(otherExpr)
}
