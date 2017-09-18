package main

import (
	"fmt"
	"strconv"
	"strings"
)

func CompileStr(s string) string {
	return "\"" + s + "\""
}

func CompileImports(imports []string) string {
	var out string
	for _, imp := range imports {
		out += "\n\t" + CompileStr(imp)
	}
	return "import (" + out + "\n)"
}

func CompileArgDecls(decls []ArgDecl) string {
	declStrings := make([]string, len(decls))
	for i, decl := range decls {
		declStrings[i] = decl.Name + ": " + string(decl.Type)
	}
	return strings.Join(declStrings, ", ")
}

func CompileArgs(args []Expr) string {
	argStrings := make([]string, len(args))
	for i, arg := range args {
		argStrings[i] = CompileExpr(arg)
	}
	return strings.Join(argStrings, ", ")
}

func CompileCallExpr(callExpr CallExpr) string {
	return CompileExpr(callExpr.Callable) +
		"(" + CompileArgs(callExpr.Arguments) + ")"
}

func CompileLetStmt(ls LetStmt) string {
	return ls.Binding + " := " + CompileExpr(ls.Expr)
}

func CompileStmt(stmt Stmt) string {
	switch stmt.tag {
	case stmtTypeLetStmt:
		return CompileLetStmt(stmt.letStmt)
	case stmtTypeExpr:
		return CompileExpr(stmt.expr)
	default:
		panic(fmt.Sprint("Invalid statement type:", stmt.tag))
	}
}

func CompileBlockExpr(blockExpr BlockExpr) string {
	var out string
	for _, stmt := range blockExpr.Stmts {
		out += "\n\t" + CompileStmt(stmt)
	}
	if blockExpr.Expr != nil {
		out += "\n\t" + CompileExpr(*blockExpr.Expr)
	}
	return "{" + out + "\n}"
}

func CompileExpr(expr Expr) string {
	switch expr.tag {
	case exprTypeIdent:
		return expr.s
	case exprTypeStr:
		return CompileStr(expr.s)
	case exprTypeInt:
		return strconv.Itoa(expr.i)
	case exprTypeCallExpr:
		return CompileCallExpr(*expr.callExpr)
	case exprTypeBlockExpr:
		return CompileBlockExpr(expr.blockExpr)
	default:
		panic("Expr type not implemented: " + expr.String())
	}
}

func CompileDecls(decls []FuncDecl) string {
	out := make([]string, len(decls))
	for i, decl := range decls {
		var ret string
		if decl.Ret != "" {
			ret = " -> " + string(decl.Ret)
		}
		out[i] = fmt.Sprintf(
			"func %s(%s)%s %s",
			decl.Name,
			CompileArgDecls(decl.Args),
			ret,
			CompileExpr(decl.Body),
		)
	}
	return strings.Join(out, "\n\n")
}

func CompileFile(f File) string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		"package "+f.Package,
		CompileImports(f.Imports),
		CompileDecls(f.Decls),
	)
}
