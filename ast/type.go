package ast

import (
	"strconv"
	"strings"
)

type Primitive string

func (p Primitive) RenderGo() string { return string(p) }

func (p Primitive) RenderGoIdent() string { return string(p) }

func (p Primitive) Replace(map[TypeVar]Type) Type { return p }

func (p Primitive) EqualType(other Type) bool {
	otherPrimitive, ok := other.(Primitive)
	return ok && p == otherPrimitive
}

func (p Primitive) RenderGoLit(tr TypeRef) string { return tr.RenderGoIdent() }

func (p Primitive) String() string {
	return string(p)
}

type ArgSpec struct {
	Name string
	Type Type // optional
}

func (as ArgSpec) Equal(other ArgSpec) bool {
	if as.Type != nil {
		return as.Name == other.Name && as.Type.EqualType(other.Type)
	}
	return as.Name == other.Name && other.Type == nil
}

func (as ArgSpec) EqualDecl(other Decl) bool {
	otherArgSpec, ok := other.(ArgSpec)
	return ok && as.Equal(otherArgSpec)
}

func (as ArgSpec) RenderGo() string {
	return as.Name + " " + as.Type.RenderGo()
}

func (as ArgSpec) String() string {
	return as.Name + " " + as.Type.String()
}

type TupleSpec []Type

func (ts TupleSpec) Equal(other TupleSpec) bool {
	if len(ts) != len(other) {
		return false
	}
	for i, t := range ts {
		if !t.EqualType(other[i]) {
			return false
		}
	}
	return true
}

func (ts TupleSpec) EqualType(other Type) bool {
	otherTupleSpec, ok := other.(TupleSpec)
	return ok && ts.Equal(otherTupleSpec)
}

func (ts TupleSpec) Replace(types map[TypeVar]Type) Type {
	ts2 := make(TupleSpec, len(ts))
	for i, t := range ts {
		ts2[i] = t.Replace(types)
	}
	return ts2
}

func (ts TupleSpec) RenderGo() string {
	args := make([]string, len(ts))
	for i, t := range ts {
		args[i] = "_" + strconv.Itoa(i) + " " + t.RenderGo()
	}
	return "struct{" + strings.Join(args, "; ") + "}"
}

func (ts TupleSpec) RenderGoIdent() string {
	typeStrings := make([]string, len(ts))
	for i, t := range ts {
		typeStrings[i] = t.RenderGoIdent()
	}
	return "Tuple" + pi + strings.Join(typeStrings, pi)
}

func (ts TupleSpec) RenderGoLit(tr TypeRef) string {
	return tr.RenderGoIdent()
}

func (ts TupleSpec) String() string {
	strs := make([]string, len(ts))
	for i, t := range ts {
		strs[i] = t.String()
	}
	return "(" + strings.Join(strs, ", ") + ")"
}

type FuncSpec struct {
	Arg Type
	Ret Type
}

func (fs FuncSpec) Equal(other FuncSpec) bool {
	return fs.Arg.EqualType(other.Arg) && fs.Ret.EqualType(other.Ret)
}

func (fs FuncSpec) EqualType(other Type) bool {
	otherFuncSpec, ok := other.(FuncSpec)
	return ok && fs.Equal(otherFuncSpec)
}

func (fs FuncSpec) Replace(types map[TypeVar]Type) Type {
	return FuncSpec{
		fs.Arg.Replace(types),
		fs.Ret.Replace(types),
	}
}

func (fs FuncSpec) RenderGo() string {
	out := "func(" + fs.Arg.RenderGo()
	ret := fs.Ret
	for {
		if r, ok := ret.(FuncSpec); ok {
			out += ", " + r.Arg.RenderGo()
			ret = r.Ret
			continue
		}
		break
	}
	return out + ") " + ret.RenderGo()
}

func (fs FuncSpec) RenderGoIdent() string {
	return strings.Join(
		[]string{
			"func",
			fs.Arg.RenderGoIdent(),
			fs.Ret.RenderGoIdent(),
		},
		beta,
	)
}

func (fs FuncSpec) RenderGoLit(tr TypeRef) string { return fs.RenderGo() }

func (fs FuncSpec) String() string {
	return fs.Arg.String() + " -> " + fs.Ret.String()
}

const beta = "ß"
const pi = "π"
const omega = "Ω"

type TypeRef struct {
	Name string
	Decl *TypeDecl
	Arg  Type
}

func (tr TypeRef) Equal(other TypeRef) bool {
	if tr.Name != other.Name {
		return false
	}
	if tr.Arg != nil {
		if !tr.Arg.EqualType(other.Arg) {
			return false
		}
	} else {
		if other.Arg != nil {
			return false
		}
	}
	if tr.Decl != nil && other.Decl != nil {
		return tr.Decl.Equal(*other.Decl)
	}
	return tr.Decl == other.Decl
}

func (tr TypeRef) EqualType(other Type) bool {
	otherTypeRef, ok := other.(TypeRef)
	return ok && tr.Equal(otherTypeRef)
}

func (tr TypeRef) Replace(types map[TypeVar]Type) Type {
	return TypeRef{Decl: tr.Decl, Arg: tr.Arg.Replace(types)}
}

func (tr TypeRef) RenderGo() string {
	return tr.RenderGoLit(tr)
}

func (tr TypeRef) RenderGoLit(TypeRef) string {
	types := map[TypeVar]Type{}
	tr2 := tr
	for _, v := range tr.Decl.Args {
		types[v] = tr2.Arg
		if tr, ok := tr2.Arg.(TypeRef); ok {
			tr2 = tr
		}
	}
	return tr.Decl.Type.Replace(types).RenderGoLit(tr)
}

func (tr TypeRef) String() string {
	if tr.Arg == nil {
		return tr.Name
	}
	return tr.Name + " " + tr.Arg.String()
}

func (p Primitive) Visit(tv TypeVisitor) {
	tv.VisitPrimitive(p)
}
func (fs FuncSpec) Visit(tv TypeVisitor) {
	tv.VisitFuncSpec(fs)
}
func (ts TupleSpec) Visit(tv TypeVisitor) {
	tv.VisitTupleSpec(ts)
}
func (tr TypeRef) Visit(tv TypeVisitor) {
	tv.VisitTypeRef(tr)
}
func (tv TypeVar) Visit(tvis TypeVisitor) {
	tvis.VisitTypeVar(tv)
}

type TypeVisitor interface {
	VisitPrimitive(p Primitive)
	VisitFuncSpec(fs FuncSpec)
	VisitTupleSpec(ts TupleSpec)
	VisitTypeRef(tr TypeRef)
	VisitTypeVar(tv TypeVar)
}

func (tr TypeRef) RenderGoIdent() string {
	if tr.Arg == nil {
		return tr.Decl.Name
	}
	// TODO: This is likely incorrect
	return tr.Decl.Name + beta + tr.Arg.RenderGoIdent()
}

type TypeDecl struct {
	Name string
	Type Type
	Args []TypeVar
}

func (td TypeDecl) EqualDecl(other Decl) bool {
	otherTypeDecl, ok := other.(TypeDecl)
	return ok && td.Equal(otherTypeDecl)
}

func (td TypeDecl) EqualNode(other Node) bool {
	otherTypeDecl, ok := other.(TypeDecl)
	return ok && td.Equal(otherTypeDecl)
}

func (td TypeDecl) EqualStmt(other Stmt) bool {
	otherTypeDecl, ok := other.(TypeDecl)
	return ok && td.Equal(otherTypeDecl)
}

func (td TypeDecl) Equal(other TypeDecl) bool {
	if td.Name != other.Name ||
		!td.Type.EqualType(other.Type) ||
		len(td.Args) != len(other.Args) {
		return false
	}
	for i, arg := range td.Args {
		if arg != other.Args[i] {
			return false
		}
	}
	return true
}

func (td TypeDecl) RenderGo() string {
	return "type " + td.Name + " " + td.Type.RenderGo()
}

type TypeVar string

func (tv TypeVar) EqualType(other Type) bool {
	otherTypeVar, ok := other.(TypeVar)
	return ok && tv == otherTypeVar
}

func (tv TypeVar) Replace(types map[TypeVar]Type) Type {
	if t, found := types[tv]; found {
		return t
	}
	return tv
}

func (tv TypeVar) RenderGo() string { panic("TypeVar.RenderGo()") }

func (tv TypeVar) RenderGoIdent() string { panic("TypeVar.RenderGoIdent()") }

func (tv TypeVar) RenderGoLit(tr TypeRef) string {
	panic("TypeVar.RenderGoLit()")
}

func (tv TypeVar) String() string {
	return "'" + string(tv)
}

type Type interface {
	Visit(TypeVisitor)
	RenderGo() string
	RenderGoIdent() string
	RenderGoLit(TypeRef) string
	Replace(map[TypeVar]Type) Type
	EqualType(other Type) bool
	String() string
}
