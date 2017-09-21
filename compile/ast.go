package main

import (
	"fmt"
	"strconv"
)

type Ident string

func (i Ident) Equal(other Node) bool {
	if other, ok := other.(Ident); ok {
		return i == other
	}
	return false
}

func (i Ident) Pretty(base string) string {
	return "Ident(" + string(i) + ")"
}

type Int int

func (i Int) Equal(other Node) bool {
	if j, ok := other.(Int); ok {
		return j == i
	}
	return false
}

func (i Int) Pretty(base string) string {
	return "Int(" + strconv.Itoa(int(i)) + ")"
}

type String string

func (s String) Equal(other Node) bool {
	if s2, ok := other.(String); ok {
		return s == s2
	}
	return false
}

func (s String) Pretty(base string) string {
	return "String(" + string(s) + ")"
}

func (s String) String() string {
	return s.Pretty("")
}

type Expr struct {
	Head Term
	Tail *Expr
}

func NewExpr(head Term, tail ...Term) Expr {
	if len(tail) < 1 {
		return Expr{head, nil}
	}
	exprTail := NewExpr(tail[0], tail[1:]...)
	return Expr{head, &exprTail}
}

func (e Expr) EqualExpr(other Expr) bool {
	if !e.Head.EqualTerm(other.Head) {
		return false
	}
	if e.Tail != nil && other.Tail != nil {
		return e.Tail.EqualExpr(*other.Tail)
	}
	return e.Tail == other.Tail
}

func (e Expr) Equal(other Node) bool {
	if otherExpr, ok := other.(Expr); ok {
		return e.EqualExpr(otherExpr)
	}
	return false
}

func (e Expr) Pretty(base string) string {
	if e.Tail == nil {
		return "Expr{" + e.Head.Pretty(base) + "}"
	}
	return fmt.Sprintf(
		"Expr{\n%sHead: %s,\n%sTail: %s,\n%s}",
		base+indent,
		e.Head.Pretty(base+indent),
		base+indent,
		e.Tail.Pretty(base+indent),
		base,
	)
}

func (e Expr) String() string {
	return e.Pretty("")
}

func (e Expr) Ptr() *Expr {
	return &e
}

type Block struct {
	Stmts []Stmt
	Expr  *Expr // optional
}

func (b Block) Equal(other Node) bool {
	otherBlock, ok := other.(Block)
	return ok && b.EqualBlock(otherBlock)
}

func (b Block) EqualBlock(other Block) bool {
	if len(b.Stmts) != len(other.Stmts) {
		return false
	}
	for i, stmt := range b.Stmts {
		if !stmt.EqualStmt(other.Stmts[i]) {
			return false
		}
	}
	if b.Expr != nil && other.Expr != nil {
		return b.Expr.Equal(*other.Expr)
	}
	return b.Expr == other.Expr
}

func (b Block) Pretty(base string) string {
	nested := base + indent
	doubleNested := nested + indent
	stmtsString := ""
	for _, stmt := range b.Stmts {
		stmtsString += "\n" + doubleNested + stmt.Pretty(doubleNested) + ","
	}

	if b.Expr != nil {
		return fmt.Sprintf(
			"Block{\n%sStmts: [%s\n%s],\n%sExpr: %s,\n%s}",
			nested,
			stmtsString,
			nested,
			nested,
			b.Expr.Pretty(nested),
			base,
		)
	}
	return fmt.Sprintf("Block{Stmts: [%s\n%s]}", stmtsString, base)
}

type termType int

const (
	termTypeIdent termType = iota
	termTypeInt
	termTypeString
	termTypeExpr
	termTypeBlock
	termTypeLambda
)

type Term struct {
	tag    termType
	s      string
	i      int
	expr   *Expr
	block  *Block
	lambda *Lambda
}

func (t Term) String() string {
	return t.Pretty("")
}

func TermIdent(i Ident) Term {
	return Term{tag: termTypeIdent, s: string(i)}
}

func TermString(s String) Term {
	return Term{tag: termTypeString, s: string(s)}
}

func TermInt(i Int) Term {
	return Term{tag: termTypeInt, i: int(i)}
}

func TermExpr(expr Expr) Term {
	return Term{tag: termTypeExpr, expr: &expr}
}

func TermBlock(b Block) Term {
	return Term{tag: termTypeBlock, block: &b}
}

func TermLambda(l Lambda) Term {
	return Term{tag: termTypeLambda, lambda: &l}
}

// TermFromNode creates an Term from a Node. Panics if node isn't an term type.
func TermFromNode(n Node) Term {
	switch x := n.(type) {
	case Ident:
		return TermIdent(x)
	case String:
		return TermString(x)
	case Int:
		return TermInt(x)
	case Expr:
		return TermExpr(x)
	case Block:
		return TermBlock(x)
	case Lambda:
		return TermLambda(x)
	default:
		panic(fmt.Sprintf("Non-term node type '%T': %v", n, x))
	}
}

func (t Term) EqualTerm(other Term) bool {
	if t.tag == other.tag {
		switch t.tag {
		case termTypeIdent, termTypeString:
			return t.s == other.s
		case termTypeInt:
			return t.i == other.i
		case termTypeExpr:
			return t.expr.EqualExpr(*other.expr)
		case termTypeBlock:
			return t.block.EqualBlock(*other.block)
		case termTypeLambda:
			return t.lambda.EqualLambda(*other.lambda)
		default:
			panic(fmt.Sprint("Invalid term type:", t.tag))
		}
	}
	return false
}

func (t Term) Equal(other Node) bool {
	otherTerm, ok := other.(Term)
	return ok && t.EqualTerm(otherTerm)
}

func (t Term) Pretty(base string) string {
	switch t.tag {
	case termTypeIdent:
		return "Term(" + Ident(t.s).Pretty(base) + ")"
	case termTypeInt:
		return "Term(" + Int(t.i).Pretty(base) + ")"
	case termTypeString:
		return "Term(" + String(t.s).Pretty(base) + ")"
	case termTypeExpr:
		return "Term(" + t.expr.Pretty(base) + ")"
	case termTypeBlock:
		return "Term(" + t.block.Pretty(base) + ")"
	case termTypeLambda:
		return "Term(" + t.lambda.Pretty(base) + ")"
	default:
		panic(fmt.Sprint("Invalid term type:", t.tag))
	}
}

type BindingDecl struct {
	Binding Ident
	Expr    Expr
}

func (bd BindingDecl) Equal(other Node) bool {
	if otherBindingDecl, ok := other.(BindingDecl); ok {
		return bd.EqualBindingDecl(otherBindingDecl)
	}
	return false
}

func (bd BindingDecl) EqualBindingDecl(other BindingDecl) bool {
	return bd.Binding == other.Binding && bd.Expr.EqualExpr(other.Expr)
}

func (bd BindingDecl) Pretty(base string) string {
	nested := base + indent
	return fmt.Sprintf(
		"BindingDecl{\n%sBinding: %s,\n%sExpr: %s,\n%s}",
		nested,
		bd.Binding.Pretty(nested),
		nested,
		bd.Expr.Pretty(nested),
		base,
	)
}

type Stmt struct {
	isBinding bool
	expr      Expr
	binding   BindingDecl
}

func StmtExpr(expr Expr) Stmt {
	return Stmt{expr: expr}
}

func StmtBinding(decl BindingDecl) Stmt {
	return Stmt{isBinding: true, binding: decl}
}

func (s Stmt) Equal(other Node) bool {
	if otherStmt, ok := other.(Stmt); ok {
		return s.EqualStmt(otherStmt)
	}
	return false
}

func (s Stmt) EqualStmt(other Stmt) bool {
	if s.isBinding != other.isBinding {
		return false
	}
	if s.isBinding {
		return s.binding.EqualBindingDecl(other.binding)
	}
	return s.expr.EqualExpr(other.expr)
}

func (s Stmt) Pretty(base string) string {
	if s.isBinding {
		return "Stmt(" + s.binding.Pretty(base) + ")"
	}
	return "Stmt(" + s.expr.Pretty(base) + ")"
}

func StmtFromNode(n Node) Stmt {
	switch x := n.(type) {
	case BindingDecl:
		return StmtBinding(x)
	case Expr:
		return StmtExpr(x)
	default:
		panic(fmt.Sprintf("Non-term node type '%T': %v", n, x))
	}
}

// TODO TypeExpr: Type exprs are a subset of all exprs, specifically type exprs
// don't support lambdas or conditionals. basically just a, a b, a (b c), etc.

type TypeExpr struct {
	Head Ident
	Tail *TypeExpr
}

func (te TypeExpr) Equal(other Node) bool {
	otherTypeExpr, ok := other.(TypeExpr)
	return ok && te.EqualTypeExpr(otherTypeExpr)
}

func (te TypeExpr) EqualTypeExpr(other TypeExpr) bool {
	if te.Head != other.Head {
		return false
	}
	if te.Tail != nil && other.Tail != nil {
		return te.Tail.EqualTypeExpr(*other.Tail)
	}
	return te.Tail == other.Tail
}

func (te TypeExpr) Pretty(base string) string {
	if te.Tail != nil {
		nested := base + indent
		return fmt.Sprintf(
			"TypeExpr{\n%sHead: %s,\n%sTail: %s,\n%s}",
			nested,
			te.Head.Pretty(nested),
			nested,
			te.Tail.Pretty(nested),
			base,
		)
	}
	return fmt.Sprintf("TypeExpr{Head: %s}", te.Head.Pretty(base))
}

func (te TypeExpr) String() string {
	return te.Pretty("")
}

type ArgSpec struct {
	Name Ident
	Type *TypeExpr // optional
}

func (as ArgSpec) Equal(other Node) bool {
	otherArgSpec, ok := other.(ArgSpec)
	return ok && as.EqualArgSpec(otherArgSpec)
}

func (as ArgSpec) EqualArgSpec(other ArgSpec) bool {
	if as.Name != other.Name {
		return false
	}
	if as.Type != nil && other.Type != nil {
		return as.Type.EqualTypeExpr(*other.Type)
	}
	return as.Type == other.Type
}

func (as ArgSpec) Pretty(base string) string {
	if as.Type != nil {
		nested := base + indent
		return fmt.Sprintf(
			"ArgSpec{\n%sName: %s,\n%sType: %s,\n%s}",
			nested,
			as.Name.Pretty(nested),
			nested,
			as.Type.Pretty(nested),
			base,
		)
	}
	return fmt.Sprintf("ArgSpec{Name: %s}", as.Name.Pretty(base))
}

func (as ArgSpec) String() string {
	return as.Pretty("")
}

type Lambda struct {
	Args []ArgSpec
	Ret  *TypeExpr
	Body Expr
}

func (l Lambda) Equal(other Node) bool {
	otherLambda, ok := other.(Lambda)
	return ok && l.EqualLambda(otherLambda)
}

func (l Lambda) EqualLambda(other Lambda) bool {
	if len(l.Args) != len(other.Args) ||
		!l.Body.EqualExpr(other.Body) {
		return false
	}
	if l.Ret != nil && other.Ret != nil {
		if !l.Ret.EqualTypeExpr(*other.Ret) {
			return false
		}
	}
	for i, arg := range l.Args {
		if !arg.EqualArgSpec(other.Args[i]) {
			return false
		}
	}
	return true
}

func (l Lambda) Pretty(base string) string {
	nested := base + indent
	doubleNested := nested + indent

	argsString := ""
	for _, argSpec := range l.Args {
		argsString += "\n" + doubleNested + argSpec.Pretty(doubleNested) + ","
	}

	if l.Ret != nil {
		return fmt.Sprintf(
			"Lambda{\n%sArgs: [%s\n%s],\n%sRet: %s,\n%sBody: %s,\n%s}",
			nested,
			argsString,
			nested,
			nested,
			l.Ret.Pretty(nested),
			nested,
			l.Body.Pretty(nested),
			base,
		)
	}

	return fmt.Sprintf(
		"Lambda{\n%sArgs: [%s\n%s],\n%sBody: %s,\n%s}",
		nested,
		argsString,
		nested,
		nested,
		l.Body.Pretty(nested),
		base,
	)
}

func (l Lambda) String() string { return l.Pretty("") }

type File struct {
	Package string
	Imports []string
	Decls   []BindingDecl
}

func (f File) Equal(other Node) bool {
	otherFile, ok := other.(File)
	return ok && f.EqualFile(otherFile)
}

func (f File) EqualFile(other File) bool {
	if f.Package != f.Package ||
		len(f.Imports) != len(other.Imports) ||
		len(f.Decls) != len(other.Decls) {
		return false
	}
	for i, imp := range f.Imports {
		if imp != other.Imports[i] {
			return false
		}
	}
	for i, decl := range f.Decls {
		if !decl.EqualBindingDecl(other.Decls[i]) {
			return false
		}
	}
	return true
}

func (f File) Pretty(base string) string {
	nested := base + indent
	doubleNested := nested + indent
	importsString := ""
	for _, imp := range f.Imports {
		importsString += "\n" + doubleNested + imp + ","
	}
	declsString := ""
	for _, decl := range f.Decls {
		declsString += "\n" + doubleNested + decl.Pretty(doubleNested) + ","
	}
	return fmt.Sprintf(
		"File{\n%sPackage: %s,\n%sImports: [%s\n%s],\n%sDecls: [%s\n%s],\n%s}",
		nested,
		"\""+f.Package+"\"",
		nested,
		importsString,
		nested,
		nested,
		declsString,
		nested,
		base,
	)
}

func (f File) String() string {
	return f.Pretty("")
}
