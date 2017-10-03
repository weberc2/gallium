package ast

type Node interface {
	node()
	EqualNode(other Node) bool
}

func (f File) node()      {}
func (td TypeDecl) node() {}

type Decl interface {
	declNode()
	EqualDecl(other Decl) bool
}

func (td TypeDecl) declNode() {}
