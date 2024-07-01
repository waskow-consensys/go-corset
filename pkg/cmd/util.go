package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/spf13/cobra"
)

// Get an expected flag, or panic if an error arises.
func getFlag(cmd *cobra.Command, flag string) bool {
	r, err := cmd.Flags().GetBool(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return r
}

// Get an expectedsigned integer, or panic if an error arises.
func getInt(cmd *cobra.Command, flag string) int {
	r, err := cmd.Flags().GetInt(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return r
}

// Get an expected unsigned integer, or panic if an error arises.
func getUint(cmd *cobra.Command, flag string) uint {
	r, err := cmd.Flags().GetUint(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// Get an expected string, or panic if an error arises.
func getString(cmd *cobra.Command, flag string) string {
	r, err := cmd.Flags().GetString(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// Write a given trace file to disk
func writeTraceFile(filename string, tr trace.Trace) {
	var err error

	var bytes []byte
	// Check file extension
	ext := path.Ext(filename)
	//
	switch ext {
	case ".json":
		js := json.ToJsonString(tr)
		//
		if err = os.WriteFile(filename, []byte(js), 0644); err == nil {
			return
		}
	case ".lt":
		bytes, err = lt.ToBytes(tr)
		//
		if err == nil {
			if err = os.WriteFile(filename, bytes, 0644); err == nil {
				return
			}
		}
	default:
		err = fmt.Errorf("Unknown trace file format: %s", ext)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
}

// Parse a trace file using a parser based on the extension of the filename.
func readTraceFile(filename string) trace.Trace {
	var tr trace.Trace
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			tr, err = json.FromBytes(bytes)
			if err == nil {
				return tr
			}
		case ".lt":
			tr, err = lt.FromBytes(bytes)
			if err == nil {
				return tr
			}
		default:
			err = fmt.Errorf("Unknown trace file format: %s", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Parse a constraints schema file using a parser based on the extension of the
// filename.
func readSchemaFile(filename string) *hir.Schema {
	var schema *hir.Schema
	// Read schema file
	bytes, err := os.ReadFile(filename)
	// Handle errors
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".lisp":
			// Parse bytes into an S-Expression
			schema, err = hir.ParseSchemaString(string(bytes))
			if err == nil {
				return schema
			}
		case ".bin":
			schema, err = binfile.HirSchemaFromJson(bytes)
			if err == nil {
				return schema
			}
		default:
			err = fmt.Errorf("Unknown schema file format: %s\n", ext)
		}
	}
	// Handle error
	if e, ok := err.(*sexp.SyntaxError); ok {
		printSyntaxError(filename, e, string(bytes))
	} else {
		fmt.Println(err)
	}

	os.Exit(2)
	// unreachable
	return nil
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(filename string, err *sexp.SyntaxError, text string) {
	span := err.Span()
	// Construct empty source map in order to determine enclosing line.
	srcmap := sexp.NewSourceMap[sexp.SExp]([]rune(text))
	//
	line := srcmap.FindFirstEnclosingLine(span)
	// Print error + line number
	fmt.Printf("%s:%d: %s\n", filename, line.Number(), err.Message())
	// Print separator line
	fmt.Println()
	// Print line
	fmt.Println(line.String())
	// Print indent (todo: account for tabs)
	lineOffset := span.Start() - line.Start()
	fmt.Print(strings.Repeat(" ", lineOffset))
	// Calculate length (ensures don't overflow line)
	length := min(line.Length()-lineOffset, span.Length())
	// Print highlight
	fmt.Println(strings.Repeat("^", length))
}

// QualifiedColumnName returns a fully qualified column name based on its column
// index.
func QualifiedColumnName(cid uint, tr trace.Trace) string {
	col := tr.Columns().Get(cid)
	mod := tr.Modules().Get(col.Module())
	// Check whether qualification required
	if mod.Name() != "" {
		return fmt.Sprintf("%s.%s", mod.Name(), col.Name())
	}
	// Prelude module
	return col.Name()
}
