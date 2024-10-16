package schema

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	log "github.com/sirupsen/logrus"
)

// ============================================================================
// TraceEnumerator
// ============================================================================

// TraceEnumerator is an adaptor which surrounds an enumerator and, essentially,
// converts flat sequences of elements into traces.
type TraceEnumerator struct {
	// Schema for which traces are being generated
	schema Schema
	// Number of lines
	lines uint
	// Enumerate sequences of elements
	enumerator util.Enumerator[[]fr.Element]
}

// NewTraceEnumerator constructs an enumerator for all traces matching the
// given column specifications using elements sourced from the given pool.
func NewTraceEnumerator(lines uint, schema Schema, pool []fr.Element) util.Enumerator[tr.Trace] {
	ncells := schema.InputColumns().Count() * lines
	// Construct the enumerator
	enumerator := util.EnumerateElements[fr.Element](ncells, pool)
	// Done
	return &TraceEnumerator{schema, lines, enumerator}
}

// Next returns the next trace in the enumeration
func (p *TraceEnumerator) Next() tr.Trace {
	ncols := p.schema.InputColumns().Count()
	elems := p.enumerator.Next()
	cols := make([]tr.RawColumn, ncols)
	//
	i, j := 0, 0
	// Construct each column from the sequence
	for iter := p.schema.InputColumns(); iter.HasNext(); {
		col := iter.Next()
		data := util.NewFrArray(p.lines, 256)
		// Slice nrows values from elems
		for k := uint(0); k < p.lines; k++ {
			data.Set(k, elems[j])
			// Consume element from generated sequence
			j++
		}
		// Construct raw column
		modName := p.schema.Modules().Nth(col.context.Module()).name
		cols[i] = tr.RawColumn{Module: modName, Name: col.Name(), Data: data}
		i++
	}
	// Finally, build the trace.
	builder := NewTraceBuilder(p.schema).Expand(true).Parallel(false).Padding(0)
	// Build the trace
	trace, errs := builder.Build(cols)
	// Handle errors
	if errs != nil {
		// Should be unreachable, since control the trace!
		for _, err := range errs {
			log.Error(err)
		}
		// Fail
		panic("invalid trace constructed")
	}
	// Done
	return trace
}

// HasNext checks whether the enumeration has more elements (or not).
func (p *TraceEnumerator) HasNext() bool {
	return p.enumerator.HasNext()
}