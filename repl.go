package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kr/pretty"
	"github.com/weberc2/gallium/ast"
	"github.com/weberc2/gallium/combinator"
	"github.com/weberc2/gallium/infer"
	"github.com/weberc2/gallium/parser"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	env := infer.Environment{
		"add": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("int"),
			},
		},
		"eq": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("bool"),
			},
		},
		"ne": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("bool"),
			},
		},
		"lt": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("bool"),
			},
		},
		"gt": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("bool"),
			},
		},
		"le": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("bool"),
			},
		},
		"ge": ast.FuncSpec{
			ast.Primitive("int"),
			ast.FuncSpec{
				ast.Primitive("int"),
				ast.Primitive("bool"),
			},
		},
	}

	for {
		fmt.Print(" > ")
		if !scanner.Scan() {
			break
		}

		result := combinator.Any(
			parser.LetDecl,
			parser.Expr,
		)(combinator.Input(scanner.Text()))
		if result.Err != nil {
			fmt.Println(result.Err)
			continue
		}
		switch v := result.Value.(type) {
		case ast.Expr:
			expr, err := infer.Infer(env, v)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(expr.Type.String())
			// fmt.Println(expr.Type.RenderGo())
		case ast.LetDecl:
			expr, err := infer.Infer(env, v.Binding)
			if err != nil {
				fmt.Println(err)
				continue
			}
			env[v.Ident] = expr.Type
		default:
			panic("NOT AN EXPR OR DECL: " + pretty.Sprint(result.Value))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
