package main

import (
	"fmt"
	"unicode"
)

func ParseInt(input Input) Result {
	last := input
	head, tail := input.Cons()
	if !unicode.IsDigit(head) {
		return Err(
			input,
			fmt.Errorf("Wanted <digit>, got '%s'", string(head)),
		)
	}

	i := 0
	for {
		if !unicode.IsDigit(head) {
			return Ok(Int(i), last)
		}
		i = i*10 + int(head-rune('0'))
		last = tail
		head, tail = tail.Cons()
	}
}

func ParseChars(input Input) Result {
	runes := []rune{}
	last := input
	for h, t := input.Cons(); h != rune(0); h, t = t.Cons() {
		if h == '"' {
			return Ok(String(string(runes)), last)
		}
		runes = append(runes, h)
		last = t
	}
	return Err(
		input,
		fmt.Errorf("Encountered '\"' before end of CHARS"),
	)
}

func ParseString(input Input) Result {
	quo := Lit('"')
	return Wrap(Map(Seq(quo, ParseChars, quo), func(n Node) Node {
		if nodes, ok := n.(Nodes); ok {
			if len(nodes) != 3 {
				panic("Invalid string parser! Expected len(nodes) == 3!")
			}
			return nodes[1]
		}
		panic("Invalid string parser! Expected node-type Nodes!")
	}))(input)
}

func ParseIdent(input Input) Result {
	h, t := input.Cons()
	if h != '_' && !unicode.IsLetter(h) {
		return Err(input, fmt.Errorf("Wanted <letter> or '_'"))
	}

	last := t
	runes := []rune{h}
	for {
		h, t = t.Cons()
		if h != '_' && !unicode.IsDigit(h) && !unicode.IsLetter(h) {
			return Ok(Ident(string(runes)), last)
		}
		runes = append(runes, h)
		last = t
	}
}

func ParseParenGroup(input Input) Result {
	return Wrap(Map(Seq(Lit('('), ParseExpr, Lit(')')), func(n Node) Node {
		return n.(Nodes)[1].(Expr)
	}))(input)
}

func ParseTerm(input Input) Result {
	return Wrap(Map(
		Any(
			ParseInt,
			ParseString,
			ParseIdent,
			ParseParenGroup,
			ParseBlock,
			ParseLambda,
		),
		func(n Node) Node { return TermFromNode(n) },
	))(input)
}

func ParseExpr(input Input) Result {
	head := ParseTerm(input)
	if head.IsErr() {
		return head
	}
	tail := ParseExpr(head.Rest.SkipWS())
	if tail.IsErr() {
		return Ok(Expr{head.Node.(Term), nil}, tail.Rest)
	}

	if expr, ok := tail.Node.(Expr); ok {
		return Ok(Expr{head.Node.(Term), &expr}, tail.Rest)
	}

	panic("Invalid `Expr` node!")
}

func ParseBindingDecl(input Input) Result {
	return Wrap(Map(Seq(
		StrLit("let"),
		MustWS,
		ParseIdent,
		CanWS,
		Lit('='),
		CanWS,
		ParseExpr,
	), func(n Node) Node {
		nodes := n.(Nodes)
		return BindingDecl{
			Binding: nodes[2].(Ident),
			Expr:    nodes[len(nodes)-1].(Expr),
		}
	}))(input)
}

func ParseStmt(input Input) Result {
	return Wrap(Map(
		Seq(Any(ParseBindingDecl, ParseExpr), CanWS, Lit(';')),
		// The thing we care about is in the first node, the rest is
		// optional whitespace followed by a semicolon
		func(n Node) Node { return StmtFromNode(n.(Nodes)[0]) },
	))(input)
}

func ParseBlock(input Input) Result {
	return Wrap(Map(
		Seq(
			Lit('{'),
			Repeat(Seq(CanWS, ParseStmt)),
			CanWS,
			Opt(ParseExpr),
			CanWS,
			Lit('}'),
		),
		func(n Node) Node {
			nodes := n.(Nodes)

			stmtNodes := nodes[1].(Nodes)
			stmts := make([]Stmt, len(stmtNodes))
			for i, n := range stmtNodes {
				// Each node is a Seq(CanWS, ParseStmt); grab the stmt
				stmts[i] = n.(Nodes)[1].(Stmt)
			}

			exprNode := nodes[3]
			var expr *Expr
			if exprNode != nil {
				x := exprNode.(Expr)
				expr = &x
			}

			return Block{Stmts: stmts, Expr: expr}
		},
	))(input)
}

func ParseTypeExpr(input Input) Result {
	return Wrap(Map(
		ParseIdent,
		func(n Node) Node { return TypeExpr{Head: n.(Ident)} },
	))(input)
}

func ParseArgSpec(input Input) Result {
	return Wrap(Map(
		Seq(
			ParseIdent,
			CanWS,
			Opt(Map(
				Seq(Lit(':'), CanWS, ParseTypeExpr),
				// Grab only the type expr
				func(n Node) Node { return n.(Nodes)[2] },
			)),
		),
		// Grab the ident and type expr (if it exists) and turn it into an
		// ArgSpec
		func(n Node) Node {
			nodes := n.(Nodes)
			if nodes[2] != nil {
				typeExpr := nodes[2].(TypeExpr)
				return ArgSpec{Name: nodes[0].(Ident), Type: &typeExpr}
			}
			return ArgSpec{Name: nodes[0].(Ident)}
		},
	))(input)
}

func ParseArgSpecList(input Input) Result {
	return Wrap(Map(
		Opt(
			Map(
				Seq(
					// Nodes[ArgSpec]
					Repeat(
						Map(
							Seq(ParseArgSpec, CanWS, Lit(','), CanWS),
							// Grab the arg spec from the sequence
							func(n Node) Node { return n.(Nodes)[0] },
						),
					),
					// Node[ArgSpec]
					ParseArgSpec,
				),
				// Attach the first N nodes (those with trailing commas) to the
				// last node (no trailing commma)
				func(n Node) Node {
					nodes := n.(Nodes)
					// Nodes[ArgSpec]
					return append(nodes[0].(Nodes), nodes[1])
				},
			),
		),
		// If no args are specified, return an empty list of nodes
		func(n Node) Node {
			if n == nil {
				return Nodes{}
			}
			return n
		},
	))(input)
}

func ParseLambda(input Input) Result {
	return Wrap(Map(
		Seq(
			Lit('|'),         // 0
			CanWS,            // 1
			ParseArgSpecList, // 2
			CanWS,            // 3
			Lit('|'),         // 4
			CanWS,            // 5
			Map(
				Opt(Seq(CanWS, StrLit("->"), CanWS, ParseTypeExpr)),
				func(n Node) Node {
					if n != nil {
						return n.(Nodes)[3]
					}
					return nil
				},
			), // 6
			CanWS,     // 7
			ParseExpr, // 8
		),
		func(n Node) Node {
			nodes := n.(Nodes)
			argNodes := nodes[2].(Nodes)
			args := make([]ArgSpec, len(argNodes))
			for i, n := range argNodes {
				args[i] = n.(ArgSpec)
			}
			var ret *TypeExpr
			if nodes[6] != nil {
				typeExpr := nodes[6].(TypeExpr)
				ret = &typeExpr
			}
			body := nodes[8].(Expr)
			return Lambda{Args: args, Ret: ret, Body: body}
		},
	))(input)
}

func ParseFile(input Input) Result {
	pkg := Map(
		Seq(StrLit("package"), MustWS, ParseIdent, Lit('\n')),
		func(n Node) Node { return n.(Nodes)[2] },
	)
	imports := Opt(Map(
		Seq(
			StrLit("import"),
			CanWS,
			Lit('('),
			CanWS,
			Repeat(
				Map(
					Seq(ParseString, CanWS),
					// Grab the string part
					func(n Node) Node { return n.(Nodes)[0] },
				),
			),
			Lit(')'),
		),
		// Grab the import strings
		func(n Node) Node { return n.(Nodes)[4] },
	))

	return Wrap(Map(
		Seq(
			pkg, CanWS,
			imports, CanWS,
			Repeat(Map(
				Seq(ParseBindingDecl, CanWS, Lit(';'), CanWS),
				func(n Node) Node { return n.(Nodes)[0] },
			)),
			CanWS,
			Lit(rune(0)), // must eof
		),
		func(n Node) Node {
			var f File
			nodes := n.(Nodes)

			f.Package = string(nodes[0].(Ident))

			// If there are imports...
			if nodes[2] != nil {
				importNodes := nodes[2].(Nodes)
				f.Imports = make([]string, len(importNodes))
				for i, n := range importNodes {
					f.Imports[i] = string(n.(String))
				}
			}

			declNodes := nodes[4].(Nodes)
			f.Decls = make([]BindingDecl, len(declNodes))
			for i, n := range declNodes {
				f.Decls[i] = n.(BindingDecl)
			}
			return f
		},
	))(input)
}
