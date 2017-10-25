package parser

import (
	"github.com/weberc2/gallium/ast"
	"github.com/weberc2/gallium/combinator"
)

func List(p combinator.Parser, delim combinator.Parser) combinator.Parser {
	return combinator.Seq(
		p,
		combinator.Repeat(combinator.Seq(delim, p).Get(1)),
	).MapSlice(
		func(vs []interface{}) interface{} {
			return append([]interface{}{vs[0]}, vs[1].([]interface{})...)
		},
	).Rename("List")
}

func Ref(p *combinator.Parser) combinator.Parser {
	return func(input combinator.Input) combinator.Result {
		return (*p)(input)
	}
}

func TupleSpec(input combinator.Input) combinator.Result {
	return combinator.Seq(
		combinator.Lit('('),
		List(
			Type,
			combinator.Seq(
				combinator.CanWS,
				combinator.Lit(','),
				combinator.CanWS,
			),
		),
		combinator.Lit(')'),
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

func TypeExpr(input combinator.Input) combinator.Result {
	return combinator.Seq(
		combinator.Ident,
		combinator.Opt(combinator.Seq(combinator.WS, Type).Get(1)),
	).MapSlice(
		func(vs []interface{}) interface{} {
			var arg ast.Type
			if vs[1] != nil {
				arg = vs[1].(ast.Type)
			}
			return ast.TypeRef{Name: vs[0].(string), Arg: arg}
		},
	).Wrap()(input)
}

func Type(input combinator.Input) combinator.Result {
	return combinator.Any(FuncSpec, TypeExpr, TupleSpec).Wrap()(input)
}

// func FuncSpec(input combinator.Input) combinator.Result {
// 	panic("Fix me")
// 	// return combinator.Seq(
// 	// 	combinator.StrLit("fn"), // 0
// 	// 	combinator.CanWS,        // 1
// 	// 	combinator.Lit('('),     // 2
// 	// 	combinator.Opt(List(
// 	// 		ArgSpec,
// 	// 		combinator.Seq(
// 	// 			combinator.CanWS,
// 	// 			combinator.Lit(','),
// 	// 			combinator.CanWS,
// 	// 		),
// 	// 	)), // 3
// 	// 	combinator.Lit(')'), // 4
// 	// 	combinator.Opt(combinator.Seq(
// 	// 		combinator.CanWS,
// 	// 		combinator.StrLit("->"),
// 	// 		combinator.CanWS,
// 	// 		Type,
// 	// 	).Get(3)), // 5
// 	// ).MapSlice(func(vs []interface{}) interface{} {
// 	// 	var args []ast.ArgSpec
// 	// 	if vs[3] != nil {
// 	// 		argValues := vs[3].([]interface{})
// 	// 		args = make([]ast.ArgSpec, len(argValues))
// 	// 		for i, v := range argValues {
// 	// 			args[i] = v.(ast.ArgSpec)
// 	// 		}
// 	// 	}
// 	// 	var ret ast.Type
// 	// 	if vs[5] != nil {
// 	// 		ret = vs[5].(ast.Type)
// 	// 	}
// 	// 	return ast.FuncSpec{Args: args, Ret: ret}
// 	// }).Wrap()(input)
// }

func ArgSpec(input combinator.Input) combinator.Result {
	return combinator.Seq(
		combinator.Ident,
		combinator.Opt(combinator.Seq(combinator.WS, Type).Get(1)),
	).MapSlice(
		func(vs []interface{}) interface{} {
			var typ ast.Type
			if vs[1] != nil {
				typ = vs[1].(ast.Type)
			}
			return ast.ArgSpec{Name: vs[0].(string), Type: typ}
		},
	).Wrap()(input)
}

func ExprList(input combinator.Input) combinator.Result {
	return combinator.Seq(
		Expr,
		combinator.Opt(combinator.Seq(
			combinator.CanWS,
			combinator.Lit(','),
			combinator.CanWS,
			ExprList,
		).Get(3)),
	).MapSlice(func(vs []interface{}) interface{} {
		var tail []ast.Expr
		if vs[1] != nil {
			tail = vs[1].([]ast.Expr)
		}
		return append([]ast.Expr{vs[0].(ast.Expr)}, tail...)
	}).Wrap()(input)
}

func Atom(input combinator.Input) combinator.Result {
	return combinator.Any(TupleLit, Ident, IntLit, StringLit).Map(
		func(v interface{}) interface{} {
			return ast.Expr{Node: v.(ast.ExprNode)}
		},
	).Rename("Atom")(input)
}

func Expr(input combinator.Input) combinator.Result {
	return combinator.Any(
		combinator.Any(Block, Call, FuncLit).Map(
			func(v interface{}) interface{} {
				return ast.Expr{Node: v.(ast.ExprNode)}
			},
		),
		Atom,
	).Wrap()(input)
}

func TupleLit(input combinator.Input) combinator.Result {
	multi := combinator.Seq(
		combinator.Lit('('),
		// get the first n-1 elements and then the last element and return it
		// all as a TupleLit
		combinator.Seq(
			combinator.Repeat(
				combinator.Seq(
					combinator.CanWS,
					Expr,
					combinator.CanWS,
					combinator.Lit(','),
				).Get(1),
			).MapSlice(func(vs []interface{}) interface{} {
				exprs := make(ast.TupleLit, len(vs))
				for i, v := range vs {
					exprs[i] = v.(ast.Expr)
				}
				return exprs
			}),
			combinator.CanWS,
			Expr,
			combinator.CanWS,
		).MapSlice(func(vs []interface{}) interface{} {
			return append(vs[0].(ast.TupleLit), vs[2].(ast.Expr))
		}),
		combinator.Lit(')'),
	).Get(1)
	unit := combinator.Seq(
		combinator.Lit('('),
		combinator.CanWS,
		combinator.Lit(')'),
	).Map(func(v interface{}) interface{} { return ast.TupleLit{} })
	return combinator.Any(unit, multi).Wrap()(input)
}

func Block(input combinator.Input) combinator.Result {
	return combinator.Seq(
		combinator.Lit('{'),
		combinator.CanWS,
		combinator.Repeat(combinator.Seq(
			Stmt,
			combinator.CanWS,
		).Get(0)).MapSlice(func(vs []interface{}) interface{} {
			var stmts []ast.Stmt
			for _, v := range vs {
				stmts = append(stmts, v.(ast.Stmt))
			}
			return stmts
		}),
		combinator.Opt(combinator.Seq(Expr, combinator.CanWS).Get(0)),
		combinator.Lit('}'),
	).MapSlice(func(vs []interface{}) interface{} {
		var expr ast.Expr
		if vs[3] != nil {
			expr = vs[3].(ast.Expr)
		}
		return ast.Block{Stmts: vs[2].([]ast.Stmt), Expr: expr}
	}).Wrap()(input)
}

func Call(input combinator.Input) combinator.Result {
	seqToCall := func(vs []interface{}) interface{} {
		return ast.Call{Fn: vs[0].(ast.Expr), Arg: vs[2].(ast.Expr)}
	}
	callToExpr := func(v interface{}) interface{} {
		return ast.Expr{Node: v.(ast.Call)}
	}
	simple := combinator.Seq(Atom, combinator.WS, Atom).MapSlice(seqToCall)
	complex := combinator.Seq(
		simple.Map(callToExpr),
		combinator.WS,
		Atom,
	).MapSlice(seqToCall)
	return combinator.Any(complex, simple).Wrap()(input)
}

func FuncSpec(input combinator.Input) combinator.Result {
	return combinator.Seq(
		Ident,
		combinator.CanWS,
		combinator.StrLit("->"),
		combinator.CanWS,
		Type,
	).MapSlice(func(vs []interface{}) interface{} {
		return ast.FuncSpec{
			Arg: ast.TypeRef{Name: string(vs[0].(ast.Ident))},
			Ret: vs[1].(ast.Type),
		}
	}).Wrap()(input)
}

func FuncLit(input combinator.Input) combinator.Result {
	return combinator.Seq(
		Ident,
		combinator.CanWS,
		combinator.StrLit("->"),
		combinator.CanWS,
		Expr,
	).MapSlice(func(vs []interface{}) interface{} {
		return ast.FuncLit{Arg: vs[0].(ast.Ident), Body: vs[4].(ast.Expr)}
	}).Wrap()(input)
}

func LetDecl(input combinator.Input) combinator.Result {
	return combinator.Seq(
		combinator.StrLit("let"), // 0
		combinator.WS,            // 1
		Ident,                    // 2
		combinator.CanWS,         // 3
		combinator.Lit('='),      // 4
		combinator.CanWS,         // 5
		Expr,                     // 6
	).MapSlice(func(vs []interface{}) interface{} {
		return ast.LetDecl{vs[2].(ast.Ident), vs[6].(ast.Expr)}
	}).Wrap()(input)
}

func Decl(input combinator.Input) combinator.Result {
	return combinator.Any(LetDecl, TypeDecl).Wrap()(input)
}

func Stmt(input combinator.Input) combinator.Result {
	return combinator.Seq(
		combinator.Any(Decl, Expr),
		combinator.EOS,
	).Get(0).Wrap()(input)
}

var (
	TypeLit = combinator.Any(
		combinator.Ident.Map(func(v interface{}) interface{} {
			return ast.TypeRef{Name: v.(string)}
		}),
		TupleSpec,
		FuncSpec,
	).Rename("TypeLit")

	TypeDecl = combinator.Seq(
		combinator.StrLit("type"), // 0
		combinator.WS,             // 1
		combinator.Seq(
			combinator.Ident,
			combinator.Repeat(
				combinator.Seq(combinator.WS, combinator.Ident).Get(1),
			),
		), // 2
		combinator.CanWS,    // 3
		combinator.Lit('='), // 4
		combinator.CanWS,    // 5
		Type,                // 6
	).MapSlice(func(vs []interface{}) interface{} {
		typeExpr := vs[2].([]interface{})

		argNodes := typeExpr[1].([]interface{})
		var args []ast.TypeVar
		if len(argNodes) > 0 {
			args = make([]ast.TypeVar, len(argNodes))
			for i, v := range argNodes {
				args[i] = ast.TypeVar(v.(string))
			}
		}

		return ast.TypeDecl{
			Name: typeExpr[0].(string),
			Type: vs[6].(ast.Type),
			Args: args,
		}
	}).Rename("TypeDecl")

	IntLit = combinator.Int.Map(func(v interface{}) interface{} {
		return ast.IntLit(v.(int))
	}).Rename("IntLit")

	StringLit = combinator.String.Map(func(v interface{}) interface{} {
		return ast.StringLit(v.(string))
	}).Rename("StringLit")

	// FuncLit = Seq(FuncSpec, WS, Expr)

	Ident = combinator.Ident.Map(func(v interface{}) interface{} {
		return ast.Ident(v.(string))
	}).Rename("Ident")

	File = combinator.Seq(
		combinator.StrLit("package"), // 0
		combinator.WS,                // 1
		combinator.Ident,             // 2
		combinator.Repeat(combinator.Seq(
			combinator.CanWS,
			Stmt,
		).Get(1)), // 3
		combinator.CanWS, // 4
		combinator.EOF,   // 5
	).MapSlice(func(vs []interface{}) interface{} {
		stmtNodes := vs[3].([]interface{})
		stmts := make([]ast.Stmt, len(stmtNodes))
		for i, v := range stmtNodes {
			stmts[i] = v.(ast.Stmt)
		}
		return ast.File{Package: vs[2].(string), Stmts: stmts}
	}).Rename("File")
)
