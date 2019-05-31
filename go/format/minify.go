package format

import (
	"github.com/anaminus/luasyntax/go/extend"
	"github.com/anaminus/luasyntax/go/token"
	"github.com/anaminus/luasyntax/go/tree"
)

type minify struct{}

func (m *minify) Visit(tree.Node) tree.Visitor {
	return m
}

func (m *minify) VisitToken(_ tree.Node, _ int, tok *tree.Token) {
	if !tok.Type.IsValid() {
		return
	}
	tok.Prefix = tok.Prefix[:0]
}

const chars = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789`
const second = len(chars)
const first = second - 10

// VarGen generates unique variable names optimized for minimum length.
type VarGen struct {
	// Strings to skip over.
	skip map[string]struct{}
	// Scratch space to hold current string.
	buf []byte
	// Current index.
	index int
	// Length of buf increments when index reaches theshold.
	threshold int
	// Previous threshold subtracted from index to realign.
	prev int
}

// NewVarGen returns an initialized VarGen. skip is a list of variable names to
// skip. Lua keywords are already skipped automatically.
func NewVarGen(skip []string) *VarGen {
	g := VarGen{
		skip:      map[string]struct{}{},
		buf:       make([]byte, 1, 3),
		index:     0,
		threshold: first,
	}
	var tok token.Type
	for ; !tok.IsValid(); tok++ {
	}
	for ; tok.IsValid(); tok++ {
		if tok.IsKeyword() {
			g.skip[tok.String()] = struct{}{}
		}
	}
	for _, s := range skip {
		g.skip[s] = struct{}{}
	}
	return &g
}

// Next generates a variable name and advances the counter.
func (g *VarGen) Next() string {
	for {
		n := g.index - g.prev
		d := n % first
		n = (n - d) / first
		g.buf[0] = chars[d]
		for i := 1; i < len(g.buf); i++ {
			d := n % second
			n = (n - d) / second
			g.buf[i] = chars[d]
		}
		s := string(g.buf)

		g.index++
		if g.index >= g.threshold {
			g.buf = append(g.buf, 0)
			g.prev = g.threshold
			p := first
			for i := 1; i < len(g.buf); i++ {
				p *= second
			}
			g.threshold += p
		}
		if _, ok := g.skip[s]; !ok {
			return s
		}
	}
}

func (g *VarGen) Reset() {
	g.buf = g.buf[:1]
	g.index = 0
	g.threshold = first
	g.prev = 0
}

func scopeContains(fs *extend.FileScope, s *extend.Scope, v *extend.Variable) bool {
	for _, item := range s.Items {
		switch item := item.(type) {
		case *tree.Token:
			if v == fs.VariableMap[item] {
				return true
			}
		case *extend.Scope:
			if scopeContains(fs, item, v) {
				return true
			}
		}
	}
	return false
}

func descendItems(items []interface{}, cb func(items []interface{}, i int, item interface{})) {
	for i, item := range items {
		cb(items, i, item)
		if scope, ok := item.(*extend.Scope); ok {
			descendItems(scope.Items, cb)
		}
	}
}

func Minify(file *tree.File) {
	fileScope := extend.BuildFileScope(file)
	type K struct {
		scope *extend.Scope
		name  string
	}
	used := map[K]*extend.Variable{}
	visited := map[*extend.Variable]struct{}{}
	g := NewVarGen(nil)

	descendItems(fileScope.Root.Items, func(items []interface{}, i int, item interface{}) {
		token, ok := item.(*tree.Token)
		if !ok {
			return
		}

		variable, ok := fileScope.VariableMap[token]
		if _, ok := visited[variable]; ok {
			return
		}
		visited[variable] = struct{}{}
		if variable.Type == extend.LocalVar {
			g.Reset()
			for {
				name := g.Next()
				v, ok := used[K{variable.Scopes[0], name}]
				if !ok || !variable.VisiblityOverlapsWith(v) {
					variable.Name = name
					name := []byte(variable.Name)
					for _, r := range variable.References {
						r.Bytes = name
					}
					break
				}
			}
		}
		used[K{variable.Scopes[0], variable.Name}] = variable

		descendItems(items[i+1:], func(_ []interface{}, _ int, item interface{}) {
			if scope, ok := item.(*extend.Scope); ok {
				if scopeContains(fileScope, scope, variable) {
					used[K{scope, variable.Name}] = variable
				}
			}
		})
	})

	var m minify
	tree.Walk(&m, file)
	tree.FixAdjoinedTokens(file)
	tree.FixTokenOffsets(file, 0)
}
