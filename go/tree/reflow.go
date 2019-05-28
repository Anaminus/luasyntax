package tree

import (
	"github.com/anaminus/luasyntax/go/token"
)

// offsetFixer maintains the current offset while fixing the token offsets of a
// parse tree.
type offsetFixer struct {
	info *token.File
	off  int
}

func (r *offsetFixer) scanNewlines(b []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			r.info.AddLine(r.off + i)
		}
	}
}

func (r *offsetFixer) Visit(Node) Visitor {
	return r
}

func (r *offsetFixer) VisitToken(_ Node, _ int, tok *Token) {
	if !tok.Type.IsValid() {
		return
	}
	for _, prefix := range tok.Prefix {
		if !prefix.Type.IsValid() {
			continue
		}
		if r.info != nil {
			r.scanNewlines(prefix.Bytes)
		}
		r.off += len(prefix.Bytes)
	}
	r.scanNewlines(tok.Bytes)
	tok.Offset = r.off
	r.off += len(tok.Bytes)
}

// FixTokenOffsets walks through a parse tree, adjusting the offset of each
// token so that it is correct for the current bytes of the token. The offset
// argument specifies the starting offset.
//
// If the node is a File, then the file's line information will be rewritten.
func FixTokenOffsets(node Node, offset int) {
	r := offsetFixer{off: offset}
	if file, ok := node.(*File); ok {
		r.info = file.Info
		r.info.ClearLines()
	}
	Walk(&r, node)
}
