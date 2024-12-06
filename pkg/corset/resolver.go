package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/util"
)

// ResolveCircuit resolves all symbols declared and used within a circuit,
// producing an environment which can subsequently be used to look up the
// relevant module or column identifiers.  This process can fail, of course, it
// a symbol (e.g. a column) is referred to which doesn't exist.  Likewise, if
// two modules or columns with identical names are declared in the same scope,
// etc.
func ResolveCircuit(srcmap *sexp.SourceMaps[Node], circuit *Circuit) (*GlobalScope, []SyntaxError) {
	// Construct top-level scope
	scope := NewGlobalScope()
	// Register the root module (which should always exist)
	scope.DeclareModule("")
	// Register other modules
	for _, m := range circuit.Modules {
		scope.DeclareModule(m.Name)
	}
	// Construct resolver
	r := resolver{srcmap}
	// Allocate declared input columns
	errs := r.resolveDeclarations(scope, circuit)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Done
	return scope, errs
}

// Resolver packages up information necessary for resolving a circuit and
// checking that everything makes sense.
type resolver struct {
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
}

// Process all assignment column declarations.  These are more complex than for
// input columns, since there can be dependencies between them.  Thus, we cannot
// simply resolve them in one linear scan.
func (r *resolver) resolveDeclarations(scope *GlobalScope, circuit *Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.resolveDeclarationsInModule(scope.Module(""), circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.resolveDeclarationsInModule(scope.Module(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Resolve all columns declared in a given module.  This is tricky because
// assignments can depend on the declaration of other columns.  Hence, we have
// to process all columns before we can sure that they are all declared
// correctly.
func (r *resolver) resolveDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	if errors := r.initialiseDeclarationsInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	// Iterate until all columns finalised
	return r.finaliseDeclarationsInModule(scope, decls)
}

// Initialise all declarations in the given module scope.  That means allocating
// all bindings into the scope, whilst also ensuring that we never have two
// bindings for the same symbol, etc.  The key is that, at this stage, all
// bindings are potentially "non-finalised".  That means they may be missing key
// information which is yet to be determined (e.g. information about types, or
// contexts, etc).
func (r *resolver) initialiseDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	module := scope.EnclosingModule()
	errors := make([]SyntaxError, 0)
	//
	for _, d := range decls {
		for iter := d.Definitions(); iter.HasNext(); {
			def := iter.Next()
			// Attempt to declare symbol
			if !scope.Declare(def) {
				msg := fmt.Sprintf("symbol %s already declared in %s", def.Name(), module)
				err := r.srcmap.SyntaxError(def, msg)
				errors = append(errors, *err)
			}
		}
	}
	// Done
	return errors
}

// Finalise all declarations given in a module.  This requires an iterative
// process as we cannot finalise a declaration until all of its dependencies
// have been themselves finalised.  For example, a function which depends upon
// an interleaved column.  Until the interleaved column is finalised, its type
// won't be available and, hence, we cannot type the function.
func (r *resolver) finaliseDeclarationsInModule(scope *ModuleScope, decls []Declaration) []SyntaxError {
	// Changed indicates whether or not a new assignment was finalised during a
	// given iteration.  This is important to know since, if the assignment is
	// not complete and we didn't finalise any more assignments --- then, we've
	// reached a fixed point where the final assignment is incomplete (i.e.
	// there is some error somewhere).
	changed := true
	// Complete tells us whether or not the assignment is complete.  The
	// assignment is not complete if there it at least one declaration which is
	// not yet finalised.
	complete := false
	// For an incomplete assignment, this identifies the last declaration that
	// could not be finalised (i.e. as an example so we have at least one for
	// error reporting).
	var (
		incomplete Node = nil
		counter    uint = 4
	)
	//
	for changed && !complete && counter > 0 {
		errors := make([]SyntaxError, 0)
		changed = false
		complete = true
		//
		for _, d := range decls {
			ready, errs := r.declarationDependenciesAreFinalised(scope, d.Dependencies())
			// See what arosed
			if errs != nil {
				errors = append(errors, errs...)
			} else if ready {
				// Finalise declaration and handle errors
				errs := r.finaliseDeclaration(scope, d)
				errors = append(errors, errs...)
				// Record that a new assignment is available.
				changed = changed || len(errs) == 0
			} else {
				// Declaration not ready yet
				complete = false
				incomplete = d
			}
		}
		// Sanity check for any errors caught during this iteration.
		if len(errors) > 0 {
			return errors
		}
		// Decrement counter
		counter--
	}
	// Check whether we actually finished the allocation.
	if counter == 0 {
		err := r.srcmap.SyntaxError(incomplete, "unable to complete resolution")
		return []SyntaxError{*err}
	} else if !complete {
		// No, we didn't.  So, something is wrong --- assume it must be a cyclic
		// definition for now.
		err := r.srcmap.SyntaxError(incomplete, "cyclic declaration")
		return []SyntaxError{*err}
	}
	// Done
	return nil
}

// Check that a given set of source columns have been finalised.  This is
// important, since we cannot finalise a declaration until all of its
// dependencies have themselves been finalised.
func (r *resolver) declarationDependenciesAreFinalised(scope *ModuleScope,
	symbols util.Iterator[Symbol]) (bool, []SyntaxError) {
	var (
		errors    []SyntaxError
		finalised bool = true
	)
	//
	for iter := symbols; iter.HasNext(); {
		symbol := iter.Next()
		// Attempt to resolve
		if !symbol.IsResolved() && !scope.Bind(symbol) {
			errors = append(errors, *r.srcmap.SyntaxError(symbol, "unknown symbol"))
			// not finalised yet
			finalised = false
		} else if !symbol.Binding().IsFinalised() {
			// no, not finalised
			finalised = false
		}
	}
	//
	return finalised, errors
}

// Finalise a declaration.
func (r *resolver) finaliseDeclaration(scope *ModuleScope, decl Declaration) []SyntaxError {
	if d, ok := decl.(*DefConstraint); ok {
		return r.finaliseDefConstraintInModule(scope, d)
	} else if d, ok := decl.(*DefFun); ok {
		return r.finaliseDefFunInModule(scope, d)
	} else if d, ok := decl.(*DefInRange); ok {
		return r.finaliseDefInRangeInModule(scope, d)
	} else if d, ok := decl.(*DefInterleaved); ok {
		return r.finaliseDefInterleavedInModule(d)
	} else if d, ok := decl.(*DefLookup); ok {
		return r.finaliseDefLookupInModule(scope, d)
	} else if d, ok := decl.(*DefPermutation); ok {
		return r.finaliseDefPermutationInModule(d)
	} else if d, ok := decl.(*DefProperty); ok {
		return r.finaliseDefPropertyInModule(scope, d)
	}
	//
	return nil
}

// Finalise a vanishing constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefConstraintInModule(enclosing Scope, decl *DefConstraint) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Resolve guard
	if decl.Guard != nil {
		errors = r.finaliseExpressionInModule(scope, decl.Guard)
	}
	// Resolve constraint body
	errors = append(errors, r.finaliseExpressionInModule(scope, decl.Constraint)...)
	// Done
	return errors
}

// Finalise an interleaving assignment.  Since the assignment would already been
// initialised, all we need to do is determine the appropriate type and length
// multiplier for the interleaved column.  This can still result in an error,
// for example, if the multipliers between interleaved columns are incompatible,
// etc.
func (r *resolver) finaliseDefInterleavedInModule(decl *DefInterleaved) []SyntaxError {
	var (
		// Length multiplier being determined
		length_multiplier uint
		// Column type being determined
		datatype schema.Type
		// Errors discovered
		errors []SyntaxError
	)
	// Determine type and length multiplier
	for i, source := range decl.Sources {
		// Lookup binding of column being interleaved.
		binding := source.Binding().(*ColumnBinding)
		//
		if i == 0 {
			length_multiplier = binding.multiplier
			datatype = binding.dataType
		} else if binding.multiplier != length_multiplier {
			// Columns to be interleaved must have the same length multiplier.
			err := r.srcmap.SyntaxError(decl, fmt.Sprintf("source column %s has incompatible length multiplier", source.Name()))
			errors = append(errors, *err)
		}
		// Combine datatypes.
		datatype = schema.Join(datatype, binding.dataType)
	}
	// Finalise details only if no errors
	if len(errors) == 0 {
		// Determine actual length multiplier
		length_multiplier *= uint(len(decl.Sources))
		// Lookup existing declaration
		binding := decl.Target.Binding().(*ColumnBinding)
		// Update with completed information
		binding.multiplier = length_multiplier
		binding.dataType = datatype
	}
	// Done
	return errors
}

// Finalise a permutation assignment after all symbols have been resolved.  This
// requires checking the contexts of all columns is consistent.
func (r *resolver) finaliseDefPermutationInModule(decl *DefPermutation) []SyntaxError {
	var (
		multiplier uint = 0
		errors     []SyntaxError
	)
	// Finalise each column in turn
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i]
		// Lookup source of column being permuted
		source := ith.Binding().(*ColumnBinding)
		// Sanity check length multiplier
		if i == 0 && source.dataType.AsUint() == nil {
			errors = append(errors, *r.srcmap.SyntaxError(ith, "fixed-width type required"))
		} else if i == 0 {
			multiplier = source.multiplier
		} else if multiplier != source.multiplier {
			// Problem
			errors = append(errors, *r.srcmap.SyntaxError(ith, "incompatible length multiplier"))
		}
		// All good, finalise target column
		target := decl.Targets[i].Binding().(*ColumnBinding)
		// Update with completed information
		target.multiplier = source.multiplier
		target.dataType = source.dataType
	}
	// Done
	return errors
}

// Finalise a range constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefInRangeInModule(enclosing Scope, decl *DefInRange) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Resolve property body
	errors = append(errors, r.finaliseExpressionInModule(scope, decl.Expr)...)
	// Done
	return errors
}

// Finalise a function definition after all symbols have been resolved. This
// involves: (a) checking the context is valid for the body; (b) checking the
// body is well-typed; (c) for pure functions checking that no columns are
// accessed; (d) finally, resolving any parameters used within the body of this
// function.
func (r *resolver) finaliseDefFunInModule(enclosing Scope, decl *DefFun) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Declare parameters in local scope
	for _, p := range decl.Parameters() {
		scope.DeclareLocal(p.Name)
	}
	// Resolve property body
	errors = append(errors, r.finaliseExpressionInModule(scope, decl.Body())...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this lookup constraint.
func (r *resolver) finaliseDefLookupInModule(enclosing Scope, decl *DefLookup) []SyntaxError {
	var (
		errors      []SyntaxError
		sourceScope = NewLocalScope(enclosing, true)
		targetScope = NewLocalScope(enclosing, true)
	)
	// Resolve source expressions
	errors = append(errors, r.finaliseExpressionsInModule(sourceScope, decl.Sources)...)
	// Resolve target expressions
	errors = append(errors, r.finaliseExpressionsInModule(targetScope, decl.Targets)...)
	// Done
	return errors
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) finaliseDefPropertyInModule(enclosing Scope, decl *DefProperty) []SyntaxError {
	var (
		errors []SyntaxError
		scope  = NewLocalScope(enclosing, false)
	)
	// Resolve property body
	errors = append(errors, r.finaliseExpressionInModule(scope, decl.Assertion)...)
	// Done
	return errors
}

// Resolve a sequence of zero or more expressions within a given module.  This
// simply resolves each of the arguments in turn, collecting any errors arising.
func (r *resolver) finaliseExpressionsInModule(scope LocalScope, args []Expr) []SyntaxError {
	var errors []SyntaxError
	// Visit each argument
	for _, arg := range args {
		if arg != nil {
			errs := r.finaliseExpressionInModule(scope, arg)
			errors = append(errors, errs...)
		}
	}
	// Done
	return errors
}

// Resolve any variable accesses with this expression (which is declared in a
// given module).  The enclosing module is required to resolve unqualified
// variable accesses.  As above, the goal is ensure variable refers to something
// that was declared and, more specifically, what kind of access it is (e.g.
// column access, constant access, etc).
func (r *resolver) finaliseExpressionInModule(scope LocalScope, expr Expr) []SyntaxError {
	if _, ok := expr.(*Constant); ok {
		return nil
	} else if v, ok := expr.(*Add); ok {
		return r.finaliseExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*Exp); ok {
		return r.finaliseExpressionInModule(scope, v.Arg)
	} else if v, ok := expr.(*IfZero); ok {
		return r.finaliseExpressionsInModule(scope, []Expr{v.Condition, v.TrueBranch, v.FalseBranch})
	} else if v, ok := expr.(*Invoke); ok {
		return r.finaliseInvokeInModule(scope, v)
	} else if v, ok := expr.(*List); ok {
		return r.finaliseExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*Mul); ok {
		return r.finaliseExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*Normalise); ok {
		return r.finaliseExpressionInModule(scope, v.Arg)
	} else if v, ok := expr.(*Sub); ok {
		return r.finaliseExpressionsInModule(scope, v.Args)
	} else if v, ok := expr.(*VariableAccess); ok {
		return r.finaliseVariableInModule(scope, v)
	} else {
		return r.srcmap.SyntaxErrors(expr, "unknown expression")
	}
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseInvokeInModule(scope LocalScope, expr *Invoke) []SyntaxError {
	// Resolve arguments
	if errors := r.finaliseExpressionsInModule(scope, expr.Args()); errors != nil {
		return errors
	}
	// Lookup the corresponding function definition.
	if !scope.Bind(expr) {
		return r.srcmap.SyntaxErrors(expr, "unknown function")
	}
	// Success
	return nil
}

// Resolve a specific variable access contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseVariableInModule(scope LocalScope,
	expr *VariableAccess) []SyntaxError {
	// Check whether this is a qualified access, or not.
	if !scope.IsGlobal() && expr.IsQualified() {
		return r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	} else if expr.IsQualified() && !scope.HasModule(expr.Module()) {
		return r.srcmap.SyntaxErrors(expr, fmt.Sprintf("unknown module %s", expr.Module()))
	}
	// Symbol should be resolved at this point, but we still need to check the
	// context.
	if expr.IsResolved() {
		// Update context
		binding, ok := expr.Binding().(*ColumnBinding)
		if ok && !scope.FixContext(binding.Context()) {
			return r.srcmap.SyntaxErrors(expr, "conflicting context")
		} else if !ok {
			// Unable to resolve variable
			return r.srcmap.SyntaxErrors(expr, "not a column")
		}
		// Done
		return nil
	} else if scope.Bind(expr) {
		// Must be a local variable or parameter access, so we're all good.
		return nil
	}
	// Unable to resolve variable
	return r.srcmap.SyntaxErrors(expr, "unresolved symbol")
}