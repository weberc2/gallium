package infer

import (
	"fmt"

	"github.com/kr/pretty"
	"github.com/weberc2/gallium/ast"
)

type Environment map[ast.Ident]ast.Type

func (e Environment) Copy() Environment {
	e2 := make(Environment, len(e))
	for i, t := range e {
		e2[i] = t
	}
	return e2
}

func (e Environment) Add(ident ast.Ident, t ast.Type) Environment {
	e2 := e.Copy()
	e2[ident] = t
	return e2
}

var r = 'a' - 1

func genNewType() ast.Type {
	r += 1
	return ast.TypeVar(r)
}

func AnnotateExpr(expr ast.Expr, env Environment) (ast.Expr, error) {
	switch node := expr.Node.(type) {
	case ast.IntLit:
		return ast.Expr{Type: ast.Primitive("int"), Node: node}, nil
	case ast.StringLit:
		return ast.Expr{Type: ast.Primitive("string"), Node: node}, nil
	case ast.Ident:
		if t, found := env[node]; found {
			return ast.Expr{Type: t, Node: node}, nil
		}
		return ast.Expr{}, fmt.Errorf("Unknown identifier: '%s'", node)
	case ast.TupleLit:
		var err error
		out := make(ast.TupleLit, len(node))
		ts := make(ast.TupleSpec, len(node))
		for i, expr := range node {
			out[i], err = AnnotateExpr(expr, env)
			if err != nil {
				return ast.Expr{}, err
			}
			ts[i] = out[i].Type
		}
		return ast.Expr{Type: ts, Node: out}, nil
	case ast.Block:
		for _, stmt := range node.Stmts {
			if letDecl, ok := stmt.(ast.LetDecl); ok {
				binding, err := Infer(env, letDecl.Binding)
				if err != nil {
					return ast.Expr{}, err
				}
				env = env.Add(letDecl.Ident, binding.Type)
			}
		}
		inner, err := AnnotateExpr(node.Expr, env)
		if err != nil {
			return ast.Expr{}, err
		}
		return ast.Expr{
			Type: genNewType(),
			Node: ast.Block{Stmts: node.Stmts, Expr: inner},
		}, nil
	case ast.FuncLit:
		newEnv := env.Add(node.Arg, genNewType())
		body, err := AnnotateExpr(node.Body, newEnv)
		if err != nil {
			return ast.Expr{}, err
		}
		return ast.Expr{
			Type: ast.FuncSpec{Arg: newEnv[node.Arg], Ret: genNewType()},
			Node: ast.FuncLit{Arg: node.Arg, Body: body},
		}, nil
	case ast.Call:
		fn, err := AnnotateExpr(node.Fn, env)
		if err != nil {
			return ast.Expr{}, err
		}
		arg, err := AnnotateExpr(node.Arg, env)
		if err != nil {
			return ast.Expr{}, err
		}
		return ast.Expr{
			Type: genNewType(),
			Node: ast.Call{Fn: fn, Arg: arg},
		}, nil
	default:
		panic(fmt.Sprintf(
			"Invalid expr node: %# v",
			pretty.Formatter(expr.Node),
		))
	}
}

type Constraint struct {
	L, R ast.Type
}

func CollectExpr(expr ast.Expr) ([]Constraint, error) {
	switch node := expr.Node.(type) {
	case ast.IntLit, ast.StringLit:
		return nil, nil // No constraints to impose on literals
	case ast.Ident:
		return nil, nil // single occurence of ident gives no info
	case ast.TupleLit:
		var constraints []Constraint
		for _, expr := range node {
			cs, err := CollectExpr(expr)
			if err != nil {
				return nil, err
			}
			constraints = append(constraints, cs...)
		}
		return constraints, nil
	case ast.Block:
		return CollectExpr(node.Expr)
	case ast.FuncLit:
		if spec, isFunc := expr.Type.(ast.FuncSpec); isFunc {
			bodyConstraints, err := CollectExpr(node.Body)
			if err != nil {
				return nil, err
			}
			return append(
				bodyConstraints,
				Constraint{node.Body.Type, spec.Ret},
			), nil
		}
		return nil, fmt.Errorf(
			"Not a function: %# v",
			pretty.Formatter(expr),
		)
	case ast.Call:
		switch t := expr.Type.(type) {
		case ast.FuncSpec:
			fnConstraints, err := CollectExpr(node.Fn)
			if err != nil {
				return nil, err
			}
			argConstraints, err := CollectExpr(node.Arg)
			if err != nil {
				return nil, err
			}
			return append(
				append(fnConstraints, argConstraints...),
				Constraint{t, t.Ret},
				Constraint{t.Arg, t.Arg},
			), nil
		case ast.TypeVar:
			fnConstraints, err := CollectExpr(node.Fn)
			if err != nil {
				return nil, err
			}
			argConstraints, err := CollectExpr(node.Arg)
			if err != nil {
				return nil, err
			}
			return append(
				append(fnConstraints, argConstraints...),
				Constraint{
					node.Fn.Type,
					ast.FuncSpec{node.Arg.Type, expr.Type},
				},
			), nil
		default:
			panic(pretty.Sprint("Unexpected expr type:", expr.Type))
		}
	default:
		panic(fmt.Sprintf("Invalid expr node: %# v", pretty.Formatter(node)))
	}
}

type Substitution struct {
	Var  ast.TypeVar
	Type ast.Type
}

func (s Substitution) Equal(other Substitution) bool {
	return s.Var == other.Var && s.Type.EqualType(other.Type)
}

func Unify(constraints []Constraint) ([]Substitution, error) {
	if len(constraints) < 1 {
		return nil, nil
	}
	t2, err := Unify(constraints[1:])
	if err != nil {
		return nil, err
	}
	t1, err := UnifyOne(Apply(t2, constraints[0].L), Apply(t2, constraints[0].R))
	if err != nil {
		return nil, err
	}
	return append(t1, t2...), nil
}

func UnifyOne(t1, t2 ast.Type) ([]Substitution, error) {
	// Check for matching primitives
	if p1, ok := t1.(ast.Primitive); ok {
		if p2, ok := t2.(ast.Primitive); ok {
			if p1 == p2 {
				return nil, nil
			}
		}
	}
	if tv, ok := t1.(ast.TypeVar); ok {
		// if both types are the same typevar, don't substitute
		if tv2, ok := t2.(ast.TypeVar); ok && tv == tv2 {
			return nil, nil
		}
		return []Substitution{{tv, t2}}, nil
	}
	if tv, ok := t2.(ast.TypeVar); ok {
		return []Substitution{{tv, t1}}, nil
	}
	if spec1, ok := t1.(ast.FuncSpec); ok {
		if spec2, ok := t2.(ast.FuncSpec); ok {
			return Unify([]Constraint{
				{spec1.Arg, spec2.Arg},
				{spec1.Ret, spec2.Ret},
			})
		}
	}
	if ts1, ok := t1.(ast.TupleSpec); ok {
		if ts2, ok := t2.(ast.TupleSpec); ok {
			if len(ts1) == len(ts2) {
				constraints := make([]Constraint, len(ts1))
				for i, t := range ts1 {
					constraints[i] = Constraint{t, ts2[i]}
				}
				return Unify(constraints)
			}
		}
	}
	return nil, fmt.Errorf("Mismatched types: %v != %v", t1, t2)
}

func Substitute(replace ast.Type, tv ast.TypeVar, t ast.Type) ast.Type {
	switch typ := t.(type) {
	case ast.Primitive:
		return t
	case ast.TypeVar:
		if typ == tv {
			return replace
		}
		return t
	case ast.FuncSpec:
		return ast.FuncSpec{
			Arg: Substitute(replace, tv, typ.Arg),
			Ret: Substitute(replace, tv, typ.Ret),
		}
	case ast.TypeRef:
		// TODO: It's not currently clear where this fits into the reference
		// implementation; maybe this is made redundant by TypeVar?
		panic("TypeRef not supported")
	case ast.TupleSpec:
		out := make(ast.TupleSpec, len(typ))
		for i, t := range typ {
			out[i] = Substitute(replace, tv, t)
		}
		return out
	default:
		panic(fmt.Sprintf(
			"Substitute() not implemented for %# v",
			pretty.Formatter(t),
		))
	}
}

func Apply(subs []Substitution, t ast.Type) ast.Type {
	// iterate right-to-left
	for i := len(subs) - 1; i >= 0; i-- {
		t = Substitute(subs[i].Type, subs[i].Var, t)
	}

	return t
}

func ApplyExpr(subs []Substitution, expr ast.Expr) ast.Expr {
	switch node := expr.Node.(type) {
	case ast.IntLit:
		return ast.Expr{Node: node, Type: Apply(subs, expr.Type)}
	case ast.StringLit:
		return ast.Expr{Node: node, Type: Apply(subs, expr.Type)}
	case ast.Ident:
		return ast.Expr{Node: node, Type: Apply(subs, expr.Type)}
	case ast.TupleLit:
		tl := make(ast.TupleLit, len(node))
		for i, expr := range node {
			tl[i] = ApplyExpr(subs, expr)
		}
		return ast.Expr{Node: tl, Type: Apply(subs, expr.Type)}
	case ast.Block:
		inner := ApplyExpr(subs, node.Expr)
		return ast.Expr{
			Type: inner.Type,
			Node: ast.Block{Stmts: node.Stmts, Expr: inner},
		}
	case ast.FuncLit:
		return ast.Expr{
			Node: ast.FuncLit{
				Arg:  node.Arg,
				Body: ApplyExpr(subs, node.Body),
			},
			Type: Apply(subs, expr.Type),
		}
	case ast.Call:
		return ast.Expr{
			Node: ast.Call{
				Fn:  ApplyExpr(subs, node.Fn),
				Arg: ApplyExpr(subs, node.Arg),
			},
			Type: Apply(subs, expr.Type),
		}
	default:
		panic(fmt.Sprintf(
			"ApplyExpr() not implemented for %# v",
			pretty.Formatter(expr.Node),
		))
	}
}

func Infer(env Environment, expr ast.Expr) (ast.Expr, error) {
	annotated, err := AnnotateExpr(expr, env)
	r = 'a' - 1
	if err != nil {
		return ast.Expr{}, err
	}
	constraints, err := CollectExpr(annotated)
	if err != nil {
		return ast.Expr{}, err
	}
	subs, err := Unify(constraints)
	if err != nil {
		return ast.Expr{}, err
	}
	return ApplyExpr(subs, annotated), nil
}
