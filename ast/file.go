package ast

type File struct {
	Package string
	Stmts   []Stmt
}

func (f File) Equal(other File) bool {
	if len(f.Stmts) != len(other.Stmts) {
		return false
	}
	for i, stmt := range f.Stmts {
		if !stmt.EqualStmt(other.Stmts[i]) {
			return false
		}
	}
	return f.Package == other.Package
}

func (f File) EqualNode(other Node) bool {
	otherFile, ok := other.(File)
	return ok && f.Equal(otherFile)
}
