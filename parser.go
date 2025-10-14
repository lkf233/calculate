package main

import (
    "fmt"
)

// ParseExpression parses an infix expression string into an Expr AST.
// Supports parentheses, +, -, ×, ÷ and numbers: integer, a/b, k’a/b.
func ParseExpression(s string) (*Expr, error) {
    p := &parser{src: []rune(s), i: 0}
    expr, err := p.parseExpr()
    if err != nil {
        return nil, err
    }
    // Eat trailing spaces
    p.skipSpaces()
    if p.i != len(p.src) {
        // allow right side of '=' trimmed by caller; treat extra as error
        return nil, fmt.Errorf("多余字符: %s", string(p.src[p.i:]))
    }
    return expr, nil
}

type parser struct {
    src []rune
    i   int
}

func (p *parser) peek() rune {
    if p.i >= len(p.src) {
        return 0
    }
    return p.src[p.i]
}

func (p *parser) next() rune {
    if p.i >= len(p.src) {
        return 0
    }
    ch := p.src[p.i]
    p.i++
    return ch
}

func (p *parser) skipSpaces() {
    for p.i < len(p.src) {
        c := p.src[p.i]
        if c == ' ' || c == '\t' {
            p.i++
        } else {
            break
        }
    }
}

// Grammar: E -> T {( + | - ) T}*
//          T -> F {( × | ÷ ) F}*
//          F -> number | ( E )
func (p *parser) parseExpr() (*Expr, error) {
    left, err := p.parseTerm()
    if err != nil {
        return nil, err
    }
    for {
        p.skipSpaces()
        if p.i >= len(p.src) {
            break
        }
        op, ok := p.tryAddSub()
        if !ok {
            break
        }
        right, err := p.parseTerm()
        if err != nil {
            return nil, err
        }
        left = NewNode(op, left, right)
    }
    return left, nil
}

func (p *parser) parseTerm() (*Expr, error) {
    left, err := p.parseFactor()
    if err != nil {
        return nil, err
    }
    for {
        p.skipSpaces()
        if p.i >= len(p.src) {
            break
        }
        op, ok := p.tryMulDiv()
        if !ok {
            break
        }
        right, err := p.parseFactor()
        if err != nil {
            return nil, err
        }
        left = NewNode(op, left, right)
    }
    return left, nil
}

func (p *parser) parseFactor() (*Expr, error) {
    p.skipSpaces()
    if p.i >= len(p.src) {
        return nil, fmt.Errorf("缺少因子")
    }
    ch := p.peek()
    if ch == '(' {
        p.next() // consume '('
        e, err := p.parseExpr()
        if err != nil {
            return nil, err
        }
        p.skipSpaces()
        if p.i >= len(p.src) || p.peek() != ')' {
            return nil, fmt.Errorf("缺少右括号")
        }
        p.next() // consume ')'
        return e, nil
    }
    // number
    start := p.i
    for p.i < len(p.src) {
        c := p.src[p.i]
        // Allowed in number: digits, '/', apostrophe both types, sign only at start
        if (c >= '0' && c <= '9') || c == '/' || c == '\'' || c == '’' || (c == '-' && p.i == start) {
            p.i++
            continue
        }
        break
    }
    if start == p.i {
        return nil, fmt.Errorf("缺少数字")
    }
    s := string(p.src[start:p.i])
    v, err := ParseNumber(s)
    if err != nil {
        return nil, err
    }
    return NewLeaf(v), nil
}

func (p *parser) tryAddSub() (Op, bool) {
    p.skipSpaces()
    if p.i >= len(p.src) {
        return OpNone, false
    }
    c := p.peek()
    if c == '+' {
        p.next()
        return OpAdd, true
    }
    if c == '-' {
        p.next()
        return OpSub, true
    }
    return OpNone, false
}

func (p *parser) tryMulDiv() (Op, bool) {
    p.skipSpaces()
    if p.i >= len(p.src) {
        return OpNone, false
    }
    c := p.peek()
    if c == '×' || c == '*' {
        p.next()
        return OpMul, true
    }
    if c == '÷' || c == '/' {
        // Careful: '/' also part of number; but parseFactor consumes numbers fully, so here '/' is operator between factors.
        p.next()
        return OpDiv, true
    }
    return OpNone, false
}