package syntax

import (
	"fmt"
	"hash/fnv"
	"os"

	"github.com/andrewpillar/req/token"
)

type parser struct {
	*scanner

	errc int
}

func ParseFile(fname string, errh func(token.Pos, string)) ([]Node, error) {
	f, err := os.Open(fname)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	p := parser{
		scanner: newScanner(newSource(fname, f, errh)),
	}

	if p.errc > 0 {
		return nil, fmt.Errorf("parser encountered %d error(s)", p.errc)
	}
	return p.parse(), nil
}

func (p *parser) got(tok token.Token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) errAt(pos token.Pos, msg string) {
	p.errc++
	p.scanner.source.errh(pos, msg)
}

func (p *parser) err(msg string) {
	p.errAt(p.pos, msg)
}

func (p *parser) expected(tok token.Token) {
	p.err("expected " + tok.String())
}

func (p *parser) unexpected(tok token.Token) {
	p.err("unexpected " + tok.String())
}

func (p *parser) want(tok token.Token) {
	if !p.got(tok) {
		p.expected(tok)
	}
}

func (p *parser) node() node {
	return node{
		pos: p.pos,
	}
}

func (p *parser) ident() *Ident {
	if p.tok != token.Name {
		return nil
	}

	n := &Ident{
		node: p.node(),
		Name: p.lit,
	}
	p.next()
	return n
}

func (p *parser) literal() *Lit {
	if p.tok != token.Literal {
		return nil
	}

	n := &Lit{
		node:  p.node(),
		Type:  p.typ,
		Value: p.lit,
	}
	p.next()
	return n
}

func (p *parser) ref() *Ref {
	if p.tok != token.Ref {
		return nil
	}

	p.got(token.Ref)

	ref := &Ref{
		Left: p.ident(),
	}

loop:
	for {
		pos := p.pos

		switch p.tok {
		case token.Dot:
			p.next()

			if p.tok != token.Name {
				p.expected(token.Name)
				p.next()
				return nil
			}

			left := ref.Left

			ref.Left = &DotExpr{
				node:  node{pos: pos},
				Left:  left,
				Right: p.ident(),
			}
		case token.Lbrack:
			p.next()

			if p.tok == token.Rbrack {
				p.err("expected string, int, or variable")
				p.next()
				break
			}

			left := ref.Left
			ind := &IndExpr{
				node: node{pos: pos},
				Left: left,
			}

			switch p.tok {
			case token.Literal:
				ind.Right = p.literal()
			case token.Ref:
				ind.Right = p.ref()
			default:
				p.unexpected(p.tok)
				p.next()
			}
			ref.Left = ind
		default:
			break loop
		}
	}
	return ref
}

func (p *parser) list(sep, end token.Token, parse func()) {
	for p.tok != token.EOF && p.tok != end {
		parse()

		if !p.got(sep) && p.tok != end {
			p.err("expected " + sep.String() + " or " + end.String())
			p.next()
		}
	}
	p.want(end)
}

func (p *parser) obj() *Object {
	p.want(token.Lbrace)

	n := &Object{
		node: p.node(),
	}

	p.list(token.Comma, token.Rbrace, func() {
		if p.tok != token.Name {
			p.expected(token.Name)
			p.next()
			return
		}

		key := p.ident()

		p.want(token.Colon)

		val := p.operand()

		n.Pairs = append(n.Pairs, &KeyExpr{
			node:  p.node(),
			Key:   key,
			Value: val,
		})
	})
	return n
}

func (p *parser) arr() *Array {
	p.want(token.Lbrack)

	n := &Array{
		node: p.node(),
	}

	p.list(token.Comma, token.Rbrack, func() {
		if p.tok != token.Literal {
			p.expected(token.Literal)
			p.next()
			return
		}
		n.Items = append(n.Items, p.literal())
	})
	return n
}

func (p *parser) blockstmt() *BlockStmt {
	p.want(token.Lbrace)

	n := &BlockStmt{
		node: p.node(),
	}

	for p.tok != token.Rbrace && p.tok != token.EOF {
		n.Nodes = append(n.Nodes, p.stmt())
	}

	p.want(token.Rbrace)
	return n
}

func (p *parser) yield() *YieldStmt {
	if !p.got(token.Yield) {
		return nil
	}

	return &YieldStmt{
		node:  p.node(),
		Value: p.operand(),
	}
}

func (p *parser) casestmt(jmptab map[uint32]Node) {
	var lit string

	if p.tok == token.Name {
		if p.lit != "_" {
			p.unexpected(p.tok)
			return
		}

		lit = p.lit
		p.next()
		goto right
	}

	if p.tok != token.Literal {
		p.unexpected(p.tok)
		return
	}

	lit = p.lit
	p.next()

right:
	p.want(token.Arrow)

	h := fnv.New32()
	h.Write([]byte(lit))

	sum := h.Sum32()

	switch p.tok {
	case token.Lbrace:
		jmptab[sum] = p.blockstmt()
	case token.Yield:
		jmptab[sum] = p.yield()
	default:
		p.unexpected(p.tok)
	}
}

func (p *parser) matchstmt() *MatchStmt {
	if p.tok != token.Match {
		return nil
	}

	n := &MatchStmt{
		node:   p.node(),
		Jmptab: make(map[uint32]Node),
	}

	p.next()

	switch p.tok {
	case token.Literal:
		n.Cond = p.literal()
	case token.Ref:
		n.Cond = p.ref()
	default:
		p.unexpected(p.tok)
		p.next()
	}

	p.want(token.Lbrace)

	for p.tok != token.Rbrace {
		p.casestmt(n.Jmptab)

		if p.tok != token.Comma && p.tok != token.Rbrace {
			p.err("expected comma or }")
			p.next()
			continue
		}
		p.got(token.Comma)
	}

	p.got(token.Rbrace)
	return n
}

func (p *parser) chain(cmd *CommandStmt) *ChainStmt {
	n := &ChainStmt{
		Commands: []*CommandStmt{cmd},
	}

	for p.tok != token.Semi && p.tok != token.EOF {
		if p.tok != token.Name {
			p.expected(token.Name)
			p.next()
			continue
		}

		ident := p.ident()
		n.Commands = append(n.Commands, p.command(ident.Name))

		if !p.got(token.Arrow) && p.tok != token.Semi && p.tok != token.EOF {
			p.err("expected " + token.Arrow.String() + " or " + token.Semi.String())
			p.next()
		}
	}
	return n
}

func (p *parser) operand() Node {
	var n Node

	switch p.tok {
	case token.Literal:
		n = p.literal()
	case token.Ref:
		n = p.ref()
	case token.Lbrace:
		n = p.obj()
	case token.Lbrack:
		n = p.arr()
	default:
		p.unexpected(p.tok)
		p.next()
	}
	return n
}

func (p *parser) expr() Node {
	switch p.tok {
	case token.Name:
		ident := p.ident()

		n := p.command(ident.Name)

		if p.got(token.Arrow) {
			return p.chain(n)
		}
		return n
	case token.Match:
		return p.matchstmt()
	default:
		return p.operand()
	}
}

func (p *parser) command(name string) *CommandStmt {
	n := &CommandStmt{
		node: p.node(),
		Name: name,
	}

	for p.tok != token.Arrow && p.tok != token.Semi && p.tok != token.EOF {
		n.Args = append(n.Args, p.expr())
	}
	return n
}

func (p *parser) vardecl(ident *Ident) *VarDecl {
	n := &VarDecl{
		node:  p.node(),
		Ident: ident,
	}

	if !p.got(token.Assign) {
		return nil
	}

	n.Value = p.expr()

	return n
}

func (p *parser) stmt() Node {
	var n Node

	switch p.tok {
	case token.Name:
		ident := p.ident()

		if p.tok == token.Assign {
			n = p.vardecl(ident)
			break
		}

		cmd := p.command(ident.Name)
		n = cmd

		if p.got(token.Arrow) {
			n = p.chain(cmd)
		}
	case token.Match:
		n = p.matchstmt()
	case token.Yield:
		n = p.yield()
	default:
		p.unexpected(p.tok)
		p.next()
	}

	if p.tok != token.EOF {
		if !p.got(token.Semi) {
			p.expected(token.Semi)
		}
	}
	return n
}

func (p *parser) parse() []Node {
	nn := make([]Node, 0)

	for p.tok != token.EOF {
		nn = append(nn, p.stmt())
	}
	return nn
}
