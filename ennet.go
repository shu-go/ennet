// Package ennet parses Emmet-like abbreviations and expands them into XML.
//
// # Limitations
//
//   - No implicit tag names (`ul>.cls` causes an error)
//   - Generates always empty-element tags (yes: <a />, no: <a></a>)
//   - (internal) each TEXT {...}, QTEXT "..." is a token, unlike attr-list [, ..., ]
package ennet

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
)

// Expand expands Emmet abbreviation in string s.
func Expand(s string) (string, error) {
	b := expandBufPool.Get().(*bytes.Buffer)
	b.Reset()
	b.WriteString(s)

	nl := NewNodeBuilder(WithPool(&nodePool))
	err := Parse(b.Bytes(), &nl)
	if err != nil {
		return "", err
	}

	result := expand(nl.Root)

	gcNodes(&nl, &nodePool, nl.Root)

	expandBufPool.Put(b)
	return result, nil
}

func gcNodes(nb *NodeBuilder, pool *sync.Pool, n *Node) {
	if pool == nil {
		return
	}

	c := n.FirstChild
	for c != nil {
		gcNodes(nb, pool, c)
		c = c.NextSibling
	}
	pool.Put(n)
}

func expand(n *Node) string {
	//debug(n.Type, n.Data)

	switch n.Type {
	case Text:
		return n.Data
	case Root, Group:
		if n.FirstChild == nil {
			return ""
		}

		b := strings.Builder{}
		b.Grow(128)
		curr := n.FirstChild
		for curr != nil {
			b.WriteString(expand(curr))
			curr = curr.NextSibling
		}

		return applyMul(b.String(), n.Mul)

	case Element:
		b := strings.Builder{}
		b.Grow(128)
		b.WriteString("<")
		b.WriteString(n.Data)
		if len(n.Attribute) > 0 {
			keys := maps.Keys(n.Attribute)
			slices.Sort(keys)
			for _, k := range keys {
				b.WriteString(" ")
				b.WriteString(k)
				b.WriteString(`="`)
				b.WriteString(strings.ReplaceAll(n.Attribute[k], `"`, `\"`))
				b.WriteString(`"`)
			}
		}

		if n.FirstChild == nil {
			b.WriteString(" />")
		} else {
			b.WriteString(">")
			curr := n.FirstChild
			for curr != nil {
				b.WriteString(expand(curr))
				curr = curr.NextSibling
			}
			b.WriteString("</")
			b.WriteString(n.Data)
			b.WriteString(">")
		}
		return applyMul(b.String(), n.Mul)

	default:

	}

	return ""
}

func applyMul(templ string, mul int) string {
	if mul <= 0 {
		return templ
	}

	result := ""

	for i := 0; i < mul; i++ {
		t := templ

		// replace all /(\$+)(@(-)?(\d+)?)?/
		for {
			start := strings.IndexByte(t, '$')
			if start == -1 {
				break
			}

			pad := 1
			minus := false
			base := 1

			// /(\$+)/
			pos := start + pad
			for {
				if len(t) <= pos || t[pos] != '$' {
					break
				}
				pad++
				pos++
			}

			// /(@(-)?(\d+)?)?/
			if len(t) > pos && t[pos] == '@' {
				pos++

				// /(-)?/
				if len(t) > pos && t[pos] == '-' {
					minus = true
					pos++
				}

				// /(\d+)?/
				base = 0
				for {
					if len(t) <= pos || t[pos] < '0' || t[pos] > '9' {
						break
					}

					base = base*10 + int(t[pos]-'0')
					pos++
				}
				if base == 0 {
					base = 1
				}
			}

			if !minus {
				t = t[:start] + fmt.Sprintf("%0"+strconv.Itoa(pad)+"d", base+i) + t[pos:]
			} else {
				t = t[:start] + fmt.Sprintf("%0"+strconv.Itoa(pad)+"d", base+mul-1-i) + t[pos:]
			}
		}

		result += t
	}
	return result
}

func debug(a ...any) {
	//rog.Debug(a...)
}

var expandBufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

var nodePool = sync.Pool{
	New: func() any {
		return &Node{}
	},
}
