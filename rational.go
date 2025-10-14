package main

import (
    "fmt"
)

type Rational struct {
    Num int64
    Den int64
}

func NewRational(num, den int64) Rational {
    if den == 0 {
        panic("denominator cannot be zero")
    }
    // Normalize sign and reduce
    if den < 0 {
        num = -num
        den = -den
    }
    g := gcd(abs64(num), den)
    return Rational{Num: num / g, Den: den / g}
}

func (r Rational) Equals(o Rational) bool {
    // reduced, so simple cross multiply
    return r.Num*o.Den == o.Num*r.Den
}

func (r Rational) Add(o Rational) Rational {
    return NewRational(r.Num*o.Den+o.Num*r.Den, r.Den*o.Den)
}

func (r Rational) Sub(o Rational) Rational {
    return NewRational(r.Num*o.Den-o.Num*r.Den, r.Den*o.Den)
}

func (r Rational) Mul(o Rational) Rational {
    return NewRational(r.Num*o.Num, r.Den*o.Den)
}

func (r Rational) Div(o Rational) Rational {
    if o.Num == 0 {
        panic("division by zero")
    }
    return NewRational(r.Num*o.Den, r.Den*o.Num)
}

func (r Rational) Less(o Rational) bool {
    return r.Num*o.Den < o.Num*r.Den
}

func (r Rational) LessEq(o Rational) bool {
    return r.Num*o.Den <= o.Num*r.Den
}

func (r Rational) IsZero() bool { return r.Num == 0 }

func (r Rational) String() string {
    // format as specified: integer, proper fraction a/b, or mixed k’ a/b
    if r.Den == 1 {
        return fmt.Sprintf("%d", r.Num)
    }
    // handle negative? generator avoids; parser may produce; ensure format still works
    if r.Num < 0 {
        // Show negative mixed properly
        abs := NewRational(-r.Num, r.Den)
        s := abs.String()
        return "-" + s
    }
    n := r.Num
    d := r.Den
    if n < d {
        return fmt.Sprintf("%d/%d", n, d)
    }
    k := n / d
    rem := n % d
    if rem == 0 {
        return fmt.Sprintf("%d", k)
    }
    return fmt.Sprintf("%d’%d/%d", k, rem, d)
}

// ParseNumber parses integer, a/b, or k’a/b (supports ascii ' or unicode ’)
func ParseNumber(s string) (Rational, error) {
    // trim spaces
    s = trimSpaces(s)
    // Try mixed
    var quoteIdx = -1
    for i, ch := range s {
        if ch == '\'' || ch == '’' { // ascii apostrophe or unicode right single quote
            quoteIdx = i
            break
        }
    }
    if quoteIdx != -1 {
        // k’a/b
        kPart := s[:quoteIdx]
        rest := s[quoteIdx+1:]
        num, den, ok := splitFraction(rest)
        if !ok {
            return Rational{}, fmt.Errorf("非法分数: %s", s)
        }
        k, err := parseInt64(kPart)
        if err != nil {
            return Rational{}, err
        }
        if den == 0 {
            return Rational{}, fmt.Errorf("分母不能为0: %s", s)
        }
        return NewRational(int64(k)*int64(den)+int64(num), int64(den)), nil
    }
    // Try pure fraction "a/b"
    num, den, ok := splitFraction(s)
    if ok {
        if den == 0 {
            return Rational{}, fmt.Errorf("分母不能为0: %s", s)
        }
        return NewRational(int64(num), int64(den)), nil
    }
    // Integer
    v, err := parseInt64(s)
    if err != nil {
        return Rational{}, err
    }
    return NewRational(v, 1), nil
}

func splitFraction(s string) (int64, int64, bool) {
    idx := -1
    for i, ch := range s {
        if ch == '/' {
            idx = i
            break
        }
    }
    if idx == -1 {
        return 0, 0, false
    }
    a := trimSpaces(s[:idx])
    b := trimSpaces(s[idx+1:])
    ai, err := parseInt64(a)
    if err != nil {
        return 0, 0, false
    }
    bi, err := parseInt64(b)
    if err != nil {
        return 0, 0, false
    }
    return ai, bi, true
}

func parseInt64(s string) (int64, error) {
    var sign int64 = 1
    if len(s) == 0 {
        return 0, fmt.Errorf("空数字")
    }
    if s[0] == '+' {
        s = s[1:]
    } else if s[0] == '-' {
        sign = -1
        s = s[1:]
    }
    var v int64 = 0
    for _, ch := range s {
        if ch < '0' || ch > '9' {
            return 0, fmt.Errorf("非法数字: %s", s)
        }
        v = v*10 + int64(ch-'0')
    }
    return sign * v, nil
}

func gcd(a, b int64) int64 {
    for b != 0 {
        a, b = b, a%b
    }
    if a < 0 {
        return -a
    }
    return a
}

func abs64(x int64) int64 { if x < 0 { return -x }; return x }

func trimSpaces(s string) string {
    // simple trim both sides spaces and tabs
    start := 0
    end := len(s)
    for start < end {
        c := s[start]
        if c == ' ' || c == '\t' { start++ } else { break }
    }
    for end > start {
        c := s[end-1]
        if c == ' ' || c == '\t' { end-- } else { break }
    }
    return s[start:end]
}