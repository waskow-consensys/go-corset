package gadgets

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/table"
)

// ApplyColumnSortGadget Add sorting constraints for a column where the
// difference between any two rows (i.e. the delta) is constrained to fit within
// a given bitwidth.  The target column is assumed to have an appropriate
// (enforced) bitwidth to ensure overflow cannot arise.  The sorting constraint
// is either ascending (positively signed) or descending (negatively signed).  A
// delta column is added along with bitwidth constraints (where necessary) to
// ensure the delta is within the given width.
//
// This gadget does not attempt to sort the column data during trace expansion,
// and assumes the data either comes sorted or is sorted by some other
// computation.
func ApplyColumnSortGadget(column string, sign bool, bitwidth uint, schema *air.Schema) {
	var deltaName string
	// Configure computation
	Xk := air.NewColumnAccess(column, 0)
	Xkm1 := air.NewColumnAccess(column, -1)
	// Account for sign
	var Xdiff air.Expr
	if sign {
		Xdiff = Xk.Sub(Xkm1)
		deltaName = fmt.Sprintf("+%s", column)
	} else {
		Xdiff = Xkm1.Sub(Xk)
		deltaName = fmt.Sprintf("-%s", column)
	}
	// Add delta column
	schema.AddColumn(deltaName, true)
	// Add diff computation
	schema.AddComputation(table.NewComputedColumn(deltaName, Xdiff))
	// Add necessary bitwidth constraints
	ApplyBitwidthGadget(deltaName, bitwidth, schema)
	// Configure constraint: Delta[k] = X[k] - X[k-1]
	Dk := air.NewColumnAccess(deltaName, 0)
	schema.AddVanishingConstraint(deltaName, nil, Dk.Equate(Xdiff))
}