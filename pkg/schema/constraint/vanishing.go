package constraint

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// ZeroTest is a wrapper which converts an Evaluable expression into a Testable
// constraint.  Specifically, by checking whether or not the given expression
// vanishes (i.e. evaluates to zero).
type ZeroTest[E sc.Evaluable] struct {
	Expr E
}

// TestAt determines whether or not a given expression evaluates to zero.
// Observe that if the expression is undefined, then it is assumed not to hold.
func (p ZeroTest[E]) TestAt(row int, tr trace.Trace) bool {
	val := p.Expr.EvalAt(row, tr)
	return val.IsZero()
}

// Bounds determines the bounds for this zero test.
func (p ZeroTest[E]) Bounds() util.Bounds {
	return p.Expr.Bounds()
}

// Context determines the evaluation context (i.e. enclosing module) for this
// expression.
func (p ZeroTest[E]) Context(schema sc.Schema) trace.Context {
	return p.Expr.Context(schema)
}

// RequiredColumns returns the set of columns on which this term depends.
// That is, columns whose values may be accessed when evaluating this term
// on a given trace.
func (p ZeroTest[E]) RequiredColumns() *util.SortedSet[uint] {
	return p.Expr.RequiredColumns()
}

// RequiredCells returns the set of trace cells on which evaluation of this
// constraint element depends.
func (p ZeroTest[E]) RequiredCells(row int, tr trace.Trace) *util.AnySortedSet[trace.CellRef] {
	return p.Expr.RequiredCells(row, tr)
}

// String generates a human-readble string.
//
//nolint:revive
func (p ZeroTest[E]) String() string {
	return fmt.Sprintf("%s", any(p.Expr))
}

// VanishingFailure provides structural information about a failing vanishing constraint.
type VanishingFailure struct {
	// Handle of the failing constraint
	handle string
	// Row on which the constraint failed
	row uint
	// Cells used by the failing constraint.  This is useful for providing a
	// detailed report including the values of relevant cells.
	cells []trace.CellRef
}

// Message provides a suitable error message
func (p *VanishingFailure) Message() string {
	// Construct useful error message
	return fmt.Sprintf("constraint \"%s\" does not hold (row %d)", p.handle, p.row)
}

func (p *VanishingFailure) String() string {
	return p.Message()
}

// VanishingConstraint specifies a constraint which should hold on every row of the
// table.  The only exception is when the constraint is undefined (e.g. because
// it references a non-existent table cell).  In such case, the constraint is
// ignored.  This is parameterised by the type of the constraint expression.
// Thus, we can reuse this definition across the various intermediate
// representations (e.g. Mid-Level IR, Arithmetic IR, etc).
type VanishingConstraint[T sc.Testable] struct {
	// A unique identifier for this constraint.  This is primarily
	// useful for debugging.
	handle string
	// Evaluation context for this constraint which must match that of the
	// constrained expression itself.
	context trace.Context
	// Indicates (when nil) a global constraint that applies to all rows.
	// Otherwise, indicates a local constraint which applies to the specific row
	// given here.
	domain *int
	// The actual constraint itself (e.g. an expression which
	// should evaluate to zero, etc)
	constraint T
}

// NewVanishingConstraint constructs a new vanishing constraint!
func NewVanishingConstraint[T sc.Testable](handle string, context trace.Context,
	domain *int, constraint T) *VanishingConstraint[T] {
	return &VanishingConstraint[T]{handle, context, domain, constraint}
}

// Handle returns the handle associated with this constraint.
//
//nolint:revive
func (p *VanishingConstraint[T]) Handle() string {
	return p.handle
}

// Constraint returns the vanishing expression itself.
func (p *VanishingConstraint[T]) Constraint() T {
	return p.constraint
}

// Domain returns the domain of this constraint.  If the domain is nil, then
// this is a global constraint.  Otherwise this signals a local constraint which
// applies to a specific row (e.g. the first or last).
func (p *VanishingConstraint[T]) Domain() *int {
	return p.domain
}

// Context returns the evaluation context for this constraint.  Every constraint
// must be situated within exactly one module in order to be well-formed.
func (p *VanishingConstraint[T]) Context() trace.Context {
	return p.context
}

// Accepts checks whether a vanishing constraint evaluates to zero on every row
// of a table.  If so, return nil otherwise return an error.
//
//nolint:revive
func (p *VanishingConstraint[T]) Accepts(tr trace.Trace) schema.Failure {
	if p.domain == nil {
		// Global Constraint
		return HoldsGlobally(p.handle, p.context, p.constraint, tr)
	}
	// Local constraint
	var start uint
	// Handle negative domains
	if *p.domain < 0 {
		// Determine height of enclosing module
		height := tr.Height(p.context)
		// Negative rows calculated from end of trace.
		start = height + uint(*p.domain)
	} else {
		start = uint(*p.domain)
	}
	// Check specific row
	return HoldsLocally(start, p.handle, p.constraint, tr)
}

// HoldsGlobally checks whether a given expression vanishes (i.e. evaluates to
// zero) for all rows of a trace.  If not, report an appropriate error.
func HoldsGlobally[T sc.Testable](handle string, ctx trace.Context, constraint T, tr trace.Trace) schema.Failure {
	// Determine height of enclosing module
	height := tr.Height(ctx)
	// Determine well-definedness bounds for this constraint
	bounds := constraint.Bounds()
	// Sanity check enough rows
	if bounds.End < height {
		// Check all in-bounds values
		for k := bounds.Start; k < (height - bounds.End); k++ {
			if err := HoldsLocally(k, handle, constraint, tr); err != nil {
				return err
			}
		}
	}
	// Success
	return nil
}

// HoldsLocally checks whether a given constraint holds (e.g. vanishes) on a
// specific row of a trace. If not, report an appropriate error.
func HoldsLocally[T sc.Testable](k uint, handle string, constraint T, tr trace.Trace) schema.Failure {
	// Check whether it holds or not
	if !constraint.TestAt(int(k), tr) {
		//cells := constraint.RequiredCells(int(k), tr).ToArray()
		cells := make([]trace.CellRef, 0)
		// Evaluation failure
		return &VanishingFailure{handle, k, cells}
	}
	// Success
	return nil
}

// String generates a human-readble string.
//
//nolint:revive
func (p *VanishingConstraint[T]) String() string {
	if p.domain == nil {
		return fmt.Sprintf("(vanish %s %s)", p.handle, any(p.constraint))
	} else if *p.domain == 0 {
		return fmt.Sprintf("(vanish:first %s %s)", p.handle, any(p.constraint))
	}
	//
	return fmt.Sprintf("(vanish:last %s %s)", p.handle, any(p.constraint))
}
