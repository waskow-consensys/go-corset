package assignment

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
)

// DataColumn represents a column of user-provided values.
type DataColumn struct {
	// Module where this data column is located.
	module uint
	// Name of this datacolumn
	name string
	// Expected type of values held in this column.  Observe that this should be
	// true for the input columns for any valid trace and, furthermore, every
	// computed column should have values of this type.
	datatype schema.Type
}

// NewDataColumn constructs a new data column with a given name.
func NewDataColumn(module uint, name string, base schema.Type) *DataColumn {
	return &DataColumn{module, name, base}
}

// Module identifies the module which encloses this column.
func (p *DataColumn) Module() uint {
	return p.module
}

// Name provides access to information about the ith column in a schema.
func (p *DataColumn) Name() string {
	return p.name
}

// Type Returns the expected type of data in this column
func (p *DataColumn) Type() schema.Type {
	return p.datatype
}

//nolint:revive
func (c *DataColumn) String() string {
	if c.datatype.AsField() != nil {
		return fmt.Sprintf("(column %s)", c.Name())
	}

	return fmt.Sprintf("(column %s :%s)", c.Name(), c.datatype)
}

// ============================================================================
// Declaration Interface
// ============================================================================

// Columns returns the columns declared by this computed column.
func (p *DataColumn) Columns() util.Iterator[schema.Column] {
	// Datacolumns always have a multiplier of 1.
	column := schema.NewColumn(p.module, p.name, 1, p.datatype)
	return util.NewUnitIterator[schema.Column](column)
}

// IsComputed Determines whether or not this declaration is computed (which data
// columns never are).
func (p *DataColumn) IsComputed() bool {
	return false
}