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

// Reverse chars lookup table.
var charIndex = [256]int{
	// Valid characters.
	'a': 000, 'b': 001, 'c': 002, 'd': 003, 'e': 004, 'f': 005, 'g': 006, 'h': 007,
	'i': 010, 'j': 011, 'k': 012, 'l': 013, 'm': 014, 'n': 015, 'o': 016, 'p': 017,
	'q': 020, 'r': 021, 's': 022, 't': 023, 'u': 024, 'v': 025, 'w': 026, 'x': 027,
	'y': 030, 'z': 031, 'A': 032, 'B': 033, 'C': 034, 'D': 035, 'E': 036, 'F': 037,
	'G': 040, 'H': 041, 'I': 042, 'J': 043, 'K': 044, 'L': 045, 'M': 046, 'N': 047,
	'O': 050, 'P': 051, 'Q': 052, 'R': 053, 'S': 054, 'T': 055, 'U': 056, 'V': 057,
	'W': 060, 'X': 061, 'Y': 062, 'Z': 063, '_': 064, '0': 065, '1': 066, '2': 067,
	'3': 070, '4': 071, '5': 072, '6': 073, '7': 074, '8': 075, '9': 076,
	// Everything else.
	0000: -1, 0001: -1, 0002: -1, 0003: -1, 0004: -1, 0005: -1, 0006: -1, 0007: -1,
	0010: -1, 0011: -1, 0012: -1, 0013: -1, 0014: -1, 0015: -1, 0016: -1, 0017: -1,
	0020: -1, 0021: -1, 0022: -1, 0023: -1, 0024: -1, 0025: -1, 0026: -1, 0027: -1,
	0030: -1, 0031: -1, 0032: -1, 0033: -1, 0034: -1, 0035: -1, 0036: -1, 0037: -1,
	0040: -1, 0041: -1, 0042: -1, 0043: -1, 0044: -1, 0045: -1, 0046: -1, 0047: -1,
	0050: -1, 0051: -1, 0052: -1, 0053: -1, 0054: -1, 0055: -1, 0056: -1, 0057: -1,
	0072: -1, 0073: -1, 0074: -1, 0075: -1, 0076: -1, 0077: -1,
	0100: -1,
	0133: -1, 0134: -1, 0135: -1, 0136: -1,
	0140: -1,
	0173: -1, 0174: -1, 0175: -1, 0176: -1, 0177: -1,
	0200: -1, 0201: -1, 0202: -1, 0203: -1, 0204: -1, 0205: -1, 0206: -1, 0207: -1,
	0210: -1, 0211: -1, 0212: -1, 0213: -1, 0214: -1, 0215: -1, 0216: -1, 0217: -1,
	0220: -1, 0221: -1, 0222: -1, 0223: -1, 0224: -1, 0225: -1, 0226: -1, 0227: -1,
	0230: -1, 0231: -1, 0232: -1, 0233: -1, 0234: -1, 0235: -1, 0236: -1, 0237: -1,
	0240: -1, 0241: -1, 0242: -1, 0243: -1, 0244: -1, 0245: -1, 0246: -1, 0247: -1,
	0250: -1, 0251: -1, 0252: -1, 0253: -1, 0254: -1, 0255: -1, 0256: -1, 0257: -1,
	0260: -1, 0261: -1, 0262: -1, 0263: -1, 0264: -1, 0265: -1, 0266: -1, 0267: -1,
	0270: -1, 0271: -1, 0272: -1, 0273: -1, 0274: -1, 0275: -1, 0276: -1, 0277: -1,
	0300: -1, 0301: -1, 0302: -1, 0303: -1, 0304: -1, 0305: -1, 0306: -1, 0307: -1,
	0310: -1, 0311: -1, 0312: -1, 0313: -1, 0314: -1, 0315: -1, 0316: -1, 0317: -1,
	0320: -1, 0321: -1, 0322: -1, 0323: -1, 0324: -1, 0325: -1, 0326: -1, 0327: -1,
	0330: -1, 0331: -1, 0332: -1, 0333: -1, 0334: -1, 0335: -1, 0336: -1, 0337: -1,
	0340: -1, 0341: -1, 0342: -1, 0343: -1, 0344: -1, 0345: -1, 0346: -1, 0347: -1,
	0350: -1, 0351: -1, 0352: -1, 0353: -1, 0354: -1, 0355: -1, 0356: -1, 0357: -1,
	0360: -1, 0361: -1, 0362: -1, 0363: -1, 0364: -1, 0365: -1, 0366: -1, 0367: -1,
	0370: -1, 0371: -1, 0372: -1, 0373: -1, 0374: -1, 0375: -1, 0376: -1, 0377: -1,
}

const maxIdentIndex = 1<<31 - 1 // Corresponds to identifier "uTUbJb".
const maxIdentLength = 6        // Maximum length not exceeding maxIdentIndex.

// GenerateIdent generates an identifier string from a 32-bit positive signed
// integer. Any other value results in the empty string.
func GenerateIdent(index int) string {
	if index < 0 || index > maxIdentIndex {
		return ""
	}
	length := 0
	for i, t, n := 0, 0, index; i <= n; length++ {
		index -= t
		t = first
		for j := 0; j < length; j++ {
			t *= second
		}
		t = t
		i += t
	}

	n := index
	d := n % first
	n = (n - d) / first
	buf := make([]byte, length)
	buf[0] = chars[d]
	for i := 1; i < length; i++ {
		d := n % second
		n = (n - d) / second
		buf[i] = chars[d]
	}
	return string(buf)
}

func powInt(x, y int) int {
	p := 1
	for i := 0; i < y; i++ {
		p *= x
	}
	return p
}

// IdentIndex returns the index of a generated identifier. Any invalid
// identifier causes -1 to be returned.
func IdentIndex(s string) int {
	if s == "" || len(s) > maxIdentLength {
		return -1
	}

	i := 0
	n := charIndex[s[i]]
	if n < 0 || n >= first {
		return -1
	}
	if len(s) == 1 {
		return n
	}

	i++
	p := first * powInt(second, i-1)
	x := first*p/(first-1) - 1
	n += p * charIndex[s[i]]
	for i := 2; i < len(s); i++ {
		p := first * powInt(second, i-1)
		x = second * p / (second - 1)
		n += p * charIndex[s[i]]
	}
	n += x
	if n < 0 || n > maxIdentIndex {
		return -1
	}
	return n
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

	type indexKey struct {
		scope *extend.Scope
		index int
	}
	// Index used in a scope, mapped to the variable of the index.
	usedIndexes := map[indexKey]*extend.Variable{}
	// Variable mapped to an index.
	varIndexes := map[*extend.Variable]int{}

	// Eliminate conflicts with keywords by marking them as used indexes.
	var tok token.Type
	for ; !tok.IsValid(); tok++ {
	}
	for ; tok.IsValid(); tok++ {
		if !tok.IsKeyword() {
			continue
		}
		index := IdentIndex(tok.String())
		if index < 0 {
			continue
		}
		descendItems(fileScope.Root.Items, func(_ []interface{}, _ int, item interface{}) {
			if scope, ok := item.(*extend.Scope); ok {
				usedIndexes[indexKey{scope, index}] = nil
			}
		})
	}

	// Traverse all globals first to ensure their existence is known by all
	// other variables.
	for _, variable := range fileScope.Globals {
		index := IdentIndex(variable.Name)
		varIndexes[variable] = index
		// Mark each scope that refers to the variable.
		usedIndexes[indexKey{variable.Scopes[0], index}] = variable
		// Lifetime of a global is the entire file, so all scopes must be
		// traversed.
		descendItems(fileScope.Root.Items, func(_ []interface{}, _ int, item interface{}) {
			if scope, ok := item.(*extend.Scope); ok {
				if scopeContains(fileScope, scope, variable) {
					usedIndexes[indexKey{scope, index}] = variable
				}
			}
		})
	}
	// Traverse local variables.
	descendItems(fileScope.Root.Items, func(items []interface{}, i int, item interface{}) {
		token, ok := item.(*tree.Token)
		if !ok {
			return
		}

		variable := fileScope.VariableMap[token]
		if variable.Type != extend.LocalVar {
			return
		}

		if _, ok := varIndexes[variable]; ok {
			// Already traversed this variable.
			return
		}

		// Check each index until an available one is found.
		index := 0
		// TODO: sort locals by descending reference frequency.
		for ; ; index++ {
			v, used := usedIndexes[indexKey{variable.Scopes[0], index}]
			if !used || v != nil && !variable.VisiblityOverlapsWith(v) {
				varIndexes[variable] = index
				break
			}
		}

		// Mark each scope that refers to the variable.
		usedIndexes[indexKey{variable.Scopes[0], index}] = variable
		// Lifetime of a local starts at the current scope, after the
		// declaration of the variable.
		descendItems(items[i+1:], func(_ []interface{}, _ int, item interface{}) {
			if scope, ok := item.(*extend.Scope); ok {
				if scopeContains(fileScope, scope, variable) {
					usedIndexes[indexKey{scope, index}] = variable
				}
			}
		})
	})

	for variable, index := range varIndexes {
		if variable.Type != extend.LocalVar {
			continue
		}
		variable.Name = GenerateIdent(index)
		name := []byte(variable.Name)
		for _, r := range variable.References {
			r.Bytes = name
		}
	}

	var m minify
	tree.Walk(&m, file)
	tree.FixAdjoinedTokens(file)
	tree.FixTokenOffsets(file, 0)
}
