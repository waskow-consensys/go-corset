package table

// Schema represents a schema which can be used to manipulate a trace.
// Specifically, a schema can determine whether or not a trace is accepted;
// likewise, a schema can expand a trace according to its internal computation.
type Schema interface {
	Accepts(Trace) error
	// ExpandTrace expands a given trace to include "computed
	// columns".  These are columns which do not exist in the
	// original trace, but are added during trace expansion to
	// form the final trace.
	ExpandTrace(Trace) error

	// Size returns the number of declarations in this schema.
	Size() int

	// GetDeclaration returns the ith declaration in this schema.
	GetDeclaration(int) Declaration
}

// Declaration represents a declared element of a schema.  For example, a column
// declaration or a vanishing constraint declaration.  The purpose of this
// interface is to provide some generic interactions that are available
// regardless of the IR level.
type Declaration interface {
	// Return a human-readable string for this declaration.
	String() string
}