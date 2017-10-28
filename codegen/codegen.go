package codegen

import (
	"fmt"
	"strconv"

	"github.com/dave/jennifer/jen"
	"github.com/weberc2/gallium/ast"
)

func Type(t ast.Type) *jen.Statement {
	switch x := t.(type) {
	case ast.Primitive:
		switch x {
		case "int":
			return jen.Int()
		case "string":
			return jen.String()
		default:
			panic("codegen not supported for primitive:" + string(x))
		}
	case ast.TupleSpec:
		types := make([]jen.Code, len(x))
		for i, t := range x {
			types[i] = jen.Id("_" + strconv.Itoa(i)).Add(Type(t))
		}
		return jen.Struct(types...)
	case ast.TypeVar:
		panic("codegen not supported for generic types")
	default:
		panic(fmt.Sprintf("codegen not supported for %T", t))
	}
}

func Expr(expr ast.Expr) *jen.Statement {
	switch x := expr.Node.(type) {
	case ast.IntLit:
		return jen.Lit(int(x))
	case ast.StringLit:
		return jen.Lit(string(x))
	case ast.Ident:
		return jen.Id(string(x))
	case ast.TupleLit:
		fields := make([]jen.Code, len(x))
		for i, expr := range x {
			fields[i] = jen.Id("_" + strconv.Itoa(i)).Op(":").Add(Expr(expr))
		}
		return jen.Add(Type(expr.Type)).Values(fields...)
	case ast.FuncLit:
		fs := expr.Type.(ast.FuncSpec)
		return jen.Func().Params(
			jen.Id(string(x.Arg)).Add(Type(fs.Arg)),
		).Add(Type(fs.Ret)).Add(jen.Block(jen.Return(Expr(x.Body))))
	case ast.Call:
		return jen.Add(Expr(x.Fn)).Call(Expr(x.Arg))
	default:
		panic(fmt.Sprintf(
			"Expr() not yet implemented for %T",
			expr.Node,
		))
	}
}

func Stmt(stmt ast.Stmt) *jen.Statement {
	switch x := stmt.(type) {
	case ast.LetDecl:
		return jen.Var().Id(string(x.Ident)).Op("=").Add(Expr(x.Binding))
	default:
		panic(fmt.Sprint("Stmt() not yet implemented for %T", stmt))
	}
}

func File(f ast.File) *jen.File {
	out := jen.NewFile(f.Package)
	for _, stmt := range f.Stmts {
		out.Add(Stmt(stmt))
	}
	return out
}
