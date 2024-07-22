package mir

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/util"
)

// ApplyConstantPropagation simply collapses constant expressions down to single
// values.  For example, "(+ 1 2)" would be collapsed down to "3".
func applyConstantPropagation(e Expr) Expr {
	if p, ok := e.(*Add); ok {
		return applyConstantPropagationAdd(p.Args)
	} else if _, ok := e.(*Constant); ok {
		return e
	} else if _, ok := e.(*ColumnAccess); ok {
		return e
	} else if p, ok := e.(*Mul); ok {
		return applyConstantPropagationMul(p.Args)
	} else if p, ok := e.(*Exp); ok {
		return applyConstantPropagationExp(p.Arg, p.Pow)
	} else if p, ok := e.(*Normalise); ok {
		return applyConstantPropagationNorm(p.Arg)
	} else if p, ok := e.(*Sub); ok {
		return applyConstantPropagationSub(p.Args)
	}
	// Should be unreachable
	panic(fmt.Sprintf("unknown expression: %s", e.String()))
}

func applyConstantPropagationAdd(es []Expr) Expr {
	var zero = fr.NewElement(0)
	sum := &zero
	rs := make([]Expr, len(es))
	//
	for i, e := range es {
		rs[i] = applyConstantPropagation(e)
		// Check for constant
		c, ok := rs[i].(*Constant)
		// Try to continue sum
		if ok && sum != nil {
			sum.Add(sum, c.Value)
		} else {
			sum = nil
		}
	}
	//
	if sum != nil {
		// Propagate constant
		return &Constant{sum}
	}
	// Done
	return &Add{rs}
}

func applyConstantPropagationSub(es []Expr) Expr {
	var sum *fr.Element = nil

	rs := make([]Expr, len(es))
	//
	for i, e := range es {
		rs[i] = applyConstantPropagation(e)
		// Check for constant
		c, ok := rs[i].(*Constant)
		// Try to continue sum
		if ok && i == 0 {
			var val fr.Element
			// Clone value
			val.Set(c.Value)
			sum = &val
		} else if ok && sum != nil {
			sum.Sub(sum, c.Value)
		} else {
			sum = nil
		}
	}
	//
	if sum != nil {
		// Propagate constant
		return &Constant{sum}
	}
	// Done
	return &Sub{rs}
}

func applyConstantPropagationMul(es []Expr) Expr {
	var one = fr.NewElement(1)
	prod := &one
	rs := make([]Expr, len(es))
	//
	for i, e := range es {
		rs[i] = applyConstantPropagation(e)
		// Check for constant
		c, ok := rs[i].(*Constant)
		//
		if ok && c.Value.IsZero() {
			// No matter what, outcome is zero.
			return &Constant{c.Value}
		} else if ok && prod != nil {
			// Continue building constant
			prod.Mul(prod, c.Value)
		} else {
			prod = nil
		}
	}
	// Attempt to propagate constant
	if prod != nil {
		return &Constant{prod}
	}
	//
	return &Mul{rs}
}

func applyConstantPropagationExp(arg Expr, pow uint64) Expr {
	arg = applyConstantPropagation(arg)
	//
	if c, ok := arg.(*Constant); ok {
		var val fr.Element
		// Clone value
		val.Set(c.Value)
		// Compute exponent (in place)
		util.Pow(&val, pow)
		// Done
		return &Constant{&val}
	}
	//
	return &Exp{arg, pow}
}

func applyConstantPropagationNorm(arg Expr) Expr {
	arg = applyConstantPropagation(arg)
	//
	if c, ok := arg.(*Constant); ok {
		var val fr.Element
		// Clone value
		val.Set(c.Value)
		// Normalise (in place)
		if !val.IsZero() {
			val.SetOne()
		}
		// Done
		return &Constant{&val}
	}
	//
	return &Normalise{arg}
}