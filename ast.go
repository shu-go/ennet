package ennet

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type NodeType uint8

const (
	Root NodeType = iota
	WIP

	Element
	Text
	Group
)

var nodeType2String = map[NodeType]string{
	Root:    "ROOT",
	WIP:     "WIP",
	Element: "element",
	Text:    "text",
	Group:   "group",
}

func (t NodeType) String() string {
	if s, found := nodeType2String[t]; found {
		return s
	}
	return "???"
}

type Attribute struct {
	Name  string
	Value string
}

type Node struct {
	Type       NodeType
	Data       string
	Attributes []Attribute

	Mul int

	Parent, FirstChild, LastChild, NextSibling, PrevSibling *Node
}

func (n *Node) GetAttribute(name string) string {
	for _, attr := range n.Attributes {
		if attr.Name == name {
			return attr.Value
		}
	}
	return ""
}

func (n *Node) AppendChild(child *Node) *Node {
	if lc := n.LastChild; lc != nil {
		lc.NextSibling = child
		child.PrevSibling = lc
	} else {
		n.FirstChild = child
	}

	n.LastChild = child
	child.Parent = n

	return child
}

func (n *Node) Dump() string {
	return n.dump(0)
}

func (n *Node) dump(indent int) string {
	var s strings.Builder
	s.WriteString(strings.Repeat(" ", indent*2))
	if n.Type == Element || n.Type == Group {
		s.WriteString(n.Data + ":" + n.Type.String())
	} else if n.Type == Text {
		s.WriteString(`"` + n.Data + `"`)
	}

	if len(n.Attributes) > 0 {
		attrs := make([]Attribute, len(n.Attributes))
		copy(attrs, n.Attributes)
		slices.SortFunc(attrs, func(a, b Attribute) int {
			return strings.Compare(a.Name, b.Name)
		})
		for _, attr := range attrs {
			s.WriteString(" @" + attr.Name + "=" + attr.Value)
		}
	}

	if n.Mul > 1 {
		s.WriteString(" *" + strconv.Itoa(n.Mul))
	}

	child := n.FirstChild
	for child != nil {
		s.WriteString("\n" + child.dump(indent+1))
		child = child.NextSibling
	}
	return s.String()
}

type Builder interface {
	Element(name string) error
	ID(name string) error
	Class(name string) error
	Attribute(name, value string) error
	Text(text string) error
	Mul(count int) error

	OpChild() error
	OpSibling() error
	OpClimbup(count int) error

	GroupBegin() error
	GroupEnd() error
}

type NodeBuilder struct {
	Root *Node
	curr *Node

	pool *sync.Pool
}

func (b *NodeBuilder) NewNode() *Node {
	if b.pool != nil {
		n := b.pool.Get().(*Node)
		// keep existing slice capacity
		attrs := n.Attributes
		*n = Node{}
		n.Attributes = attrs[:0]
		return n
	}
	return &Node{}
}

type NodeBuilderOption func(*NodeBuilder)

func WithPool(pool *sync.Pool) func(*NodeBuilder) {
	return func(b *NodeBuilder) {
		b.pool = pool
	}
}

func NewNodeBuilder(opts ...NodeBuilderOption) NodeBuilder {
	var b NodeBuilder
	for i := range opts {
		opts[i](&b)
	}

	root := b.NewNode()
	root.Type = Root

	b.Root = root
	b.curr = b.NewNode()
	b.curr.Type = WIP
	b.curr = root.AppendChild(b.curr)

	return b
}

func (nb *NodeBuilder) Element(name string) error {
	if nb.curr.Type == WIP {
		nb.curr.Type = Element
		nb.curr.Data = name
	}

	return nil
}

func (nb *NodeBuilder) Attribute(name, value string) error {
	if nb.curr.Type == Text {
		return errors.New("attribute of Text")
	}
	if nb.curr.FirstChild != nil && nb.curr.FirstChild.Type == Text {
		return errors.New("attribute must appear before Text")
	}

	if nb.curr.Type == WIP {
		nb.curr.Type = Element
	}

	for i := range nb.curr.Attributes {
		if nb.curr.Attributes[i].Name == name {
			nb.curr.Attributes[i].Value += " " + value
			return nil
		}
	}
	nb.curr.Attributes = append(nb.curr.Attributes, Attribute{Name: name, Value: value})

	return nil
}

func (nb *NodeBuilder) ID(name string) error {
	return nb.Attribute("id", name)
}

func (nb *NodeBuilder) Class(name string) error {
	return nb.Attribute("class", name)
}

func (nb *NodeBuilder) Mul(count int) error {
	nb.curr.Mul = count

	return nil
}

func (nb *NodeBuilder) Text(text string) error {
	if nb.curr.Type == Text {
		nb.curr.Data += text
		return nil
	}

	if nb.curr.Type == WIP {
		nb.curr.Type = Text
		nb.curr.Data += text
		return nil
	}

	node := nb.NewNode()
	node.Type = Text
	node.Data = text
	nb.curr.AppendChild(node)

	return nil
}

func (nb *NodeBuilder) GroupBegin() error {
	if nb.curr.Type == WIP {
		nb.curr.Type = Group
		node := nb.NewNode()
		node.Type = WIP
		nb.curr.AppendChild(node)
		nb.curr = node
	} else {
		node := nb.NewNode()
		node.Type = Group
		nb.curr.AppendChild(node)
		nb.curr = node
	}

	return nil
}

func (nb *NodeBuilder) GroupEnd() error {
	if nb.curr.Parent != nil && nb.curr.Parent.Type != Root {
		nb.curr = nb.curr.Parent
	}

	for {
		if nb.curr.Parent == nil ||
			nb.curr.Parent.Type == Root ||
			nb.curr.Type == Group {
			break
		}
		nb.curr = nb.curr.Parent
	}

	return nil
}

func (nb *NodeBuilder) OpChild() error {
	node := nb.NewNode()
	node.Type = WIP
	nb.curr.AppendChild(node)
	nb.curr = node

	return nil
}

func (nb *NodeBuilder) OpSibling() error {
	node := nb.NewNode()
	node.Type = WIP
	nb.curr.Parent.AppendChild(node)
	nb.curr = node

	return nil
}

func (nb *NodeBuilder) OpClimbup(count int) error {
	for range count {
		if nb.curr.Parent == nil || nb.curr.Parent.Type == Root {
			break
		}
		nb.curr = nb.curr.Parent
	}

	node := nb.NewNode()
	node.Type = WIP
	nb.curr.Parent.AppendChild(node)
	nb.curr = node

	return nil
}
