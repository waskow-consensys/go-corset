package corset

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/util"
)

// Circuit represents the root of the Abstract Syntax Tree.  This is also
// referred to as the "prelude".  All modules are contained within the root, and
// declarations can also be declared here as well.
type Circuit struct {
	Modules      []Module
	Declarations []Declaration
}

// Module represents a top-level module declaration.  This corresponds to a
// table in the final constraint set.
type Module struct {
	Name         string
	Declarations []Declaration
}

// Node provides common functionality across all elements of the Abstract Syntax
// Tree.  For example, it ensures every element can converted back into Lisp
// form for debugging.  Furthermore, it provides a reference point for
// constructing a suitable source map for reporting syntax errors.
type Node interface {
	// Convert this node into its lisp representation.  This is primarily used
	// for debugging purposes.
	Lisp() sexp.SExp
}

// Declaration represents a top-level declaration in a Corset source file (e.g.
// defconstraint, defcolumns, etc).
type Declaration interface {
	Node
	// Returns the set of symbols being defined this declaration.  Observe that
	// these may not yet have been finalised.
	Definitions() util.Iterator[SymbolDefinition]
	// Return set of columns on which this declaration depends.
	Dependencies() util.Iterator[Symbol]
	// Check whether this declaration defines a given symbol.  The symbol in
	// question needs to have been resolved already for this to make sense.
	Defines(Symbol) bool
	// Check whether this declaration is finalised already.
	IsFinalised() bool
}

// ============================================================================
// defalias
// ============================================================================

// DefAliases represents the declaration of one or more aliases.  That is,
// alternate names for existing symbols.
type DefAliases struct {
	// Distinguishes defalias from defunalias
	functions bool
	// Aliases
	aliases []*DefAlias
	// Symbols being aliased
	symbols []Symbol
}

// Dependencies needed to signal declaration.
func (p *DefAliases) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator[Symbol](nil)
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefAliases) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefAliases) Defines(symbol Symbol) bool {
	// fine beause defaliases gets special treatement.
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefAliases) IsFinalised() bool {
	// Fine because defaliases doesn't really do anything with its symbols.
	return true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
//
//nolint:revive
func (p *DefAliases) Lisp() sexp.SExp {
	pairs := sexp.EmptyList()
	//
	for i, a := range p.aliases {
		pairs.Append(sexp.NewSymbol(a.name))
		pairs.Append(p.symbols[i].Lisp())
	}
	//
	var name *sexp.Symbol
	//
	if p.functions {
		name = sexp.NewSymbol("defunalias")
	} else {
		name = sexp.NewSymbol("defalias")
	}
	//
	return sexp.NewList([]sexp.SExp{
		name, pairs,
	})
}

// DefAlias provides a node on which to hang source information to an alias name.
type DefAlias struct {
	// Name of the alias
	name string
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
//
//nolint:revive
func (p *DefAlias) Lisp() sexp.SExp {
	return sexp.NewSymbol(p.name)
}

// ============================================================================
// defcolumns
// ============================================================================

// DefColumns captures a set of one or more columns being declared.
type DefColumns struct {
	Columns []*DefColumn
}

// Dependencies needed to signal declaration.
func (p *DefColumns) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator[Symbol](nil)
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefColumns) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewArrayIterator(p.Columns)
	return util.NewCastIterator[*DefColumn, SymbolDefinition](iter)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefColumns) Defines(symbol Symbol) bool {
	for _, sym := range p.Columns {
		if &sym.binding == symbol.Binding() {
			return true
		}
	}
	//
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefColumns) IsFinalised() bool {
	return true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefColumns) Lisp() sexp.SExp {
	list := sexp.EmptyList()
	list.Append(sexp.NewSymbol("defcolumns"))
	// Add lisp for each individual column
	for _, c := range p.Columns {
		list.Append(c.Lisp())
	}
	// Done
	return list
}

// DefColumn packages together those piece relevant to declaring an individual
// column, such its name and type.
type DefColumn struct {
	// Binding of this column (which may or may not be finalised).
	binding ColumnBinding
}

var _ SymbolDefinition = &DefColumn{}

// IsFunction is never true for a column definition.
func (e *DefColumn) IsFunction() bool {
	return false
}

// Binding returns the allocated binding for this symbol (which may or may not
// be finalised).
func (e *DefColumn) Binding() Binding {
	return &e.binding
}

// Name returns the (unqualified) name of this symbol.  For example, "X" for
// a column X defined in a module m1.
func (e *DefColumn) Name() string {
	return e.binding.path.Tail()
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (e *DefColumn) Path() *util.Path {
	return &e.binding.path
}

// DataType returns the type of this column.  If this column have not yet been
// finalised, then this will panic.
func (e *DefColumn) DataType() Type {
	if !e.binding.IsFinalised() {
		panic("unfinalised column")
	}
	//
	return e.binding.dataType
}

// LengthMultiplier returns the length multiplier of this column (where the
// height of this column is determined as the product of the enclosing module's
// height and this length multiplier).  If this column have not yet been
// finalised, then this will panic.
func (e *DefColumn) LengthMultiplier() uint {
	if !e.binding.IsFinalised() {
		panic("unfinalised column")
	}
	//
	return e.binding.multiplier
}

// MustProve determines whether or not the type of this column must be
// established by the prover (e.g. a range constraint or similar).
func (e *DefColumn) MustProve() bool {
	if !e.binding.IsFinalised() {
		panic("unfinalised column")
	}
	//
	return e.binding.mustProve
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (e *DefColumn) Lisp() sexp.SExp {
	list := sexp.EmptyList()
	list.Append(sexp.NewSymbol(e.Name()))
	//
	if e.binding.dataType != nil {
		datatype := e.binding.dataType.String()
		if e.binding.mustProve {
			datatype = fmt.Sprintf("%s@prove", datatype)
		}

		list.Append(sexp.NewSymbol(datatype))
	}
	//
	if e.binding.multiplier != 1 {
		list.Append(sexp.NewSymbol(":multiplier"))
		list.Append(sexp.NewSymbol(fmt.Sprintf("%d", e.binding.multiplier)))
	}
	//
	if list.Len() == 1 {
		return list.Get(0)
	}
	//
	return list
}

// ============================================================================
// defconst
// ============================================================================

// DefConst represents the declaration of one of more constant values which can
// be used within expressions to improve readability.
type DefConst struct {
	// List of constant pairs.  Observe that every expression in this list must
	// be constant (i.e. it cannot refer to column values or call impure
	// functions, etc).
	constants []*DefConstUnit
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefConst) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewArrayIterator[*DefConstUnit](p.constants)
	return util.NewCastIterator[*DefConstUnit, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefConst) Dependencies() util.Iterator[Symbol] {
	var deps []Symbol
	// Combine dependencies from all constants defined within.
	for _, d := range p.constants {
		deps = append(deps, d.binding.value.Dependencies()...)
	}
	// Done
	return util.NewArrayIterator[Symbol](deps)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefConst) Defines(symbol Symbol) bool {
	for _, sym := range p.constants {
		if &sym.binding == symbol.Binding() {
			return true
		}
	}
	//
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefConst) IsFinalised() bool {
	for _, c := range p.constants {
		if !c.binding.IsFinalised() {
			return false
		}
	}
	//
	return true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefConst) Lisp() sexp.SExp {
	def := sexp.EmptyList()
	def.Append(sexp.NewSymbol("defconst"))
	//
	for _, c := range p.constants {
		def.Append(sexp.NewSymbol(c.Name()))
		def.Append(c.binding.value.Lisp())
	}
	// Done
	return def
}

// DefConstUnit represents the definition of exactly one constant value.  As
// such, this is an instance of SymbolDefinition and provides a binding.
type DefConstUnit struct {
	// Binding for this constant.
	binding ConstantBinding
}

// IsFunction is never true for a constant definition.
func (e *DefConstUnit) IsFunction() bool {
	return false
}

// Binding returns the allocated binding for this symbol (which may or may not
// be finalised).
func (e *DefConstUnit) Binding() Binding {
	return &e.binding
}

// Name returns the (unqualified) name of this symbol.  For example, "X" for
// a column X defined in a module m1.
func (e *DefConstUnit) Name() string {
	return e.binding.path.Tail()
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (e *DefConstUnit) Path() *util.Path {
	return &e.binding.path
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
//
//nolint:revive
func (e *DefConstUnit) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol(e.Name()),
		e.binding.value.Lisp()})
}

// ============================================================================
// defconstraint
// ============================================================================

// DefConstraint represents a vanishing constraint, which is either "local" or
// "global".  A local constraint applies either to the first or last rows,
// whilst a global constraint applies to all rows.  For a constraint to hold,
// its expression must evaluate to zero for the rows on which it is active.  A
// constraint may also have a "guard" which is an expression that must evaluate
// to a non-zero value for the constraint to be considered active.  The
// expression for a constraint must have a single context.  That is, it can only
// be applied to columns within the same module (i.e. to ensure they have the
// same height).  Furthermore, within a given module, we require that all
// columns accessed by the constraint have the same length multiplier.
type DefConstraint struct {
	// Unique handle given to this constraint.  This is primarily useful for
	// debugging (i.e. so we know which constaint failed, etc).
	Handle string
	// Domain of this constraint which, if empty, indicates a global constraint.
	// Otherwise, a given value indicates a single row on which this constraint
	// should apply (where negative values are taken from the end, meaning that
	// -1 represents the last row of a given module).
	Domain util.Option[int]
	// A selector which determines for which rows this constraint is active.
	// Specifically, when the expression evaluates to a non-zero value then the
	// constraint is active; otherwiser, its inactive. Nil is permitted to
	// indicate no guard is present.
	Guard Expr
	// Perspective identifies the perspective to which this constraint is
	// associated (if any).
	Perspective *PerspectiveName
	// The constraint itself which (when active) should evaluate to zero for the
	// relevant set of rows.
	Constraint Expr
	//
	finalised bool
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefConstraint) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefConstraint) Dependencies() util.Iterator[Symbol] {
	var deps []Symbol
	// Extract guard's dependencies (if applicable)
	if p.Guard != nil {
		deps = p.Guard.Dependencies()
	}
	// Extract perspective (if applicable)
	if p.Perspective != nil {
		deps = append(deps, p.Perspective)
	}
	// Extract bodies dependencies
	deps = append(deps, p.Constraint.Dependencies()...)
	// Done
	return util.NewArrayIterator[Symbol](deps)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefConstraint) Defines(symbol Symbol) bool {
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefConstraint) IsFinalised() bool {
	return p.finalised
}

// Finalise this declaration, which means that its guard (if applicable) and
// body have been resolved.
func (p *DefConstraint) Finalise() {
	p.finalised = true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefConstraint) Lisp() sexp.SExp {
	modifiers := sexp.EmptyList()
	// domain
	if p.Domain.HasValue() {
		domain := fmt.Sprintf("{%d}", p.Domain.Unwrap())
		//
		modifiers.Append(sexp.NewSymbol(":domain"))
		modifiers.Append(sexp.NewSymbol(domain))
	}
	//
	if p.Guard != nil {
		modifiers.Append(sexp.NewSymbol(":guard"))
		modifiers.Append(p.Guard.Lisp())
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("defconstraint"),
		sexp.NewSymbol(p.Handle),
		modifiers,
		p.Constraint.Lisp()})
}

// ============================================================================
// definrange
// ============================================================================

// DefInRange restricts all values for a given expression to be within a range
// [0..n) for some bound n.  Any bound is supported, and the system will choose
// the best underlying implementation as needed.
type DefInRange struct {
	// The expression whose values are being constrained to within the given
	// bound.
	Expr Expr
	// The upper bound for this constraint.  Specifically, every evaluation of
	// the expression should produce a value strictly below this bound.  NOTE:
	// an fr.Element is used here to store the bound simply to make the
	// necessary comparison against table data more direct.
	Bound fr.Element
	// Indicates whether or not the expression has been resolved.
	finalised bool
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefInRange) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefInRange) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator[Symbol](p.Expr.Dependencies())
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefInRange) Defines(symbol Symbol) bool {
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefInRange) IsFinalised() bool {
	return p.finalised
}

// Finalise this declaration, meaning that the expression has been resolved.
func (p *DefInRange) Finalise() {
	p.finalised = true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefInRange) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("definrange"),
		p.Expr.Lisp(),
		sexp.NewSymbol(p.Bound.String()),
	})
}

// ============================================================================
// definterleaved
// ============================================================================

// DefInterleaved generates a new column by interleaving two or more existing
// colummns.  For example, say Z interleaves X and Y (in that order) and we have
// a trace X=[1,2], Y=[3,4].  Then, the interleaved column Z has the values
// Z=[1,3,2,4].  All columns must be defined within the same context.  Finally,
// the type of the interleaved column is the widest type of any source columns.
// For example, consider an interleaving of two columns X and Y with types i16
// and i8 respectively.  Then, the type of the resulting column is i16 (as this
// is required to hold an element from any source column).
type DefInterleaved struct {
	// The target column being defined
	Target *DefColumn
	// The source columns used to define the interleaved target column.
	Sources []Symbol
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefInterleaved) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewUnitIterator(p.Target)
	return util.NewCastIterator[*DefColumn, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefInterleaved) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Sources)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefInterleaved) Defines(symbol Symbol) bool {
	return &p.Target.binding == symbol.Binding()
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefInterleaved) IsFinalised() bool {
	return p.Target.binding.IsFinalised()
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefInterleaved) Lisp() sexp.SExp {
	sources := make([]sexp.SExp, len(p.Sources))
	// Sources
	for i, t := range p.Sources {
		sources[i] = t.Lisp()
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("definterleaved"),
		p.Target.Lisp(),
		sexp.NewList(sources),
	})
}

// ============================================================================
// deflookup
// ============================================================================

// DefLookup represents a lookup constraint between a set N of source
// expressions and a set of N target expressions.  The source expressions must
// have a single context (i.e. all be in the same module) and likewise for the
// target expressions (though the source and target contexts can differ).  The
// constraint can be viewed as a "subset constraint".  Let the set of "source
// tuples" be those obtained by evaluating the source expressions over all rows
// in the source context, and likewise the "target tuples" those for the target
// expressions in the target context.  Then the lookup constraint holds if the
// set of source tuples is a subset of the target tuples.  This does not need to
// be a strict subset, so the two sets can be identical.  Furthermore, these are
// not treated as multi-sets, hence the number of occurrences of a given tuple
// is not relevant.
type DefLookup struct {
	// Unique handle given to this constraint.  This is primarily useful for
	// debugging (i.e. so we know which constaint failed, etc).
	Handle string
	// Source expressions for lookup (i.e. these values must all be contained
	// within the targets).
	Sources []Expr
	// Target expressions for lookup (i.e. these values must contain all of the
	// source values, but may contain more).
	Targets []Expr
	// Indicates whether or not target and source expressions have been resolved.
	finalised bool
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefLookup) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefLookup) Dependencies() util.Iterator[Symbol] {
	sourceDeps := DependenciesOfExpressions(p.Sources)
	targetDeps := DependenciesOfExpressions(p.Targets)
	// Combine deps
	return util.NewArrayIterator(append(sourceDeps, targetDeps...))
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefLookup) Defines(symbol Symbol) bool {
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefLookup) IsFinalised() bool {
	return p.finalised
}

// Finalise this declaration, which means that all source and target expressions
// have been resolved.
func (p *DefLookup) Finalise() {
	p.finalised = true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefLookup) Lisp() sexp.SExp {
	targets := make([]sexp.SExp, len(p.Targets))
	sources := make([]sexp.SExp, len(p.Sources))
	// Targets
	for i, t := range p.Targets {
		targets[i] = t.Lisp()
	}
	// Sources
	for i, t := range p.Sources {
		sources[i] = t.Lisp()
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("deflookup"),
		sexp.NewSymbol(p.Handle),
		sexp.NewList(targets),
		sexp.NewList(sources),
	})
}

// ============================================================================
// defpermutation
// ============================================================================

// DefPermutation represents a (lexicographically sorted) permutation of a set
// of source columns in a given source context, manifested as an assignment to a
// corresponding set of target columns.  The sort direction for each of the
// source columns can be specified as increasing or decreasing.
type DefPermutation struct {
	Targets []*DefColumn
	Sources []Symbol
	Signs   []bool
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefPermutation) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewArrayIterator(p.Targets)
	return util.NewCastIterator[*DefColumn, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefPermutation) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Sources)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefPermutation) Defines(symbol Symbol) bool {
	for _, col := range p.Targets {
		if &col.binding == symbol.Binding() {
			return true
		}
	}
	// Done
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefPermutation) IsFinalised() bool {
	for _, col := range p.Targets {
		if !col.binding.IsFinalised() {
			return false
		}
	}
	// Done
	return true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefPermutation) Lisp() sexp.SExp {
	targets := make([]sexp.SExp, len(p.Targets))
	sources := make([]sexp.SExp, len(p.Sources))
	// Targets
	for i, t := range p.Targets {
		targets[i] = t.Lisp()
	}
	// Sources
	for i, t := range p.Sources {
		var sign string
		if p.Signs[i] {
			sign = "+"
		} else {
			sign = "-"
		}
		//
		sources[i] = sexp.NewList([]sexp.SExp{
			sexp.NewSymbol(sign),
			t.Lisp()})
	}
	//
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("defpermutation"),
		sexp.NewList(targets),
		sexp.NewList(sources)})
}

// ============================================================================
// defperspective
// ============================================================================

// DefPerspective captures the definition of a perspective, its selector and  a
// set of one or more columns being declared within.
type DefPerspective struct {
	// Name of the perspective.
	symbol *PerspectiveName
	// Selector for the perspective.
	Selector Expr
	// Columns defined in this perspective.
	Columns []*DefColumn
}

// Name returns the (unqualified) name of this symbol.  For example, "X" for
// a column X defined in a module m1.
func (p *DefPerspective) Name() string {
	return p.symbol.Path().Tail()
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (p *DefPerspective) Path() *util.Path {
	return &p.symbol.path
}

// IsFunction is always true for a function definition!
func (p *DefPerspective) IsFunction() bool {
	return false
}

// Finalise this perspective, which indicates the selector expression has been
// finalised.
func (p *DefPerspective) Finalise() {
	p.symbol.binding.Finalise()
}

// Binding returns the allocated binding for this symbol (which may or may not
// be finalised).
func (p *DefPerspective) Binding() Binding {
	return p.symbol.binding
}

// Dependencies needed to signal declaration.
func (p *DefPerspective) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Selector.Dependencies())
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefPerspective) Definitions() util.Iterator[SymbolDefinition] {
	iter1 := util.NewArrayIterator(p.Columns)
	iter2 := util.NewCastIterator[*DefColumn, SymbolDefinition](iter1)
	iter3 := util.NewUnitIterator[SymbolDefinition](p)
	// Construct casting iterator
	return iter2.Append(iter3)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefPerspective) Defines(symbol Symbol) bool {
	for _, sym := range p.Columns {
		if &sym.binding == symbol.Binding() {
			return true
		}
	}
	//
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefPerspective) IsFinalised() bool {
	return p.symbol.binding.IsFinalised()
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefPerspective) Lisp() sexp.SExp {
	panic("todo")
}

// ============================================================================
// defproperty
// ============================================================================

// DefProperty represents an assertion to be used only for debugging / testing /
// verification.  Unlike vanishing constraints, property assertions do not
// represent something that the prover can enforce.  Rather, they represent
// properties which are expected to hold for every valid trace. That is, they
// should be implied by the actual constraints.  Thus, whilst the prover cannot
// enforce such properties, external tools (such as for formal verification) can
// attempt to ensure they do indeed always hold.
type DefProperty struct {
	// Unique handle given to this constraint.  This is primarily useful for
	// debugging (i.e. so we know which constaint failed, etc).
	Handle string
	// The assertion itself which (when active) should evaluate to zero for the
	// relevant set of rows.
	Assertion Expr
	// Indicates whether or not the assertion has been resolved.
	finalised bool
}

// Definitions returns the set of symbols defined by this declaration.  Observe that
// these may not yet have been finalised.
func (p *DefProperty) Definitions() util.Iterator[SymbolDefinition] {
	return util.NewArrayIterator[SymbolDefinition](nil)
}

// Dependencies needed to signal declaration.
func (p *DefProperty) Dependencies() util.Iterator[Symbol] {
	return util.NewArrayIterator(p.Assertion.Dependencies())
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefProperty) Defines(symbol Symbol) bool {
	return false
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefProperty) IsFinalised() bool {
	return p.finalised
}

// Finalise this property, meaning that the assertion has been resolved.
func (p *DefProperty) Finalise() {
	p.finalised = true
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefProperty) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("defproperty"),
		sexp.NewSymbol(p.Handle),
		p.Assertion.Lisp()})
}

// ============================================================================
// depurefun & defun
// ============================================================================

// DefFun represents defines a (possibly pure) "function" (which, in actuality,
// is more like a macro).  Specifically, whenever an invocation of this function
// is encountered we can imagine that, in the final constraint set, the body of
// the function is inlined at the point of the call.  A pure function is not
// permitted to access any columns in scope (i.e. it can only manipulate its
// parameters).  In contrast, an impure function can access those columns
// defined within its enclosing context.
type DefFun struct {
	symbol *FunctionName
	// Parameters
	parameters []*DefParameter
}

var _ SymbolDefinition = &DefFun{}

// IsFunction is always true for a function definition!
func (p *DefFun) IsFunction() bool {
	return true
}

// IsPure indicates whether or not this is a pure function.  That is, a function
// which is not permitted to access any columns from the enclosing environment
// (either directly itself, or indirectly via functions it calls).
func (p *DefFun) IsPure() bool {
	return p.symbol.binding.pure
}

// Parameters returns information about the parameters defined by this
// declaration.
func (p *DefFun) Parameters() []*DefParameter {
	return p.parameters
}

// Body Access information about the parameters defined by this declaration.
func (p *DefFun) Body() Expr {
	return p.symbol.binding.body
}

// Binding returns the allocated binding for this symbol (which may or may not
// be finalised).
func (p *DefFun) Binding() Binding {
	return p.symbol.binding
}

// Name returns the (unqualified) name of this symbol.  For example, "X" for
// a column X defined in a module m1.
func (p *DefFun) Name() string {
	return p.symbol.Path().Tail()
}

// Path returns the qualified name (i.e. absolute path) of this symbol.  For
// example, "m1.X" for a column X defined in module m1.
func (p *DefFun) Path() *util.Path {
	return &p.symbol.path
}

// Finalise this declaration
func (p *DefFun) Finalise() {
	p.symbol.binding.Finalise()
}

// Definitions returns the set of symbols defined by this declaration.  Observe
// that these may not yet have been finalised.
func (p *DefFun) Definitions() util.Iterator[SymbolDefinition] {
	iter := util.NewUnitIterator(p.symbol)
	return util.NewCastIterator[*FunctionName, SymbolDefinition](iter)
}

// Dependencies needed to signal declaration.
func (p *DefFun) Dependencies() util.Iterator[Symbol] {
	deps := p.symbol.binding.body.Dependencies()
	ndeps := make([]Symbol, 0)
	// Filter out all parameters declared in this function, since these are not
	// external dependencies.
	for _, d := range deps {
		n := d.Path()
		if n.IsAbsolute() || d.IsFunction() || n.Depth() > 1 || !p.hasParameter(n.Head()) {
			ndeps = append(ndeps, d)
		}
	}
	// Done
	return util.NewArrayIterator(ndeps)
}

// Defines checks whether this declaration defines the given symbol.  The symbol
// in question needs to have been resolved already for this to make sense.
func (p *DefFun) Defines(symbol Symbol) bool {
	return p.symbol.binding == symbol.Binding()
}

// IsFinalised checks whether this declaration has already been finalised.  If
// so, then we don't need to finalise it again.
func (p *DefFun) IsFinalised() bool {
	return p.symbol.binding.IsFinalised()
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefFun) Lisp() sexp.SExp {
	return sexp.NewList([]sexp.SExp{
		sexp.NewSymbol("defun"),
		sexp.NewSymbol(p.symbol.path.Tail()),
		sexp.NewSymbol("..."), // todo
	})
}

// hasParameter checks whether this function has a parameter with the given
// name, or not.
func (p *DefFun) hasParameter(name string) bool {
	for _, v := range p.parameters {
		if v.Binding.name == name {
			return true
		}
	}
	//
	return false
}

// DefParameter packages together those piece relevant to declaring an individual
// parameter, such its name and type.
type DefParameter struct {
	Binding LocalVariableBinding
}

// Lisp converts this node into its lisp representation.  This is primarily used
// for debugging purposes.
func (p *DefParameter) Lisp() sexp.SExp {
	return sexp.NewSymbol(p.Binding.name)
}
