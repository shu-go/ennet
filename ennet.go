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
	"io"
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

	nodeBuilder := NewNodeBuilder(WithPool(&nodePool))
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

func expand(n *Node, w io.Writer) {
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

func expandNode(n *Node, w io.Writer) {
	switch n.Type {
	case Text:
		io.WriteString(w, n.Data)
	case Root, Group:
		curr := n.FirstChild
		for curr != nil {
			expand(curr, w)
			curr = curr.NextSibling
		}
	case Element:
		io.WriteString(w, "<")
		io.WriteString(w, n.Data)
		if len(n.Attributes) > 0 {
			slices.SortFunc(n.Attributes, func(a, b Attribute) int {
				return strings.Compare(a.Name, b.Name)
			})
			for _, attr := range n.Attributes {
				io.WriteString(w, " ")
				io.WriteString(w, attr.Name)
				io.WriteString(w, `="`)
				// manual escape
				val := attr.Value
				last := 0
				for i := 0; i < len(val); i++ {
					if val[i] == '"' {
						io.WriteString(w, val[last:i])
						io.WriteString(w, `\"`)
						last = i + 1
					}
				}
				io.WriteString(w, val[last:])
				io.WriteString(w, `"`)
			}
		}

		if n.FirstChild == nil {
			io.WriteString(w, " />")
		} else {
			io.WriteString(w, ">")
			curr := n.FirstChild
			for curr != nil {
				expand(curr, w)
				curr = curr.NextSibling
			}
			io.WriteString(w, "</")
			io.WriteString(w, n.Data)
			io.WriteString(w, ">")
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

func writeMul(w io.Writer, templ []byte, mul int) {
	segments := parseTemplate(templ)
	var numBuf [32]byte
	for i := 0; i < mul; i++ {
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
					io.WriteString(w, "0")
				}
				w.Write(b)
			}
		}
	}
}

func parseTemplate(templ []byte) []segment {
	var segments []segment
	last := 0
	for i := 0; i < len(templ); i++ {
		if templ[i] == '$' {
			if i > last {
				segments = append(segments, segment{literal: templ[last:i]})
			}
			start := i
			_ = start
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
