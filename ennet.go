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
	"regexp"
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
	err := Parse(b, &nl)
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

var mulre = regexp.MustCompile(`(\$+)(@(-)?(\d+)?)?`)

func applyMul(s string, mul int) string {
	if mul > 0 {
		templ := s
		s = ""
		for i := 0; i < mul; i++ {
			s += mulre.ReplaceAllStringFunc(templ, func(tgt string) string {
				g := mulre.FindStringSubmatch(tgt)
				//debug(tgt, g)
				minus := false
				base := 1
				pad := len(g[1])
				minus = g[3] == "-"
				if g[4] != "" {
					base, _ = strconv.Atoi(g[4])
				}

				if !minus {
					return fmt.Sprintf("%0"+strconv.Itoa(pad)+"d", base+i)
				} else {
					return fmt.Sprintf("%0"+strconv.Itoa(pad)+"d", base+mul-1-i)
				}
			})
		}
	}
	return s
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
