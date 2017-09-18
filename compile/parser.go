package main

import (
	"fmt"
	"strings"
	"unicode"
)

type Err struct {
	ParserName string
	Context    Input
	Err        error
}

func (e Err) Pretty(indent int) string {
	var errStr string
	if err, ok := e.Err.(Err); ok {
		errStr = err.Pretty(indent + 1)
	} else {
		errStr = "\n" + strings.Repeat("    ", indent+1) + e.Err.Error()
	}
	ctx := e.Context.Take(10)
	return "\n" + strings.Repeat("    ", indent) +
		e.ParserName + "(\"" + ctx + "...\"): " + errStr
}

func (e Err) Error() string {
	return e.Pretty(0)
}

func strLit(s string) func(input Input) (string, Input, error) {
	return func(input Input) (string, Input, error) {
		orig := input
		for i, r := range []rune(s) {
			if r != input.Value() {
				return "", orig, Err{
					ParserName: "StrLit[\"" + s + "\"]",
					Context:    input,
					Err: fmt.Errorf(
						"Wanted '%s'; got '%s'",
						s,
						string(append([]rune(s)[:i], input.Value())),
					),
				}
			}
			input = input.Next()
		}
		return s, input, nil
	}
}

func parseIdent(input Input) (string, Input, error) {
	start := input
	runes := []rune{}
	if r := input.Value(); r == '_' || unicode.IsLetter(r) {
		runes = append(runes, r)
		input = input.Next()

		for {
			r := input.Value()
			if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				runes = append(runes, r)
				input = input.Next()
				continue
			}
			break
		}

		return string(runes), input, nil
	}
	return "", start, Err{
		ParserName: "parseIdent",
		Context:    start,
		Err: fmt.Errorf(
			"Identifier must begin with letter or '_'; got '%s'",
			string(input.Value()),
		),
	}
}

func parseInt(input Input) (int, Input, error) {
	start := input
	var acc int
	var oneDigit bool
	for r := input.Value(); unicode.IsDigit(r); r = input.Value() {
		acc = 10*acc + int(r-'0')
		input = input.Next()
		oneDigit = true
	}
	if oneDigit {
		return acc, input, nil
	}
	return 0, start, Err{
		ParserName: "parseInt",
		Context:    start,
		Err: fmt.Errorf(
			"Wanted digit; got '%s'",
			string(start.Value()),
		),
	}
}

func parseStr(input Input) (string, Input, error) {
	start := input
	if input.Value() != '"' {
		return "", input, Err{
			ParserName: "parseStr",
			Context:    input,
			Err:        fmt.Errorf("Wanted '\"'; got '%s'", input.Value()),
		}
	}
	input = input.Next()
	runes := []rune{}
	for input.More() {
		if r := input.Value(); r != '"' {
			runes = append(runes, r)
			input = input.Next()
			continue
		}
		return string(runes), input.Next(), nil
	}
	return "", start, Err{
		ParserName: "parseStr",
		Context:    start,
		Err:        fmt.Errorf("Encountered EOF before closing quote"),
	}
}

func parseUnOp(input Input) (rune, Input, error) {
	for _, r := range []rune{'-', '*', '!'} {
		if input.Value() == r {
			return r, input.Next(), nil
		}
	}
	return rune(0), input, Err{
		ParserName: "parseUnOp",
		Context:    input,
		Err: fmt.Errorf(
			"Unable to match unary operator; wanted ('-'|'*'|'!'), got %s",
			string(input.Value()),
		),
	}
}

func parseBinOp(input Input) (string, Input, error) {
	op, rest, err := parseArithOp(input)
	if err != nil {
		op, rest, err := parseCmpOp(input)
		if err != nil {
			return "", input, Err{
				ParserName: "parseBinOp",
				Context:    input,
				Err:        fmt.Errorf("Unable to parse binary operator"),
			}
		}
		return op, rest, nil
	}
	return op, rest, nil
}

func parseArithOp(input Input) (string, Input, error) {
	for _, r := range []rune{'+', '-', '*', '/', '%'} {
		if input.Value() == r {
			return string(r), input.Next(), nil
		}
	}
	return "", input, Err{
		ParserName: "parseArithOp",
		Context:    input,
		Err:        fmt.Errorf("Unable to parse arithmetic operator"),
	}
}

func parseCmpOp(input Input) (string, Input, error) {
	stringParsers := []func(Input) (string, Input, error){
		strLit("=="),
		strLit("!="),
		strLit(">="),
		strLit("<="),
		strLit(">"),
		strLit("<"),
	}
	for _, parse := range stringParsers {
		s, rest, err := parse(input)
		if err != nil {
			continue
		}
		return s, rest, nil
	}
	return "", input, Err{
		ParserName: "parseCmpOp",
		Context:    input,
		Err:        fmt.Errorf("Unable to parse comparison operator"),
	}
}

func parseUnExpr(input Input) (UnExpr, Input, error) {
	unop, rest, err := parseUnOp(input)
	if err != nil {
		return UnExpr{}, input, err
	}

	expr, rest, err := parseAtom(rest)
	if err != nil {
		return UnExpr{}, input, err
	}

	return UnExpr{Operator: unop, Operand: expr}, rest, nil
}

func parseAtom(input Input) (Expr, Input, error) {
	for _, parse := range []func(input Input) (Expr, Input, error){
		func(input Input) (Expr, Input, error) {
			ident, rest, err := parseIdent(input)
			return ExprIdent(ident), rest, err
		},
		func(input Input) (Expr, Input, error) {
			i, rest, err := parseInt(input)
			return ExprInt(i), rest, err
		},
		func(input Input) (Expr, Input, error) {
			s, rest, err := parseStr(input)
			return ExprStr(s), rest, err
		},
	} {
		expr, rest, err := parse(input)
		if err != nil {
			continue
		}
		return expr, rest, nil
	}
	return Expr{}, input, Err{
		ParserName: "parseAtom",
		Context:    input,
		Err:        fmt.Errorf("Failed to parse atom"),
	}
}

func parseExprStmt(input Input) (Expr, Input, error) {
	expr, rest, err := parseExpr(input)
	if err != nil {
		return Expr{}, input, Err{
			ParserName: "parseExprStmt",
			Context:    input,
			Err:        err,
		}
	}
	rest = skipWS(rest)
	if rest.Value() != ';' {
		return Expr{}, input, Err{
			ParserName: "parseExprStmt",
			Context:    input,
			Err: fmt.Errorf(
				"Wanted ';'; got '%s'",
				string(rest.Value()),
			),
		}
	}
	return expr, rest.Next(), nil
}

func parseLetStmt(input Input) (LetStmt, Input, error) {
	_, rest, err := strLit("let")(input)
	if err != nil {
		return LetStmt{}, input, Err{
			ParserName: "parseLetStmt",
			Context:    input,
			Err:        err,
		}
	}

	binding, rest, err := parseIdent(skipWS(rest))
	if err != nil {
		return LetStmt{}, input, Err{
			ParserName: "parseLetStmt",
			Context:    input,
			Err:        err,
		}
	}

	rest = skipWS(rest)
	if rest.Value() != '=' {
		return LetStmt{}, input, Err{
			ParserName: "parseLetStmt",
			Context:    input,
			Err:        fmt.Errorf("Missing '=' after 'let'"),
		}
	}

	expr, rest, err := parseExpr(skipWS(rest.Next()))
	if err != nil {
		return LetStmt{}, input, Err{
			ParserName: "parseLetStmt",
			Context:    input,
			Err:        err,
		}
	}

	rest = skipWS(rest)
	if rest.Value() != ';' {
		return LetStmt{}, input, Err{
			ParserName: "parseLetStmt",
			Context:    input,
			Err: fmt.Errorf(
				"Wanted ';'; got '%s'",
				string(rest.Value()),
			),
		}
	}

	return LetStmt{Binding: binding, Expr: expr}, rest.Next(), nil
}

func parseStmt(input Input) (Stmt, Input, error) {
	letStmt, rest, err := parseLetStmt(input)
	if err == nil {
		return StmtLetStmt(letStmt), rest, err
	}
	expr, rest, err := parseExprStmt(input)
	if err == nil {
		return StmtExpr(expr), rest, err
	}
	return Stmt{}, input, Err{
		ParserName: "parseStmt",
		Context:    input,
		Err:        fmt.Errorf("Failed to parse statement"),
	}
}

func parseBlockExpr(input Input) (BlockExpr, Input, error) {
	if input.Value() != '{' {
		return BlockExpr{}, input, Err{
			Context:    input,
			ParserName: "parseBlockExpr",
			Err: fmt.Errorf(
				"Wanted '{', got '%s'",
				string(input.Value()),
			),
		}
	}
	start := input
	input = input.Next()

	var be BlockExpr
	for {
		statement, rest, err := parseStmt(skipWS(input))
		if err != nil {
			break
		}
		be.Stmts = append(be.Stmts, statement)
		input = rest
	}

	expr, rest, err := parseExpr(skipWS(input))
	if err == nil {
		be.Expr = &expr
	}

	rest, err = lit('}', skipWS(rest))
	if err != nil {
		return BlockExpr{}, start, Err{"parseBlockExpr", start, err}
	}
	return be, rest, nil
}

func lit(r rune, input Input) (Input, error) {
	if input.Value() == r {
		return input.Next(), nil
	}
	return input, fmt.Errorf(
		"Wanted '%s'; got '%s'",
		string(r),
		string(input.Value()),
	)
}

func parseArgList(input Input) ([]Expr, Input, error) {
	rest, err := lit('(', input)
	if err != nil {
		return nil, input, Err{"parseCallExpr", input, err}
	}

	rest, err = lit(')', skipWS(rest))
	if err == nil {
		return []Expr{}, rest, nil
	}

	arg, rest, err := parseExpr(skipWS(rest))
	if err != nil {
		return nil, input, Err{"parseCallExpr", input, err}
	}

	arglist := []Expr{arg}
	for {
		rest, err = lit(')', skipWS(rest))
		if err == nil {
			return arglist, rest, nil
		}

		rest, err = lit(',', skipWS(rest))
		if err != nil {
			return nil, input, Err{"parseCallExpr", input, err}
		}

		var arg Expr
		arg, rest, err = parseExpr(skipWS(rest))
		if err != nil {
			return nil, input, Err{"parseCallExpr", input, err}
		}

		arglist = append(arglist, arg)
	}
}

func parseTerm(input Input) (Expr, Input, error) {
	atom, rest, err := parseAtom(input)
	if err != nil {
		unexpr, rest2, err := parseUnExpr(input)
		if err != nil {
			return Expr{}, input, Err{
				ParserName: "parseTerm",
				Context:    input,
				Err:        err,
			}
		}
		atom = ExprUnExpr(&unexpr)
		rest = rest2
	}
	return atom, rest, nil
}

func parseBinExprRhs(lhs Expr, input Input) (BinExpr, Input, error) {
	binOp, rest, err := parseBinOp(input)
	if err != nil {
		return BinExpr{}, input, Err{"parseBinExprRhs", input, err}
	}

	rightExpr, rest, err := parseExpr(skipWS(rest))
	if err != nil {
		return BinExpr{}, input, Err{"parseBinExprRhs", input, err}
	}
	return BinExpr{Left: lhs, Right: rightExpr, Operator: binOp}, rest, nil
}

func parseCallExprRhs(lhs Expr, input Input) (CallExpr, Input, error) {
	args, rest, err := parseArgList(input)
	if err != nil {
		return CallExpr{}, input, Err{"parseCallExprRhs", input, err}
	}
	return CallExpr{lhs, args}, rest, nil
}

func parseExprRhs(lhs Expr, input Input) (Expr, Input, error) {
	callExpr, rest, err := parseCallExprRhs(lhs, input)
	if err == nil {
		// try to parse more
		expr2, rest2, err := parseExprRhs(ExprCallExpr(&callExpr), rest)

		// if successful, return the next part of the error
		if err == nil {
			return expr2, rest2, nil
		}

		// if unsuccessful, return the original
		return ExprCallExpr(&callExpr), rest, nil
	}

	binExpr, rest, err := parseBinExprRhs(lhs, input)
	if err == nil {
		return ExprBinExpr(&binExpr), rest, nil
	}

	return Expr{}, input, Err{
		"parseExprRhs",
		input,
		fmt.Errorf("Failed to parse expr"),
	}
}

func parseExpr(input Input) (Expr, Input, error) {
	be, rest, err := parseBlockExpr(input)
	if err == nil {
		return ExprBlockExpr(be), rest, nil
	}

	// Parse a term (unary, literal, ident, etc). This is either the entire
	// expression or the start of some left-recursive expression.
	term, rest, err := parseTerm(input)
	if err != nil {
		return Expr{}, input, Err{"parseExpr", input, err}
	}

	expr, rest, err := parseExprRhs(term, skipWS(rest))
	if err == nil {
		return expr, rest, nil
	}

	return term, rest, nil
}

func parseArgDecl(input Input) (ArgDecl, Input, error) {
	name, rest, err := parseIdent(input)
	if err != nil {
		return ArgDecl{}, input, Err{
			ParserName: "parseArgDecl",
			Context:    input,
			Err: fmt.Errorf(
				"Unable to parse argument declaration; malformed identifier",
			),
		}
	}

	if rest.Value() != ':' {
		return ArgDecl{}, input, Err{
			ParserName: "parseArgDecl",
			Context:    input,
			Err: fmt.Errorf(
				"Unable to parse argument declaration; missing ':' after " +
					"identifier",
			),
		}
	}

	type_, rest, err := parseIdent(skipWS(rest.Next()))
	if err != nil {
		return ArgDecl{}, input, Err{
			ParserName: "parseArgDecl",
			Context:    input,
			Err:        err,
		}
	}

	return ArgDecl{Name: name, Type: Type(type_)}, rest, nil
}

func parseArgDeclList(input Input) ([]ArgDecl, Input, error) {
	start := input
	if input.Value() != '(' {
		return nil, input, Err{
			ParserName: "parseArgDeclList",
			Context:    input,
			Err: fmt.Errorf(
				"Unable to parse argument declaration list; " +
					"missing leading parenthesis",
			),
		}
	}

	if input.Next().Value() == ')' {
		return nil, input.Next().Next(), nil
	}

	decl, rest, err := parseArgDecl(input.Next())
	if err != nil {
		return nil, input, Err{
			ParserName: "parseArgDeclList",
			Context:    input,
			Err:        err,
		}
	}

	argDecls := []ArgDecl{decl}
	for {
		if rest.Value() == ')' {
			return argDecls, rest.Next(), nil
		}

		if rest.Value() != ',' {
			return nil, input, Err{
				ParserName: "parseArgDeclList",
				Context:    input,
				Err: fmt.Errorf(
					"Unable to parse argument declaration list; expected comma",
				),
			}
		}

		decl, rest, err = parseArgDecl(rest.Next())
		if err != nil {
			return nil, start, Err{
				ParserName: "parseArgDeclList",
				Context:    start,
				Err:        err,
			}
		}
		argDecls = append(argDecls, decl)
	}
}

func skipWS(input Input) Input {
	for {
		if r := input.Value(); !unicode.IsSpace(r) {
			return input
		}
		input = input.Next()
	}
}

func parseFuncDecl(input Input) (FuncDecl, Input, error) {
	_, rest, err := strLit("fn ")(input)
	if err != nil {
		return FuncDecl{}, input, Err{"parseFuncDecl", input, err}
	}

	name, rest, err := parseIdent(rest)
	if err != nil {
		return FuncDecl{}, input, Err{"parseFuncDecl", input, err}
	}

	argDecls, rest, err := parseArgDeclList(rest)
	if err != nil {
		return FuncDecl{}, input, Err{"parseFuncDecl", input, err}
	}

	// Parse the arrow; if the arrow is missing, try parsing the body (implicit
	// return type is unit)
	_, rest, err = strLit("->")(skipWS(rest))
	if err != nil {
		// If we don't have an arrow, the return type is () or there is an
		// error
		body, rest, err := parseExpr(skipWS(rest))
		if err != nil {
			return FuncDecl{}, input, Err{"parseFuncDecl", input, err}
		}
		return FuncDecl{
			Name: name,
			Args: argDecls,
			Body: body,
		}, rest, nil
	}

	// Parse the return type
	ret, rest, err := parseIdent(skipWS(rest))
	if err != nil {
		return FuncDecl{}, input, Err{"parseFuncDecl", input, err}
	}

	// Parse the body
	body, rest, err := parseExpr(skipWS(rest))
	if err != nil {
		return FuncDecl{}, input, Err{"parseFuncDecl", input, err}
	}

	return FuncDecl{
		Name: name,
		Args: argDecls,
		Body: body,
		Ret:  Type(ret),
	}, rest, nil
}

func parsePackage(input Input) (string, Input, error) {
	_, rest, err := strLit("package")(input)
	if err != nil {
		return "", input, Err{"parsePackage", input, err}
	}
	pkg, rest, err := parseIdent(skipWS(rest))
	if err != nil {
		return "", input, Err{"parsePackage", input, err}
	}
	return pkg, rest, nil
}

func parseImports(input Input) ([]string, Input, error) {
	_, rest, err := strLit("import")(input)
	if err != nil {
		return nil, input, Err{"parseImports", input, err}
	}
	rest, err = lit('(', skipWS(rest))
	if err != nil {
		return nil, input, Err{"parseImports", input, err}
	}
	var imports []string
	for {
		// If we hit a closing parens, return imports. Otherwise keep parsing
		// imports.
		if rest, err = lit(')', skipWS(rest)); err == nil {
			return imports, rest, nil
		}
		var imp string
		imp, rest, err = parseStr(skipWS(rest))
		if err != nil {
			return nil, input, Err{"parseImports", input, err}
		}
		imports = append(imports, imp)
	}
}

func parseFile(input Input) (File, Input, error) {
	pkg, rest, err := parsePackage(skipWS(input))
	if err != nil {
		return File{}, input, Err{
			"parseFile",
			input,
			err,
		}
	}

	// Imports are optional; ignore errors
	imports, rest, _ := parseImports(skipWS(rest))

	var decls []FuncDecl
	for {
		var decl FuncDecl
		decl, rest, err = parseFuncDecl(skipWS(rest))
		if err != nil {
			if rest.Value() == rune(0) {
				return File{pkg, imports, decls}, rest, nil
			}
			return File{}, input, Err{"parseFile", input, err}
		}
		decls = append(decls, decl)
	}
}
