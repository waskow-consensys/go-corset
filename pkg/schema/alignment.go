package schema

import (
	"fmt"

	tr "github.com/consensys/go-corset/pkg/trace"
)

// AlignInputs attempts to align this trace with the input columns of a given
// schema.  This means ensuring the order of columns in this trace matches the
// order of input columns in the schema. Thus, column indexes used by
// constraints in the schema can directly access in this trace (i.e. without
// name lookup). Alignment can fail, however, if there is a mismatch between
// columns in the trace and those expected by the schema.
func AlignInputs(p tr.Trace, schema Schema) error {
	return alignWith(false, p, schema)
}

// Align attempts to align this trace with a given schema.  This means ensuring
// the order of columns in this trace matches the order in the schema. Thus,
// column indexes used by constraints in the schema can directly access in this
// trace (i.e. without name lookup).  Alignment can fail, however, if there is a
// mismatch between columns in the trace and those expected by the schema.
func Align(p tr.Trace, schema Schema) error {
	return alignWith(true, p, schema)
}

// Alignment algorithm which operates either in unexpanded or expanded mode.  In
// expanded mode, all columns must be accounted for and will be aligned.  In
// unexpanded mode, the trace is only expected to contain input (i.e.
// non-computed) columns.  Furthermore, in the schema these are expected to be
// allocated before computed columns.  As such, alignment of these input
// columns is performed.
func alignWith(expand bool, p tr.Trace, schema Schema) error {
	columns := p.Columns()
	modules := p.Modules()
	ncols := p.Columns().Len()
	modIndex := uint(0)
	// Check alignment of modules
	for i := schema.Modules(); i.HasNext(); {
		schemaMod := i.Next()
		traceMod := p.Modules().Get(modIndex)

		if schemaMod.Name() != traceMod.Name() {
			// Not aligned --- so fix
			k, ok := p.Modules().IndexOf(schemaMod.Name())
			// Check module exists
			if !ok {
				// This situation can occur when a module is declared in the
				// schema, but which has no column declarations (hence, by
				// definition, it would be missing from the trace).  Commonly,
				// this happens when no columns are declared in the prelude,
				// because schema's constructed by the builder always have a
				// prelude module.  In such a situation, its reasonable to
				// create an empty module as this is of no real consequence.
				k = p.Modules().Add(schemaMod.Name(), 0)
			} else if k < modIndex {
				// Sanity check
				panic("internal failure")
			}
			// Swap modules
			p.Modules().Swap(modIndex, k)
		}

		modIndex++
	}
	//
	colIndex := uint(0)
	// Check alignment of columns.  Observe that we don't currently care whether
	// modules are aligned.  That is because modules don't really serve any
	// significant purpose.  However, this might change at some point.
	for i := schema.Declarations(); i.HasNext(); {
		ith := i.Next()
		if expand || !ith.IsComputed() {
			for j := ith.Columns(); j.HasNext(); {
				// Extract schema column & module
				schemaCol := j.Next()
				schemaMod := schema.Modules().Nth(schemaCol.Module())
				// Sanity check column exists
				if colIndex >= ncols {
					return fmt.Errorf("trace missing column %s.%s (too few columns)", schemaMod.Name(), schemaCol.Name())
				}
				// Extract trace column and module
				traceCol := columns.Get(colIndex)
				traceMod := modules.Get(traceCol.Module())
				// Check alignment
				if traceCol.Name() != schemaCol.Name() || traceMod.Name() != schemaMod.Name() {
					// Not aligned --- so fix
					k, ok := p.Columns().IndexOf(schemaCol.Module(), schemaCol.Name())
					// check exists
					if !ok {
						return fmt.Errorf("trace missing column %s.%s", schemaMod.Name(), schemaCol.Name())
					}
					// Swap columns
					columns.Swap(colIndex, k)
				}
				// Continue
				colIndex++
			}
		}
	}
	// Check whether all columns matched
	if colIndex == ncols {
		// Yes, alignment complete.
		return nil
	}
	// Error Case.
	n := ncols - colIndex
	unknowns := make([]string, n)
	// Determine names of unknown columns.
	for i := colIndex; i < ncols; i++ {
		unknowns[i-colIndex] = columns.Get(i).Name()
	}
	//
	return fmt.Errorf("trace contains unknown columns: %v", unknowns)
}