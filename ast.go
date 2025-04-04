package ennet

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
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

type Node struct {
	Type      NodeType
	Data      string
	Attribute map[string]string

	Mul int

	Parent, FirstChild, LastChild, NextSibling, PrevSibling *Node
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
	s := strings.Repeat(" ", indent*2)
	if n.Type == Element || n.Type == Group {
		s += n.Data + ":" + n.Type.String()
	} else if n.Type == Text {
		s += `"` + n.Data + `"`
	}

	if len(n.Attribute) > 0 {
		keys := maps.Keys(n.Attribute)
		slices.Sort(keys)
		for _, k := range keys {
			s += " @" + k + "=" + n.Attribute[k]
		}
	}

	if n.Mul > 1 {
		s += " *" + strconv.Itoa(n.Mul)
	}

	child := n.FirstChild
	for child != nil {
		s += "\n" + child.dump(indent+1)
		child = child.NextSibling
	}
	return s
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
		*n = Node{}
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

	if nb.curr.Attribute == nil {
		nb.curr.Attribute = make(map[string]string)
	}

	if v, found := nb.curr.Attribute[name]; found {
		nb.curr.Attribute[name] = v + " " + value
	} else {
		nb.curr.Attribute[name] = value
	}

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
