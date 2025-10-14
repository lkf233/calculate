package main

import (
    "errors"
    "math/rand"
)

// GenerateProblems generates n unique problems under constraints with value range r
func GenerateProblems(n, r int) ([]string, []string, error) {
    // 预估容量，减少哈希表扩容与重新哈希的开销
    seen := make(map[string]struct{}, n*2)
    exercises := make([]string, 0, n)
    answers := make([]string, 0, n)

    // attempts cap to avoid infinite loops; allow generous tries
    maxAttempts := n * 200
    attempts := 0
    for len(exercises) < n && attempts < maxAttempts {
        attempts++
        maxOps := 1 + rand.Intn(3) // 1..3 operators
        expr := genExpr(r, maxOps)
        // enforce constraints by construction, but double-check critical ones
        if !validateExprConstraints(expr) {
            continue
        }
        key := expr.CanonicalKey()
        if _, ok := seen[key]; ok {
            continue
        }
        seen[key] = struct{}{}

        exStr := expr.ExprString(true) + " = " // spaces around '=' per spec
        ans := expr.Eval().String()
        exercises = append(exercises, exStr)
        answers = append(answers, ans)
    }
    if len(exercises) < n {
        return nil, nil, errors.New("未能在合理尝试次数内生成足够的不重复题目，请增大范围或减少数量")
    }
    return exercises, answers, nil
}

func validateExprConstraints(e *Expr) bool {
    // Recursively ensure: for every subtraction child: left >= right; for division: result < 1 (proper fraction) and denom non-zero
    if e.Op == OpNone {
        return true
    }
    if !validateExprConstraints(e.Left) || !validateExprConstraints(e.Right) {
        return false
    }
    lv := e.Left.Eval()
    rv := e.Right.Eval()
    switch e.Op {
    case OpSub:
        // require left >= right
        return rv.LessEq(lv)
    case OpDiv:
        if rv.IsZero() {
            return false
        }
        // result should be proper fraction: lv < rv
        return lv.Less(rv)
    default:
        return true
    }
}

// genExpr recursively builds an expression with exactly maxOps operators
func genExpr(r, maxOps int) *Expr {
    if maxOps == 0 {
        v := randomNumber(r)
        return NewLeaf(v)
    }
    // Decide operator
    op := randomOp()
    // Split operators count between left and right
    leftOps := 0
    if maxOps > 1 {
        leftOps = rand.Intn(maxOps) // 0..maxOps-1
    }
    rightOps := maxOps - 1 - leftOps
    left := genExpr(r, leftOps)
    right := genExpr(r, rightOps)

    // Fix constraints by swapping/regenerating when necessary
    switch op {
    case OpSub:
        // ensure left >= right
        lv := left.Eval()
        rv := right.Eval()
        if lv.Less(rv) {
            // swap to satisfy
            left, right = right, left
        }
    case OpDiv:
        // ensure right != 0 and left < right
        lv := left.Eval()
        rv := right.Eval()
        // regenerate right until non-zero (limit tries)
        tries := 0
        for rv.IsZero() && tries < 10 {
            right = genExpr(r, rightOps)
            rv = right.Eval()
            tries++
        }
        if lv.LessEq(rv) {
            // if not left < right, try swap; if still not, regenerate one side
            if !lv.Less(rv) {
                left, right = right, left
                lv = left.Eval()
                rv = right.Eval()
            }
        }
        if rv.IsZero() || !lv.Less(rv) {
            // fallback: adjust by regenerating smaller left and larger right
            left = genSmallExpr(r, leftOps)
            right = genLargeExpr(r, rightOps, left.Eval())
        }
    }
    return NewNode(op, left, right)
}

func randomOp() Op {
    switch rand.Intn(4) {
    case 0:
        return OpAdd
    case 1:
        return OpSub
    case 2:
        return OpMul
    default:
        return OpDiv
    }
}

func randomNumber(r int) Rational {
    // choose type: integer, proper fraction, mixed
    t := rand.Intn(100)
    if t < 50 {
        // integer in [0, r-1]
        return NewRational(int64(rand.Intn(max(1, r))), 1)
    } else if t < 80 {
        // proper fraction: 1/den .. (den-1)/den, den in [2, r-1]
        den := 2 + rand.Intn(max(1, r-2)) // if r<=2, fallback to 2
        num := 1 + rand.Intn(den-1)
        // clamp by r as well
        if num >= r {
            num = max(1, r-1)
            if num >= den { // ensure proper
                num = den - 1
            }
        }
        return NewRational(int64(num), int64(den))
    } else {
        // mixed: k’a/b with k in [1, r-1], den in [2, r-1], a in [1, den-1]
        if r <= 2 {
            // fallback to integer
            return NewRational(int64(rand.Intn(max(1, r))), 1)
        }
        k := 1 + rand.Intn(r-1)
        den := 2 + rand.Intn(r-2)
        num := 1 + rand.Intn(den-1)
        return NewRational(int64(k*den+num), int64(den))
    }
}

func genSmallExpr(r, ops int) *Expr {
    // bias towards small value
    if ops == 0 {
        // small integer or small proper fraction
        if r > 2 && rand.Intn(2) == 0 {
            den := 2 + rand.Intn(r-2)
            num := 1 + rand.Intn(min(2, den-1))
            return NewLeaf(NewRational(int64(num), int64(den)))
        }
        return NewLeaf(NewRational(int64(rand.Intn(min(2, max(1, r)))), 1))
    }
    // small combination: prefer subtraction producing small or division with proper fractional result
    left := genSmallExpr(r, ops-1)
    right := NewLeaf(NewRational(1, 1))
    return NewNode(OpDiv, left, right) // yields small
}

func genLargeExpr(r, ops int, greaterThan Rational) *Expr {
    // create an expr that likely evaluates larger than provided greaterThan
    if ops == 0 {
        // make integer near upper range or mixed
        if r >= 3 && rand.Intn(2) == 0 {
            k := max(1, r-1)
            den := 2 + rand.Intn(max(1, r-2))
            num := max(1, den-1)
            return NewLeaf(NewRational(int64(k*den+num), int64(den)))
        }
        return NewLeaf(NewRational(int64(max(1, r-1)), 1))
    }
    // prefer addition or multiplication to increase value
    left := genLargeExpr(r, ops-1, greaterThan)
    right := NewLeaf(NewRational(1, 1))
    return NewNode(OpAdd, left, right)
}

func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }