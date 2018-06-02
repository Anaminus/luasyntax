package extend

import (
	"github.com/anaminus/luasyntax/go/ast"
	"github.com/anaminus/luasyntax/go/token"
)

// reflowVisitor maintains the current offset while the reflowing a syntax
// tree.
type reflowVisitor struct {
	info *token.File
	off  int
}

func (r *reflowVisitor) scanNewlines(b []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			r.info.AddLine(r.off + i)
		}
	}
}

func (r *reflowVisitor) Visit(ast.Node) ast.Visitor {
	return r
}

func (r *reflowVisitor) VisitToken(_ ast.Node, _ int, tok *ast.Token) {
	for _, prefix := range tok.Prefix {
		r.scanNewlines(prefix.Bytes)
		r.off += len(prefix.Bytes)
	}
	r.scanNewlines(tok.Bytes)
	tok.Offset = r.off
	r.off += len(tok.Bytes)
}

// Reflow walks through a syntax tree, adjusting the offset of each token so
// that it is correct for the current bytes of the token.
func Reflow(file *ast.File) {
	var r reflowVisitor
	file.Info.ClearLines()
	ast.Walk(&r, file)
}
