package parser

import (
	"github.com/weberc2/gallium/ast"
)

func List(p Parser, delim Parser) Parser {
	return Seq(p, Repeat(Seq(delim, p).Get(1))).MapSlice(
		func(vs []interface{}) interface{} {
			return append([]interface{}{vs[0]}, vs[1].([]interface{})...)
		},
	).Rename("List")
}

func Ref(p *Parser) Parser {
	return func(input Input) Result { return (*p)(input) }
}

func TupleSpec(input Input) Result {
	return Seq(
		Lit('('),
		List(Type, Seq(CanWS, Lit(','), CanWS)),
		Lit(')'),
	).Get(1).MapSlice(
		func(vs []interface{}) interface{} {
			ts := make(ast.TupleSpec, len(vs))
			for i, v := range vs {
				ts[i] = v.(ast.Type)
			}
			return ts
		},
	).Wrap()(input)
}

func TypeExpr(input Input) Result {
	return Seq(Ident, Opt(Seq(WS, Type).Get(1))).MapSlice(
		func(vs []interface{}) interface{} {
			var arg ast.Type
			if vs[1] != nil {
				arg = vs[1].(ast.Type)
			}
			return ast.TypeRef{Name: vs[0].(string), Arg: arg}
		},
	).Wrap()(input)
}

func Type(input Input) Result {
	return Any(FuncSpec, TypeExpr, TupleSpec).Wrap()(input)
}

func FuncSpec(input Input) Result {
	return Seq(
		StrLit("fn"), // 0
		CanWS,        // 1
		Lit('('),     // 2
		Opt(List(ArgSpec, Seq(CanWS, Lit(','), CanWS))), // 3
		Lit(')'), // 4
		Opt(Seq(CanWS, StrLit("->"), CanWS, Type).Get(3)), // 5
	).MapSlice(func(vs []interface{}) interface{} {
		var args []ast.ArgSpec
		if vs[3] != nil {
			argValues := vs[3].([]interface{})
			args = make([]ast.ArgSpec, len(argValues))
			for i, v := range argValues {
				args[i] = v.(ast.ArgSpec)
			}
		}
		var ret ast.Type
		if vs[5] != nil {
			ret = vs[5].(ast.Type)
		}
		return ast.FuncSpec{Args: args, Ret: ret}
	}).Wrap()(input)
}

func TypeAtom(input Input) Result {
	return Any(
		TupleSpec,
		FuncSpec,
		Ident.Map(func(v interface{}) interface{} {
			return ast.TypeRef{Name: v.(string)}
		}),
	).Wrap()(input)
}

func ArgSpec(input Input) Result {
	return Seq(Ident, WS, Type).MapSlice(
		func(vs []interface{}) interface{} {
			return ast.ArgSpec{Name: vs[0].(string), Type: vs[2].(ast.Type)}
		},
	).Wrap()(input)
}

var (
	TypeLit = Any(
		Ident.Map(func(v interface{}) interface{} {
			return ast.TypeRef{Name: v.(string)}
		}),
		TupleSpec,
		FuncSpec,
	)

	TypeDecl = Seq(
		StrLit("type"), // 0
		WS,             // 1
		Seq(Ident, Repeat(Seq(WS, Ident).Get(1))), // 2
		CanWS,    // 3
		Lit('='), // 4
		CanWS,    // 5
		Type,     // 6
		EOS,      // 7
	).MapSlice(func(vs []interface{}) interface{} {
		typeExpr := vs[2].([]interface{})

		argNodes := typeExpr[1].([]interface{})
		args := make([]ast.TypeVar, len(argNodes))
		for i, v := range argNodes {
			args[i] = ast.TypeVar(v.(string))
		}

		return ast.TypeDecl{
			Name: typeExpr[0].(string),
			Type: vs[6].(ast.Type),
			Args: args,
		}
	}).Rename("TypeDecl")

	File = Seq(
		StrLit("package"), // 0
		WS,                // 1
		Ident,             // 2
		Repeat(Seq(CanWS, TypeDecl).Get(1)), // 3
		CanWS, // 4
		EOF,   // 5
	).MapSlice(func(vs []interface{}) interface{} {
		declNodes := vs[3].([]interface{})
		decls := make([]ast.Decl, len(declNodes))
		for i, v := range declNodes {
			decls[i] = v.(ast.TypeDecl)
		}
		return ast.File{Package: vs[2].(string), Decls: decls}
	}).Rename("File")
)
