package ast

type File struct {
	Package string
	Decls   []Decl
}

func (f File) Equal(other File) bool {
	if len(f.Decls) != len(other.Decls) {
		return false
	}
	for i, decl := range f.Decls {
		if !decl.EqualDecl(other.Decls[i]) {
			return false
		}
	}
	return f.Package == other.Package
}

func (f File) EqualNode(other Node) bool {
	otherFile, ok := other.(File)
	return ok && f.Equal(otherFile)
}
