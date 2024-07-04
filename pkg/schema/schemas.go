package schema

import (
	"errors"
	"math"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// JoinContexts combines one or more evaluation contexts together.  There are a
// number of scenarios.  The simple path is when each expression has the same
// evaluation context (in which case this is returned).  Its also possible one
// or more expressions have no evaluation context (signaled by math.MaxUint) and
// this can be ignored. Finally, we might have two expressions with conflicting
// evaluation contexts, and this clearly signals an error.
func JoinContexts[E Contextual](args []E, schema Schema) (uint, uint, bool) {
	var mid uint = math.MaxUint

	var multiplier = uint(1)
	//
	for _, e := range args {
		c, m, b := e.Context(schema)
		if !b {
			// Indicates conflict detected upstream, therefore propagate this
			// down.
			return 0, 0, false
		} else if mid == math.MaxUint {
			// No evaluation context determined yet, therefore can overwrite
			// with whatever we got.  Observe that this might still actually
			mid = c
			multiplier = m
		} else if c != math.MaxUint && (c != mid || m != multiplier) {
			// This indicates a conflict is detected, therefore we must
			// propagate this down.
			return 0, 0, false
		}
	}
	// If we get here, then no conflicts were detected.
	return mid, multiplier, true
}

// DetermineEnclosingModuleOfExpression determines (and checks) the enclosing
// module for a given expression.  The expectation is that there is a single
// enclosing module, and this function will panic if that does not hold.
func DetermineEnclosingModuleOfExpression[E Contextual](expr E, schema Schema) (uint, uint) {
	if mid, multiplier, ok := expr.Context(schema); ok && mid != math.MaxUint {
		return mid, multiplier
	}
	//
	panic("expression has no evaluation context")
}

// DetermineEnclosingModuleOfExpressions determines (and checks) the enclosing
// module for a given set of expressions.  The expectation is that there is a single
// enclosing module, and this function will panic if that does not hold.
func DetermineEnclosingModuleOfExpressions[E Contextual](exprs []E, schema Schema) (uint, uint, error) {
	// Sanity check input
	if len(exprs) == 0 {
		panic("cannot determine enclosing module for empty expression array")
	}
	// Determine first
	mid, multiplier, ok := exprs[0].Context(schema)
	// Sanity check this made sense
	if !ok {
		return 0, 0, errors.New("conflicting enclosing modules")
	}
	// Check rest against this
	for i := 1; i < len(exprs); i++ {
		m, f, ok := exprs[i].Context(schema)
		if !ok {
			return uint(i), 0, errors.New("conflicting enclosing modules")
		} else if mid == math.MaxUint {
			mid = m
		} else if m != math.MaxUint && m != mid {
			return uint(i), 0, errors.New("conflicting enclosing modules")
		} else if m != math.MaxUint && f != multiplier {
			return uint(i), 0, errors.New("conflicting length multipliers")
		}
	}
	// success
	return mid, multiplier, nil
}

// DetermineEnclosingModuleOfColumns determines (and checks) the enclosing module for a
// given set of columns.  The expectation is that there is a single enclosing
// module, and this function will panic if that does not hold.
func DetermineEnclosingModuleOfColumns(cols []uint, schema Schema) (uint, uint) {
	head := schema.Columns().Nth(cols[0])
	// First, determine module of first column.
	mid := head.Module()
	multiplier := head.LengthMultiplier()
	// Second, check other columns in the same module.
	//
	// NOTE: this could potentially be made more efficient by checking the
	// columns of the module for the first column.
	for i := 1; i < len(cols); i++ {
		col := schema.Columns().Nth(cols[i])
		if mid != col.Module() {
			// This is an internal failure which should be prevented by upstream
			// checking (e.g. in the parser).
			panic("columns have different enclosing module")
		} else if multiplier != col.LengthMultiplier() {
			panic("columns have different length multipliers")
		}
	}
	// Done
	return mid, multiplier
}

// RequiredSpillage returns the minimum amount of spillage required to ensure
// valid traces are accepted in the presence of arbitrary padding.  Spillage can
// only arise from computations as this is where values outside of the user's
// control are determined.
func RequiredSpillage(schema Schema) uint {
	// Ensures always at least one row of spillage (referred to as the "initial
	// padding row")
	mx := uint(1)
	// Determine if any more spillage required
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		// Incorporate its spillage requirements
		mx = max(mx, ith.RequiredSpillage())
	}

	return mx
}

// ExpandTrace expands a given trace according to this schema.  More
// specifically, that means computing the actual values for any assignments.
// Observe that assignments have to be computed in the correct order.
func ExpandTrace(schema Schema, trace tr.Trace) error {
	// Compute each assignment in turn
	for i := schema.Assignments(); i.HasNext(); {
		// Get ith assignment
		ith := i.Next()
		// Compute ith assignment(s)
		if err := ith.ExpandTrace(trace); err != nil {
			return err
		}
	}
	// Done
	return nil
}

// Accepts determines whether this schema will accept a given trace.  That
// is, whether or not the given trace adheres to the schema.  A trace can fail
// to adhere to the schema for a variety of reasons, such as having a constraint
// which does not hold.
//
//nolint:revive
func Accepts(schema Schema, trace tr.Trace) error {
	// Check each constraint in turn
	for i := schema.Constraints(); i.HasNext(); {
		// Get ith constraint
		ith := i.Next()
		// Check it holds (or report an error)
		if err := ith.Accepts(trace); err != nil {
			return err
		}
	}
	// Success
	return nil
}

// ColumnIndexOf returns the column index of the column with the given name, or
// returns false if no matching column exists.
func ColumnIndexOf(schema Schema, module uint, name string) (uint, bool) {
	return schema.Columns().Find(func(c Column) bool {
		return c.Module() == module && c.Name() == name
	})
}