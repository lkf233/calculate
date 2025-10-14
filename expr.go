package main

import (
    "fmt"
    "strings"
)

type Op int

const (
    OpNone Op = iota
    OpAdd
    OpSub
    OpMul
    OpDiv
)

func opSymbol(op Op) string {
    switch op {
    case OpAdd:
        return "+"
    case OpSub:
        return "-"
    case OpMul:
        return "×"
    case OpDiv:
        return "÷"
    default:
        return ""
    }
}

type Expr struct {
    // Leaf when OpNone
    Op        Op
    Left      *Expr
    Right     *Expr
    Value     Rational // for leaf only
    Operators int      // number of ops in subtree
}

func NewLeaf(v Rational) *Expr { return &Expr{Op: OpNone, Value: v, Operators: 0} }

func NewNode(op Op, l, r *Expr) *Expr {
    return &Expr{Op: op, Left: l, Right: r, Operators: l.Operators + r.Operators + 1}
}

// ExprString prints expression with parentheses around non-root binary nodes
func (e *Expr) ExprString(isRoot bool) string {
    if e.Op == OpNone {
        return e.Value.String()
    }
    ls := e.Left.ExprString(false)
    rs := e.Right.ExprString(false)
    s := fmt.Sprintf("%s %s %s", ls, opSymbol(e.Op), rs)
    if isRoot {
        return s
    }
    return fmt.Sprintf("(%s)", s)
}

// CanonicalKey returns a canonical form string to detect duplicates, where
// + and × children are sorted lexicographically (commutative), but not flattened (non-associative key).
// Leaves are represented as irreducible fraction "num/den".
func (e *Expr) CanonicalKey() string {
    if e.Op == OpNone {
        // canonical leaf
        return fmt.Sprintf("%d/%d", e.Value.Num, e.Value.Den)
    }
    lkey := e.Left.CanonicalKey()
    rkey := e.Right.CanonicalKey()
    if e.Op == OpAdd || e.Op == OpMul {
        // sort pair
        if lkey > rkey {
            lkey, rkey = rkey, lkey
        }
    }
    // Always parenthesize nodes to keep shape in key
    return fmt.Sprintf("(%s %s %s)", lkey, opSymbol(e.Op), rkey)
}

// Eval evaluates the expression using Rational ops
func (e *Expr) Eval() Rational {
    if e.Op == OpNone {
        return e.Value
    }
    lv := e.Left.Eval()
    rv := e.Right.Eval()
    switch e.Op {
    case OpAdd:
        return lv.Add(rv)
    case OpSub:
        return lv.Sub(rv)
    case OpMul:
        return lv.Mul(rv)
    case OpDiv:
        return lv.Div(rv)
    default:
        panic("unknown op")
    }
}

// ParseExpression -> RPN (we will use Expr stack approach in parser.go)
// Provided here: token to op mapping helper
func opFromToken(t string) (Op, bool) {
    switch strings.TrimSpace(t) {
    case "+":
        return OpAdd, true
    case "-":
        return OpSub, true
    case "×":
        return OpMul, true
    case "*":
        return OpMul, true
    case "÷":
        return OpDiv, true
    case "/":
        return OpDiv, true
    }
    return OpNone, false
}