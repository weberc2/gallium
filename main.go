package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/weberc2/gallium/ast"
	"github.com/weberc2/gallium/codegen"
	"github.com/weberc2/gallium/combinator"
	"github.com/weberc2/gallium/infer"
	"github.com/weberc2/gallium/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "USAGE:", os.Args[0], "<FILE>")
		os.Exit(-1)
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	result := parser.File(combinator.Input(string(data)))
	if result.Err != nil {
		fmt.Fprintln(os.Stderr, result.Err)
		os.Exit(-1)
	}

	env := infer.Environment{
		ast.Ident("add"): ast.FuncSpec{
			Arg: ast.Primitive("int"),
			Ret: ast.FuncSpec{
				Arg: ast.Primitive("int"),
				Ret: ast.Primitive("int"),
			},
		},
		ast.Ident("PrintInt"): ast.FuncSpec{
			Arg: ast.Primitive("int"),
			Ret: ast.TupleSpec{},
		},
	}

	file := result.Value.(ast.File)
	for i, stmt := range file.Stmts {
		if letDecl, ok := stmt.(ast.LetDecl); ok {
			binding, err := infer.Infer(env, letDecl.Binding)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
			env = env.Add(letDecl.Ident, binding.Type)
			file.Stmts[i] = ast.LetDecl{Ident: letDecl.Ident, Binding: binding}
		}
	}

	if err := codegen.File(file).Render(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
