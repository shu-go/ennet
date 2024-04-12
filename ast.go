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
	for i := 0; i < len(opts); i++ {
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

func (nl *NodeBuilder) Element(name string) error {
	if nl.curr.Type == WIP {
		nl.curr.Type = Element
		nl.curr.Data = name
	}

	return nil
}

func (nl *NodeBuilder) Attribute(name, value string) error {
	if nl.curr.Type == Text {
		return errors.New("attribute of Text")
	}
	if nl.curr.FirstChild != nil && nl.curr.FirstChild.Type == Text {
		return errors.New("attribute must appear before Text")
	}

	if nl.curr.Type == WIP {
		nl.curr.Type = Element
	}

	if nl.curr.Attribute == nil {
		nl.curr.Attribute = make(map[string]string)
	}

	if v, found := nl.curr.Attribute[name]; found {
		nl.curr.Attribute[name] = v + " " + value
	} else {
		nl.curr.Attribute[name] = value
	}

	return nil
}

func (nl *NodeBuilder) ID(name string) error {
	return nl.Attribute("id", name)
}

func (nl *NodeBuilder) Class(name string) error {
	return nl.Attribute("class", name)
}

func (nl *NodeBuilder) Mul(count int) error {
	nl.curr.Mul = count

	return nil
}

func (nl *NodeBuilder) Text(text string) error {
	if nl.curr.Type == Text {
		nl.curr.Data += text
		return nil
	}

	if nl.curr.Type == WIP {
		nl.curr.Type = Text
		nl.curr.Data += text
		return nil
	}

	node := nl.NewNode()
	node.Type = Text
	node.Data = text
	nl.curr.AppendChild(node)

	return nil
}

func (nl *NodeBuilder) GroupBegin() error {
	if nl.curr.Type == WIP {
		nl.curr.Type = Group
		node := nl.NewNode()
		node.Type = WIP
		nl.curr.AppendChild(node)
		nl.curr = node
	} else {
		node := nl.NewNode()
		node.Type = Group
		nl.curr.AppendChild(node)
		nl.curr = node
	}

	return nil
}

func (nl *NodeBuilder) GroupEnd() error {
	if nl.curr.Parent != nil && nl.curr.Parent.Type != Root {
		nl.curr = nl.curr.Parent
	}

	for {
		if nl.curr.Parent == nil ||
			nl.curr.Parent.Type == Root ||
			nl.curr.Type == Group {
			break
		}
		nl.curr = nl.curr.Parent
	}

	return nil
}

func (nl *NodeBuilder) OpChild() error {
	node := nl.NewNode()
	node.Type = WIP
	nl.curr.AppendChild(node)
	nl.curr = node

	return nil
}

func (nl *NodeBuilder) OpSibling() error {
	node := nl.NewNode()
	node.Type = WIP
	nl.curr.Parent.AppendChild(node)
	nl.curr = node

	return nil
}

func (nl *NodeBuilder) OpClimbup(count int) error {
	for i := 0; i < count; i++ {
		if nl.curr.Parent == nil || nl.curr.Parent.Type == Root {
			break
		}
		nl.curr = nl.curr.Parent
	}

	node := nl.NewNode()
	node.Type = WIP
	nl.curr.Parent.AppendChild(node)
	nl.curr = node

	return nil
}
