package hir

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/schema/constraint"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn captures the essence of a data column at AIR level.
type DataColumn = *assignment.DataColumn

// VanishingConstraint captures the essence of a vanishing constraint at the HIR
// level. A vanishing constraint is a row constraint which must evaluate to
// zero.
type VanishingConstraint = *constraint.VanishingConstraint[ZeroArrayTest]

// LookupConstraint captures the essence of a lookup constraint at the HIR
// level.  To make this work, the UnitExpr adaptor is required, and this means
// certain expression forms cannot be permitted (e.g. the use of lists).
type LookupConstraint = *constraint.LookupConstraint[UnitExpr]

// PropertyAssertion captures the notion of an arbitrary property which should
// hold for all acceptable traces.  However, such a property is not enforced by
// the prover.
type PropertyAssertion = *sc.PropertyAssertion[ZeroArrayTest]

// Permutation captures the notion of a (sorted) permutation at the HIR level.
type Permutation = *assignment.SortedPermutation

// Schema for HIR constraints and columns.
type Schema struct {
	// The modules of the schema
	modules []sc.Module
	// The data columns of this schema.
	inputs []sc.Declaration
	// The sorted permutations of this schema.
	assignments []sc.Assignment
	// Constraints of this schema, which are either vanishing, lookup or type
	// constraints.
	constraints []sc.Constraint
	// The property assertions for this schema.
	assertions []PropertyAssertion
}

// EmptySchema is used to construct a fresh schema onto which new columns and
// constraints will be added.
func EmptySchema() *Schema {
	p := new(Schema)
	p.modules = make([]sc.Module, 0)
	p.inputs = make([]sc.Declaration, 0)
	p.assignments = make([]sc.Assignment, 0)
	p.constraints = make([]sc.Constraint, 0)
	p.assertions = make([]PropertyAssertion, 0)
	// Done
	return p
}

// AddModule adds a new module to this schema, returning its module index.
func (p *Schema) AddModule(name string) uint {
	mid := uint(len(p.modules))
	p.modules = append(p.modules, sc.NewModule(name))

	return mid
}

// AddDataColumn appends a new data column with a given type.  Furthermore, the
// type is enforced by the system when checking is enabled.
func (p *Schema) AddDataColumn(context trace.Context, name string, base sc.Type) uint {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	cid := uint(len(p.inputs))
	p.inputs = append(p.inputs, assignment.NewDataColumn(context, name, base))

	return cid
}

// AddLookupConstraint appends a new lookup constraint.
func (p *Schema) AddLookupConstraint(handle string, source trace.Context, target trace.Context,
	sources []UnitExpr, targets []UnitExpr) {
	if len(targets) != len(sources) {
		panic("differeng number of target / source lookup columns")
	}
	// TODO: sanity source columns are in the source module, and likewise target
	// columns are in the target module (though source != target is permitted).

	// Finally add constraint
	p.constraints = append(p.constraints,
		constraint.NewLookupConstraint(handle, source, target, sources, targets))
}

// AddAssignment appends a new assignment (i.e. set of computed columns) to be
// used during trace expansion for this schema.  Computed columns are introduced
// by the process of lowering from HIR / MIR to AIR.
func (p *Schema) AddAssignment(c sc.Assignment) uint {
	index := p.Columns().Count()
	p.assignments = append(p.assignments, c)

	return index
}

// AddVanishingConstraint appends a new vanishing constraint.
func (p *Schema) AddVanishingConstraint(handle string, context trace.Context, domain *int, expr Expr) {
	if context.Module() >= uint(len(p.modules)) {
		panic(fmt.Sprintf("invalid module index (%d)", context.Module()))
	}

	p.constraints = append(p.constraints,
		constraint.NewVanishingConstraint(handle, context, domain, ZeroArrayTest{expr}))
}

// AddTypeConstraint appends a new range constraint.
func (p *Schema) AddTypeConstraint(target uint, t sc.Type) {
	// Check whether is a field type, as these can actually be ignored.
	if t.AsField() == nil {
		p.constraints = append(p.constraints, constraint.NewTypeConstraint(target, t))
	}
}

// AddPropertyAssertion appends a new property assertion.
func (p *Schema) AddPropertyAssertion(module uint, handle string, property Expr) {
	p.assertions = append(p.assertions, sc.NewPropertyAssertion[ZeroArrayTest](module, handle, ZeroArrayTest{property}))
}

// ============================================================================
// Schema Interface
// ============================================================================

// Inputs returns an array over the input declarations of this sc.  That is,
// the subset of declarations whose trace values must be provided by the user.
func (p *Schema) Inputs() util.Iterator[sc.Declaration] {
	return util.NewArrayIterator(p.inputs)
}

// Assignments returns an array over the assignments of this sc.  That
// is, the subset of declarations whose trace values can be computed from
// the inputs.
func (p *Schema) Assignments() util.Iterator[sc.Assignment] {
	return util.NewArrayIterator(p.assignments)
}

// Columns returns an array over the underlying columns of this sc.
// Specifically, the index of a column in this array is its column index.
func (p *Schema) Columns() util.Iterator[sc.Column] {
	is := util.NewFlattenIterator[sc.Declaration, sc.Column](p.Inputs(),
		func(d sc.Declaration) util.Iterator[sc.Column] { return d.Columns() })
	ps := util.NewFlattenIterator[sc.Assignment, sc.Column](p.Assignments(),
		func(d sc.Assignment) util.Iterator[sc.Column] { return d.Columns() })
	//
	return is.Append(ps)
}

// Constraints returns an array over the underlying constraints of this
// sc.
func (p *Schema) Constraints() util.Iterator[sc.Constraint] {
	return util.NewArrayIterator(p.constraints)
}

// Declarations returns an array over the column declarations of this
// sc.
func (p *Schema) Declarations() util.Iterator[sc.Declaration] {
	ps := util.NewCastIterator[sc.Assignment, sc.Declaration](p.Assignments())
	return p.Inputs().Append(ps)
}

// Modules returns an iterator over the declared set of modules within this
// schema.
func (p *Schema) Modules() util.Iterator[sc.Module] {
	return util.NewArrayIterator(p.modules)
}
