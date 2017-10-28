package ast

import "fmt"

type LetDecl struct {
	Ident   Ident
	Binding Expr
}

func (ld LetDecl) Equal(other LetDecl) bool {
	return ld.Ident == other.Ident && ld.Binding.Equal(other.Binding)
}

func (ld LetDecl) EqualDecl(other Decl) bool {
	otherLetDecl, ok := other.(LetDecl)
	return ok && ld.Equal(otherLetDecl)
}

func (ld LetDecl) EqualNode(other Node) bool {
	otherLetDecl, ok := other.(LetDecl)
	return ok && ld.Equal(otherLetDecl)
}

func (ld LetDecl) EqualStmt(other Stmt) bool {
	otherLetDecl, ok := other.(LetDecl)
	return ok && ld.Equal(otherLetDecl)
}

func (ld LetDecl) RenderGo() string {
	return fmt.Sprintf("var %s = %s", ld.Ident, ld.Binding.RenderGo())
}

type Node interface {
	node()
	EqualNode(other Node) bool
}

func (expr Expr) node()   {}
func (f File) node()      {}
func (td TypeDecl) node() {}
func (ld LetDecl) node()  {}

type Stmt interface {
	Node
	stmtNode()
	EqualStmt(Stmt) bool
	String() string
}

func (td TypeDecl) stmtNode() {}
func (ld LetDecl) stmtNode()  {}
func (expr Expr) stmtNode()   {}

type Decl interface {
	declNode()
	EqualDecl(other Decl) bool
}

func (td TypeDecl) String() string {
	panic("TypeDecl.String() not yet supported")
}

func (ld LetDecl) String() string {
	return "let " + ld.Ident.String() + " = " + ld.Binding.String()
}

func (td TypeDecl) declNode() {}
func (ld LetDecl) declNode()  {}
func (as ArgSpec) declNode()  {}
