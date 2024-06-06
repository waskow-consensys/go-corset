package hir

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/table"
)

// ===================================================================
// Public
// ===================================================================

// ParseSExp parses a string representing an HIR expression formatted using
// S-expressions.
func ParseSExp(s string) (Expr, error) {
	p := newExprTranslator()
	// Parse string
	return p.ParseAndTranslate(s)
}

// ParseSchemaString parses a sequence of zero or more HIR schema declarations
// represented as a string.  Internally, this uses sexp.ParseAll and
// ParseSchemaSExp to do the work.
func ParseSchemaString(str string) (*Schema, error) {
	// Parse bytes into an S-Expression
	terms, err := sexp.ParseAll(str)
	// Check test file parsed ok
	if err != nil {
		return nil, err
	}
	// Parse terms into an HIR schema
	return ParseSchemaSExp(terms)
}

// ParseSchemaSExp parses a sequence of zero or more HIR schema declarations
// represented as S-expressions.
func ParseSchemaSExp(terms []sexp.SExp) (*Schema, error) {
	t := newExprTranslator()
	// Construct initially empty schema
	schema := EmptySchema()
	// Continue parsing string until nothing remains.
	for _, term := range terms {
		// Process declaration
		err2 := sexpDeclaration(term, schema, t)
		if err2 != nil {
			return nil, err2
		}
	}
	// Done
	return schema, nil
}

// ===================================================================
// Private
// ===================================================================

func newExprTranslator() *sexp.Translator[Expr] {
	p := sexp.NewTranslator[Expr]()
	// Configure translator
	p.AddSymbolRule(sexpConstant)
	p.AddSymbolRule(sexpColumnAccess)
	p.AddBinaryRule("shift", sexpShift)
	p.AddRecursiveRule("+", sexpAdd)
	p.AddRecursiveRule("-", sexpSub)
	p.AddRecursiveRule("*", sexpMul)
	p.AddRecursiveRule("~", sexpNorm)
	p.AddRecursiveRule("if", sexpIf)
	p.AddRecursiveRule("ifnot", sexpIfNot)
	p.AddRecursiveRule("begin", sexpBegin)

	return p
}

func sexpDeclaration(s sexp.SExp, schema *Schema, p *sexp.Translator[Expr]) error {
	if e, ok := s.(*sexp.List); ok {
		if e.Len() >= 2 && e.Len() <= 3 && e.MatchSymbols(2, "column") {
			return sexpColumn(e.Elements, schema)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish") {
			return sexpVanishing(e.Elements, nil, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish:last") {
			domain := -1
			return sexpVanishing(e.Elements, &domain, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(2, "vanish:first") {
			domain := 0
			return sexpVanishing(e.Elements, &domain, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(2, "assert") {
			return sexpAssertion(e.Elements, schema, p)
		} else if e.Len() == 3 && e.MatchSymbols(1, "permute") {
			return sexpPermutation(e.Elements, schema)
		}
	}

	return fmt.Errorf("unexpected declaration: %s", s)
}

// Parse a column declaration
func sexpColumn(elements []sexp.SExp, schema *Schema) error {
	columnName := elements[1].String()

	var columnType table.Type = &table.FieldType{}

	if len(elements) == 3 {
		var err error
		columnType, err = sexpType(elements[2].String())

		if err != nil {
			return err
		}
	}

	schema.AddDataColumn(columnName, columnType)

	return nil
}

// Parse a permutation declaration
func sexpPermutation(elements []sexp.SExp, schema *Schema) error {
	// Target columns are (sorted) permutations of source columns.
	sexpTargets := elements[1].AsList()
	// Source columns.
	sexpSources := elements[2].AsList()
	// Convert into appropriate form.
	targets := make([]string, sexpTargets.Len())
	sources := make([]string, sexpSources.Len())
	signs := make([]bool, sexpSources.Len())
	//
	for i := 0; i < sexpTargets.Len(); i++ {
		target := sexpTargets.Get(i).AsSymbol()
		// Sanity check syntax as expected
		if target == nil {
			return fmt.Errorf("expected column name, found: %s", elements[i])
		}
		// Copy over
		targets[i] = target.String()
	}
	//
	for i := 0; i < sexpSources.Len(); i++ {
		source := sexpSources.Get(i).AsSymbol()
		// Sanity check syntax as expected
		if source == nil {
			return fmt.Errorf("expected column name, found: %s", elements[i])
		}
		// Determine source column sign (i.e. sort direction)
		sortName := source.String()
		if strings.HasPrefix(sortName, "+") {
			signs[i] = true
		} else if strings.HasPrefix(sortName, "-") {
			signs[i] = false
		} else {
			return fmt.Errorf("sort direction (+/-) required, found: %s", sortName)
		}
		// Copy over column name
		sources[i] = sortName[1:]
	}
	//
	schema.AddPermutationColumns(targets, signs, sources)
	//
	return nil
}

// Parse a property assertion
func sexpAssertion(elements []sexp.SExp, schema *Schema, p *sexp.Translator[Expr]) error {
	handle := elements[1].String()

	expr, err := p.Translate(elements[2])
	if err != nil {
		return err
	}
	// Add all assertions arising.
	for _, e := range expr.LowerTo() {
		schema.AddPropertyAssertion(handle, e)
	}

	return nil
}

// Parse a vanishing declaration
func sexpVanishing(elements []sexp.SExp, domain *int, schema *Schema, p *sexp.Translator[Expr]) error {
	handle := elements[1].String()

	expr, err := p.Translate(elements[2])
	if err != nil {
		return err
	}

	schema.AddVanishingConstraint(handle, domain, expr)

	return nil
}

func sexpType(symbol string) (table.Type, error) {
	if strings.HasPrefix(symbol, ":u") {
		n, err := strconv.Atoi(symbol[2:])
		if err != nil {
			return nil, err
		}
		// FIXME: check for prove
		return table.NewUintType(uint(n), true), nil
	}

	return nil, fmt.Errorf("unexpected type: %s", symbol)
}

func sexpBegin(args []Expr) (Expr, error) {
	return &List{args}, nil
}

func sexpConstant(symbol string) (Expr, error) {
	num := new(fr.Element)
	// Attempt to parse
	c, err := num.SetString(symbol)
	// Check for errors
	if err != nil {
		return nil, err
	}
	// Done
	return &Constant{Val: c}, nil
}

func sexpColumnAccess(col string) (Expr, error) {
	return &ColumnAccess{col, 0}, nil
}

func sexpAdd(args []Expr) (Expr, error) {
	return &Add{args}, nil
}

func sexpSub(args []Expr) (Expr, error) {
	return &Sub{args}, nil
}

func sexpMul(args []Expr) (Expr, error) {
	return &Mul{args}, nil
}

func sexpIf(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], args[1], nil}, nil
	} else if len(args) == 3 {
		return &IfZero{args[0], args[1], args[2]}, nil
	}

	return nil, fmt.Errorf("incorrect number of arguments: {%d}", len(args))
}

func sexpIfNot(args []Expr) (Expr, error) {
	if len(args) == 2 {
		return &IfZero{args[0], nil, args[1]}, nil
	}

	return nil, fmt.Errorf("incorrect number of arguments: {%d}", len(args))
}

func sexpShift(col string, amt string) (Expr, error) {
	n, err := strconv.Atoi(amt)

	if err != nil {
		return nil, err
	}

	return &ColumnAccess{
		Column: col,
		Shift:  n,
	}, nil
}

func sexpNorm(args []Expr) (Expr, error) {
	if len(args) != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: {%d}", len(args))
		return nil, errors.New(msg)
	}

	return &Normalise{Arg: args[0]}, nil
}
