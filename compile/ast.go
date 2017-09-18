package main

import (
	"fmt"
	"strconv"
	"strings"
)

type File struct {
	Package string
	Imports []string
	Decls   []FuncDecl
}

func (f File) Pretty(indent int) string {
	baseIndent := strings.Repeat(indentStr, indent)
	nestedIndent := baseIndent + indentStr
	var importsString string
	for _, imp := range f.Imports {
		importsString += "\n" + nestedIndent + indentStr + imp + ","
	}
	var declsString string
	for _, decl := range f.Decls {
		declsString += "\n" + nestedIndent + indentStr +
			decl.Pretty(indent+2) + ","
	}
	return fmt.Sprintf(
		"File{\n%sPackage: %s,\n%sImports: [%s\n%s],\n%sDecls: [%s\n%s],\n%s}",
		baseIndent+indentStr,
		f.Package,
		baseIndent+indentStr,
		importsString,
		baseIndent+indentStr,
		baseIndent+indentStr,
		declsString,
		baseIndent+indentStr,
		baseIndent,
	)
}

func (f File) Equal(other File) bool {
	if f.Package != other.Package ||
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
		if !decl.Equal(other.Decls[i]) {
			return false
		}
	}
	return true
}

type Type string

type ArgDecl struct {
	Name string
	Type Type
}

func (ad ArgDecl) Pretty(indent int) string {
	ind := strings.Repeat(indentStr, indent)
	return fmt.Sprintf(
		"ArgDecl{\n%sName: %s,\n%sType: %s,\n%s}",
		ind+indentStr,
		ad.Name,
		ind+indentStr,
		ad.Type,
		ind,
	)
}

type FuncDecl struct {
	Name string
	Args []ArgDecl
	Ret  Type
	Body Expr
}

func (fd FuncDecl) Equal(other FuncDecl) bool {
	if fd.Name != other.Name ||
		fd.Ret != other.Ret ||
		!fd.Body.Equal(other.Body) ||
		len(fd.Args) != len(other.Args) {
		return false
	}
	for i, arg := range fd.Args {
		if arg != other.Args[i] {
			return false
		}
	}
	return true
}

func (fd FuncDecl) Pretty(indent int) string {
	baseIndent := strings.Repeat(indentStr, indent)
	nestedIndent := baseIndent + indentStr

	ret := "()"
	if fd.Ret != "" {
		ret = string(fd.Ret)
	}

	var argsString string
	for _, arg := range fd.Args {
		argsString += "\n" + nestedIndent + indentStr + arg.Pretty(indent+2) +
			","
	}

	return fmt.Sprintf(
		"FuncDecl {\n%sName: %s,\n%sArgs: [%s\n%s],\n%sRet: %s,\n%sBody: %s,\n%s}",
		nestedIndent,
		fd.Name,
		nestedIndent,
		argsString,
		nestedIndent,
		nestedIndent,
		ret,
		nestedIndent,
		fd.Body.Pretty(indent+1),
		baseIndent,
	)
}

type exprType int

const (
	exprTypeIdent exprType = iota
	exprTypeInt
	exprTypeStr
	exprTypeUnExpr
	exprTypeBinExpr
	exprTypeBlockExpr
	exprTypeCallExpr
)

func (et exprType) String() string {
	switch et {
	case exprTypeIdent:
		return "IDENT"
	case exprTypeInt:
		return "INT"
	case exprTypeStr:
		return "STR"
	case exprTypeUnExpr:
		return "UNEXPR"
	case exprTypeBinExpr:
		return "BINEXPR"
	case exprTypeBlockExpr:
		return "BLOCKEXPR"
	case exprTypeCallExpr:
		return "CALLEXPR"
	default:
		panic(fmt.Sprintf("Invalid expr type: %d"))
	}
}

type CallExpr struct {
	Callable  Expr
	Arguments []Expr
}

func (ce CallExpr) Pretty(indent int) string {
	baseIndent := strings.Repeat(indentStr, indent)
	argStr := "["
	for _, arg := range ce.Arguments {
		argStr += "\n" + baseIndent + indentStr + indentStr +
			arg.Pretty(indent+2) + ","
	}
	return fmt.Sprintf(
		"CallExpr{\n%sCallable: %s,\n%sArguments: %s,\n%s}",
		baseIndent+indentStr,
		ce.Callable.Pretty(indent+1),
		baseIndent+indentStr,
		argStr+"\n"+baseIndent+indentStr+"]",
		baseIndent,
	)
}

func (ce CallExpr) Equal(other CallExpr) bool {
	if !ce.Callable.Equal(other.Callable) {
		return false
	}
	if len(ce.Arguments) != len(other.Arguments) {
		return false
	}
	for i, arg := range ce.Arguments {
		if !arg.Equal(ce.Arguments[i]) {
			return false
		}
	}
	return true
}

func (ce CallExpr) String() string {
	argStrs := make([]string, len(ce.Arguments))
	for i, arg := range ce.Arguments {
		argStrs[i] = arg.String()
	}
	return ce.Callable.String() + "(" + strings.Join(argStrs, ", ") + ")"
}

type UnExpr struct {
	Operator rune
	Operand  Expr
}

var indentStr string = "    "

func (ue UnExpr) Pretty(indent int) string {
	return fmt.Sprintf(
		"UnExpr{\n%sOperator: '%s',\n%sOperand: %s,\n%s}",
		strings.Repeat(indentStr, indent+1),
		string(ue.Operator),
		strings.Repeat(indentStr, indent+1),
		ue.Operand.Pretty(indent+1),
		strings.Repeat(indentStr, indent),
	)
}

func (ue UnExpr) Equal(other UnExpr) bool {
	return ue.Operator == other.Operator && ue.Operand.Equal(other.Operand)
}

func (ue UnExpr) String() string {
	return string(ue.Operator) + ue.Operand.String()
}

type BinExpr struct {
	Left, Right Expr
	Operator    string
}

func (be BinExpr) Pretty(indent int) string {
	return fmt.Sprintf(
		"BinExpr{\n%sLeft: %s,\n%sOperator: \"%s\",\n%sRight: %s,\n%s}",
		strings.Repeat(indentStr, indent+1),
		be.Left.Pretty(indent+1),
		strings.Repeat(indentStr, indent+1),
		be.Operator,
		strings.Repeat(indentStr, indent+1),
		be.Right.Pretty(indent+1),
		strings.Repeat(indentStr, indent),
	)
}

func (be BinExpr) Equal(other BinExpr) bool {
	return be.Left.Equal(other.Left) &&
		be.Operator == other.Operator &&
		be.Right.Equal(other.Right)
}

func (be BinExpr) String() string {
	return fmt.Sprintf("%s %s %s", be.Left, be.Operator, be.Right)
}

type LetStmt struct {
	Binding string
	Expr    Expr
}

func (ls LetStmt) String() string {
	return "let " + ls.Binding + " = " + ls.Expr.String()
}

func (ls LetStmt) Pretty(indent int) string {
	baseIndent := strings.Repeat(indentStr, indent)
	nestedIndent := baseIndent + indentStr

	return fmt.Sprintf(
		"LetStmt{\n%sBinding: %s,\n%sExpr: %s,\n%s}",
		nestedIndent,
		ls.Binding,
		nestedIndent,
		ls.Expr.Pretty(indent+1),
		baseIndent,
	)
}

func (ls LetStmt) Equal(other LetStmt) bool {
	return ls.Binding == other.Binding && ls.Expr.Equal(other.Expr)
}

type stmtType int

const (
	stmtTypeLetStmt stmtType = iota
	stmtTypeExpr
)

type Stmt struct {
	tag     stmtType
	letStmt LetStmt
	expr    Expr
}

func StmtLetStmt(letStmt LetStmt) Stmt {
	return Stmt{tag: stmtTypeLetStmt, letStmt: letStmt}
}

func StmtExpr(expr Expr) Stmt {
	return Stmt{tag: stmtTypeExpr, expr: expr}
}

func (s Stmt) String() string {
	switch s.tag {
	case stmtTypeLetStmt:
		return s.letStmt.String() + ";"
	case stmtTypeExpr:
		return s.expr.String() + ";"
	default:
		panic(fmt.Sprint("Invalid statement type:", s.tag))
	}
}

func (s Stmt) Pretty(indent int) string {
	switch s.tag {
	case stmtTypeLetStmt:
		return s.letStmt.Pretty(indent)
	case stmtTypeExpr:
		return s.expr.Pretty(indent)
	default:
		panic(fmt.Sprint("Invalid statement type:", s.tag))
	}
}

func (s Stmt) Equal(other Stmt) bool {
	if s.tag != other.tag {
		return false
	}
	switch s.tag {
	case stmtTypeLetStmt:
		return s.letStmt.Equal(other.letStmt)
	case stmtTypeExpr:
		return s.expr.Equal(other.expr)
	default:
		panic(fmt.Sprint("Invalid statement type:", s.tag))
	}
}

type BlockExpr struct {
	Stmts []Stmt
	Expr  *Expr // optional
}

func (be BlockExpr) String() string {
	statementStrings := make([]string, len(be.Stmts))
	for i, statement := range be.Stmts {
		statementStrings[i] = statement.String()
	}
	exprString := ""
	if be.Expr != nil {
		exprString = "\n" + indentStr + be.Expr.String()
	}
	return fmt.Sprintf(
		"{\n%s%s%s\n}",
		indentStr,
		strings.Join(statementStrings, "\n"+indentStr),
		exprString,
	)
}

func (be BlockExpr) Pretty(indent int) string {
	baseIndent := strings.Repeat(indentStr, indent)
	nestedIndent := baseIndent + indentStr

	var stmtsString string
	for _, stmt := range be.Stmts {
		stmtsString += "\n" + nestedIndent + indentStr +
			stmt.Pretty(indent+2) + ","
	}

	exprString := "<nil>"
	if be.Expr != nil {
		exprString = be.Expr.Pretty(indent + 1)
	}

	return fmt.Sprintf(
		"BlockExpr{\n%sStmts: [%s\n%s],\n%sExpr: %s,\n%s}",
		nestedIndent,
		stmtsString,
		nestedIndent,
		nestedIndent,
		exprString,
		baseIndent,
	)
}

func (be BlockExpr) Equal(other BlockExpr) bool {
	if len(be.Stmts) != len(other.Stmts) {
		return false
	}
	for i, statement := range be.Stmts {
		if !statement.Equal(other.Stmts[i]) {
			return false
		}
	}
	if be.Expr != nil && other.Expr != nil {
		return be.Expr.Equal(*other.Expr)
	}
	return be.Expr == nil && other.Expr == nil
}

type Expr struct {
	tag       exprType
	i         int
	s         string
	unExpr    *UnExpr
	binExpr   *BinExpr
	blockExpr BlockExpr
	callExpr  *CallExpr
}

func (e Expr) Ptr() *Expr {
	return &e
}

func (e Expr) Pretty(indent int) string {
	switch e.tag {
	case exprTypeIdent:
		return e.s
	case exprTypeInt:
		return fmt.Sprint(e.i)
	case exprTypeStr:
		return "\"" + e.s + "\""
	case exprTypeUnExpr:
		return e.unExpr.Pretty(indent)
	case exprTypeBinExpr:
		return e.binExpr.Pretty(indent)
	case exprTypeBlockExpr:
		return e.blockExpr.Pretty(indent)
	case exprTypeCallExpr:
		return e.callExpr.Pretty(indent)
	default:
		panic(fmt.Sprintf("Invalid expr type: %d", e.tag))
	}
}

func (e Expr) Equal(other Expr) bool {
	if e.tag != other.tag {
		return false
	}

	switch e.tag {
	case exprTypeIdent:
		return e.s == other.s
	case exprTypeInt:
		return e.i == other.i
	case exprTypeStr:
		return e.s == other.s
	case exprTypeUnExpr:
		return e.unExpr.Equal(*other.unExpr)
	case exprTypeBinExpr:
		return e.binExpr.Equal(*other.binExpr)
	case exprTypeBlockExpr:
		return e.blockExpr.Equal(other.blockExpr)
	case exprTypeCallExpr:
		return e.callExpr.Equal(*other.callExpr)
	default:
		panic(fmt.Sprintf("Invalid expr type: %d", e.tag))
	}
}

func (e Expr) String() string {
	switch e.tag {
	case exprTypeIdent:
		return e.s
	case exprTypeInt:
		return strconv.Itoa(e.i)
	case exprTypeStr:
		return `"` + e.s + `"`
	case exprTypeUnExpr:
		return e.unExpr.String()
	case exprTypeBinExpr:
		return e.binExpr.String()
	case exprTypeBlockExpr:
		return e.blockExpr.String()
	case exprTypeCallExpr:
		return e.callExpr.String()
	default:
		panic(fmt.Sprintf("Invalid expr type: %d", e.tag))
	}
}

func ExprIdent(i string) Expr {
	return Expr{tag: exprTypeIdent, s: i}
}

func ExprInt(i int) Expr {
	return Expr{tag: exprTypeInt, i: i}
}

func ExprStr(s string) Expr {
	return Expr{tag: exprTypeStr, s: s}
}

func ExprCallExpr(callExpr *CallExpr) Expr {
	return Expr{tag: exprTypeCallExpr, callExpr: callExpr}
}

func ExprUnExpr(unExpr *UnExpr) Expr {
	return Expr{tag: exprTypeUnExpr, unExpr: unExpr}
}

func ExprBinExpr(binExpr *BinExpr) Expr {
	return Expr{tag: exprTypeBinExpr, binExpr: binExpr}
}

func ExprBlockExpr(blockExpr BlockExpr) Expr {
	return Expr{tag: exprTypeBlockExpr, blockExpr: blockExpr}
}
