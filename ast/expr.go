package ast

import (
	"strconv"
	"strings"
)

const indent = "    "

var Unit = Expr{Node: TupleLit{}, Type: TupleSpec{}}

type FuncLit struct {
	Arg  Ident
	Body Expr
}

func (fl FuncLit) RenderGo(t Type) string {
	return t.RenderGo() + " { return " + fl.Body.RenderGo() + " }"
}

func (fl FuncLit) Equal(other FuncLit) bool {
	return fl.Arg == other.Arg && fl.Body.Equal(other.Body)
}

func (fl FuncLit) EqualExprNode(other ExprNode) bool {
	otherFuncLit, ok := other.(FuncLit)
	return ok && fl.Equal(otherFuncLit)
}

func (fl FuncLit) String() string {
	return fl.Arg.String() + " -> " + fl.Body.String()
}

type IntLit int

func (il IntLit) RenderGo(t Type) string {
	return strconv.Itoa(int(il))
}

func (il IntLit) EqualExprNode(other ExprNode) bool {
	otherIntLit, ok := other.(IntLit)
	return ok && il == otherIntLit
}

func (il IntLit) String() string { return strconv.Itoa(int(il)) }

type StringLit string

func (sl StringLit) RenderGo(t Type) string {
	return "\"" + string(sl) + "\""
}

func (sl StringLit) EqualExprNode(other ExprNode) bool {
	otherStringLit, ok := other.(StringLit)
	return ok && sl == otherStringLit
}

func (sl StringLit) String() string {
	return "\"" + string(sl) + "\""
}

type TupleLit []Expr

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

func (tl TupleLit) String() string {
	args := make([]string, len(tl))
	for i, arg := range tl {
		args[i] = arg.String()
	}
	return "(" + strings.Join(args, ", ") + ")"
}

type Ident string

func (i Ident) RenderGo(t Type) string {
	return string(i)
}

func (i Ident) EqualExprNode(other ExprNode) bool {
	otherIdent, ok := other.(Ident)
	return ok && i == otherIdent
}

func (i Ident) String() string { return string(i) }

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

func (c Call) String() string {
	return c.Fn.String() + " " + c.Arg.String()
}

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

func (b Block) String() string {
	out := make([]string, len(b.Stmts)+1)
	for i, stmt := range b.Stmts {
		out[i] = stmt.String()
	}
	out[len(out)-1] = b.Expr.String()
	return "{ " + strings.Join(out, "; ") + " }"
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
	Visit(ExprNodeVisitor)
	EqualExprNode(ExprNode) bool
	String() string
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

func (expr Expr) String() string { return expr.Node.String() }

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
