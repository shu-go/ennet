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
	"slices"
	"strconv"
	"strings"
	"sync"
)

// Expand expands Emmet abbreviation in string s.
func Expand(s string) (string, error) {
	b := expandBufPool.Get().(*bytes.Buffer)
	b.Reset()
	b.WriteString(s)

	nodeBuilder := NewNodeBuilder(&nodePool)
	err := Parse(b.Bytes(), &nodeBuilder)
	if err != nil {
		expandBufPool.Put(b)
		return "", err
	}

	resBuf := expandBufPool.Get().(*bytes.Buffer)
	resBuf.Reset()
	expand(nodeBuilder.Root, resBuf)
	result := resBuf.String()

	gcNodes(&nodeBuilder, &nodePool, nodeBuilder.Root)

	expandBufPool.Put(resBuf)
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

func expand(n *Node, w *bytes.Buffer) {
	if n.Mul <= 0 {
		expandNode(n, w)
		return
	}

	buf := expandBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	expandNode(n, buf)
	writeMul(w, buf.Bytes(), n.Mul)
	expandBufPool.Put(buf)
}

func expandNode(n *Node, w *bytes.Buffer) {
	switch n.Type {
	case Text:
		w.WriteString(n.Data)
	case Root, Group:
		curr := n.FirstChild
		for curr != nil {
			expand(curr, w)
			curr = curr.NextSibling
		}
	case Element:
		w.WriteString("<")
		w.WriteString(n.Data)
		if len(n.Attributes) > 0 {
			slices.SortFunc(n.Attributes, func(a, b Attribute) int {
				return strings.Compare(a.Name, b.Name)
			})
			for _, attr := range n.Attributes {
				w.WriteString(" ")
				w.WriteString(attr.Name)
				w.WriteString(`="`)
				// manual escape
				val := attr.Value
				last := 0
				for i := 0; i < len(val); i++ {
					if val[i] == '"' {
						w.WriteString(val[last:i])
						w.WriteString(`\"`)
						last = i + 1
					}
				}
				w.WriteString(val[last:])
				w.WriteString(`"`)
			}
		}

		if n.FirstChild == nil {
			w.WriteString(" />")
		} else {
			w.WriteString(">")
			curr := n.FirstChild
			for curr != nil {
				expand(curr, w)
				curr = curr.NextSibling
			}
			w.WriteString("</")
			w.WriteString(n.Data)
			w.WriteString(">")
		}
	}
}

type segment struct {
	literal []byte
	isPH    bool
	pad     int
	base    int
	minus   bool
}

func writeMul(w *bytes.Buffer, templ []byte, mul int) {
	pSegs := segmentSlicePool.Get().(*[]segment)
	segments := parseTemplateInto(templ, (*pSegs)[:0])

	var numBuf [32]byte
	for i := range mul {
		for _, seg := range segments {
			if !seg.isPH {
				w.Write(seg.literal)
			} else {
				val := seg.base + i
				if seg.minus {
					val = seg.base + mul - 1 - i
				}
				b := strconv.AppendInt(numBuf[:0], int64(val), 10)
				for j := 0; j < seg.pad-len(b); j++ {
					w.WriteString("0")
				}
				w.Write(b)
			}
		}
	}

	*pSegs = segments
	segmentSlicePool.Put(pSegs)
}

func parseTemplateInto(templ []byte, segments []segment) []segment {
	last := 0
	for i := 0; i < len(templ); i++ {
		if templ[i] == '$' {
			if i > last {
				segments = append(segments, segment{literal: templ[last:i]})
			}
			pad := 1
			i++
			for i < len(templ) && templ[i] == '$' {
				pad++
				i++
			}
			minus := false
			base := 1
			if i < len(templ) && templ[i] == '@' {
				i++
				if i < len(templ) && templ[i] == '-' {
					minus = true
					i++
				}
				b := 0
				foundBase := false
				for i < len(templ) && templ[i] >= '0' && templ[i] <= '9' {
					b = b*10 + int(templ[i]-'0')
					i++
					foundBase = true
				}
				if foundBase && b > 0 {
					base = b
				}
			}
			segments = append(segments, segment{isPH: true, pad: pad, base: base, minus: minus})
			last = i
			i-- // back up for the outer loop's i++
		}
	}
	if last < len(templ) {
		segments = append(segments, segment{literal: templ[last:]})
	}
	return segments
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

var segmentSlicePool = sync.Pool{
	New: func() any {
		s := make([]segment, 0, 16)
		return &s
	},
}
