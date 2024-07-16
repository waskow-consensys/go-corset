package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/spf13/cobra"
)

// computeCmd represents the compute command
var checkCmd = &cobra.Command{
	Use:   "check [flags] trace_file constraint_file",
	Short: "Check a given trace against a set of constraints.",
	Long: `Check a given trace against a set of constraints.
	Traces can be given either as JSON or binary lt files.
	Constraints can be given either as lisp or bin files.`,
	Run: func(cmd *cobra.Command, args []string) {
		var hirSchema *hir.Schema
		var cfg checkConfig

		if len(args) != 2 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		cfg.air = getFlag(cmd, "air")
		cfg.mir = getFlag(cmd, "mir")
		cfg.hir = getFlag(cmd, "hir")
		cfg.expand = !getFlag(cmd, "raw")
		cfg.report = getFlag(cmd, "report")
		cfg.spillage = getInt(cmd, "spillage")
		cfg.strict = !getFlag(cmd, "warn")
		cfg.quiet = getFlag(cmd, "quiet")
		cfg.padding.Right = getUint(cmd, "padding")
		// TODO: support true ranges
		cfg.padding.Left = cfg.padding.Right
		// Parse trace
		trace := readTraceFile(args[0])
		// Parse constraints
		hirSchema = readSchemaFile(args[1])
		// Go!
		checkTraceWithLowering(trace, hirSchema, cfg)
	},
}

// check config encapsulates certain parameters to be used when
// checking traces.
type checkConfig struct {
	// Performing checking at HIR level
	hir bool
	// Performing checking at MIR level
	mir bool
	// Performing checking at AIR level
	air bool
	// Determines how much spillage to account for.  This gives the user the
	// ability to override the inferred default.  A negative value indicates
	// this default should be used.
	spillage int
	// Determines how much padding to use
	padding util.Pair[uint, uint]
	// Suppress output (e.g. warnings)
	quiet bool
	// Specified whether strict checking is performed or not.  This is enabled
	// by default, and ensures the tool fails with an error in any unexpected or
	// unusual case.
	strict bool
	// Specifies whether or not to perform trace expansion.  Trace expansion is
	// not required when a "raw" trace is given which already includes all
	// implied columns.
	expand bool
	// Specifies whether or not to report details of the failure (e.g. for
	// debugging purposes).
	report bool
}

// Check a given trace is consistently accepted (or rejected) at the different
// IR levels.
func checkTraceWithLowering(tr trace.Trace, schema *hir.Schema, cfg checkConfig) {
	if !cfg.hir && !cfg.mir && !cfg.air {
		// Process together
		checkTraceWithLoweringDefault(tr, schema, cfg)
	} else {
		// Process individually
		if cfg.hir {
			checkTraceWithLoweringHir(tr, schema, cfg)
		}

		if cfg.mir {
			checkTraceWithLoweringMir(tr, schema, cfg)
		}

		if cfg.air {
			checkTraceWithLoweringAir(tr, schema, cfg)
		}
	}
}

func checkTraceWithLoweringHir(tr trace.Trace, hirSchema *hir.Schema, cfg checkConfig) {
	trHIR, errHIR := checkTrace(tr, hirSchema, cfg)
	//
	if errHIR != nil {
		reportError("HIR", trHIR, errHIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringMir(tr trace.Trace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Check trace
	trMIR, errMIR := checkTrace(tr, mirSchema, cfg)
	//
	if errMIR != nil {
		reportError("MIR", trMIR, errMIR, cfg)
		os.Exit(1)
	}
}

func checkTraceWithLoweringAir(tr trace.Trace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	trAIR, errAIR := checkTrace(tr, airSchema, cfg)
	//
	if errAIR != nil {
		reportError("AIR", trAIR, errAIR, cfg)
		os.Exit(1)
	}
}

// The default check allows one to compare all levels against each other and
// look for any discrepenacies.
func checkTraceWithLoweringDefault(tr trace.Trace, hirSchema *hir.Schema, cfg checkConfig) {
	// Lower HIR => MIR
	mirSchema := hirSchema.LowerToMir()
	// Lower MIR => AIR
	airSchema := mirSchema.LowerToAir()
	//
	trHIR, errHIR := checkTrace(tr, hirSchema, cfg)
	trMIR, errMIR := checkTrace(tr, mirSchema, cfg)
	trAIR, errAIR := checkTrace(tr, airSchema, cfg)
	//
	if errHIR != nil || errMIR != nil || errAIR != nil {
		strHIR := toErrorString(errHIR)
		strMIR := toErrorString(errMIR)
		strAIR := toErrorString(errAIR)
		// At least one error encountered.
		if strHIR == strMIR && strMIR == strAIR {
			fmt.Println(errHIR)
		} else {
			reportError("HIR", trHIR, errHIR, cfg)
			reportError("MIR", trMIR, errMIR, cfg)
			reportError("AIR", trAIR, errAIR, cfg)
		}

		os.Exit(1)
	}
}

func checkTrace(tr trace.Trace, schema sc.Schema, cfg checkConfig) (trace.Trace, error) {
	//
	if cfg.expand {
		// Clone to prevent interefence with subsequent checks
		tr = tr.Clone()
		// Apply spillage
		if cfg.spillage >= 0 {
			// Apply user-specified spillage
			trace.PadColumns(tr, uint(cfg.spillage))
		} else {
			// Apply default inferred spillage
			trace.PadColumns(tr, sc.RequiredSpillage(schema))
		}
		// Perform Input Alignment
		if err := performAlignment(true, tr, schema, cfg); err != nil {
			return tr, err
		}
		// Validate trace
		if err := validationCheck(tr, schema); err != nil {
			return tr, err
		}
		// Expand trace
		if err := sc.ExpandTrace(schema, tr); err != nil {
			return tr, err
		}
	}
	// Perform Alignment
	if err := performAlignment(false, tr, schema, cfg); err != nil {
		return tr, err
	}
	// Apply padding (as necessary)
	for n := cfg.padding.Left; n <= cfg.padding.Right; n++ {
		if ptr, err := padAndCheckTrace(n, tr, schema); err != nil {
			return ptr, err
		}
	}
	// Done
	return nil, nil
}

func padAndCheckTrace(n uint, tr trace.Trace, schema sc.Schema) (trace.Trace, error) {
	var ptr trace.Trace = tr
	// Apply padding (if applicable)
	if n != 0 {
		// Prevent interference
		ptr := tr.Clone()
		// Apply padding
		trace.PadColumns(ptr, n)
	}
	// Check whether accepted or not.
	err := sc.Accepts(schema, ptr)
	// Done
	return ptr, err
}

// Run the alignment algorithm with optional checks determined by the configuration.
func performAlignment(inputs bool, tr trace.Trace, schema sc.Schema, cfg checkConfig) error {
	var err error

	var nSchemaCols uint
	// Determine number of trace columns
	nTraceCols := tr.Columns().Len()

	if inputs {
		nSchemaCols = schema.InputColumns().Count()
		err = sc.AlignInputs(tr, schema)
	} else {
		nSchemaCols = schema.Columns().Count()
		err = sc.Align(tr, schema)
	}
	// Sanity check error
	if err != nil {
		return err
	} else if cfg.strict && nSchemaCols != nTraceCols {
		col := tr.Columns().Get(nSchemaCols)
		mod := tr.Modules().Get(col.Context().Module())
		// Return error
		return fmt.Errorf("unknown trace column %s", sc.QualifiedColumnName(mod.Name(), col.Name()))
	} else if nSchemaCols != nTraceCols {
		// Log warning
		if !cfg.quiet {
			for i := nSchemaCols; i < nTraceCols; i++ {
				col := tr.Columns().Get(i)
				mod := tr.Modules().Get(col.Context().Module())
				fmt.Printf("[WARNING] unknown trace column %s\n", sc.QualifiedColumnName(mod.Name(), col.Name()))
			}
		}
		// Finally, remove the unknown columns.  This is important as,
		// otherwise, the column indices they occupy will clash with computed
		// columns which are added during lowering.
		tr.Columns().Trim(nSchemaCols)
	}

	return nil
}

// Validate that values held in trace columns match the expected type.  This is
// really a sanity check that the trace is not malformed.
func validationCheck(tr trace.Trace, schema sc.Schema) error {
	nTraceCols := tr.Columns().Len()
	schemaCols := schema.Columns()
	// Check each column in turn
	for i := uint(0); i < nTraceCols; i++ {
		// Extract ith column
		col := tr.Columns().Get(i)
		// Extract schema for ith column
		scCol := schemaCols.Next()
		// Determine enclosing module
		mod := schema.Modules().Nth(scCol.Context().Module())
		// Extract type for ith column
		colType := scCol.Type()
		// Check elements
		for j := 0; j < int(col.Height()); j++ {
			jth := col.Get(j)
			if !colType.Accept(jth) {
				qualColName := sc.QualifiedColumnName(mod.Name(), col.Name())
				return fmt.Errorf("row %d of column %s is out-of-bounds (%s)", j, qualColName, jth)
			}
		}
	}
	// Done
	return nil
}

func toErrorString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func reportError(ir string, tr trace.Trace, err error, cfg checkConfig) {
	if cfg.report && tr != nil {
		trace.PrintTrace(tr)
	}

	if err != nil {
		fmt.Printf("%s: %s\n", ir, err)
	} else {
		fmt.Printf("Trace should have been rejected at %s level.\n", ir)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("report", false, "report details of failure for debugging")
	checkCmd.Flags().Bool("raw", false, "assume input trace already expanded")
	checkCmd.Flags().Bool("hir", false, "check at HIR level")
	checkCmd.Flags().Bool("mir", false, "check at MIR level")
	checkCmd.Flags().Bool("air", false, "check at AIR level")
	checkCmd.Flags().BoolP("warn", "w", false, "report warnings instead of failing for certain errors"+
		"(e.g. unknown columns in the trace)")
	checkCmd.Flags().BoolP("quiet", "q", false, "suppress output (e.g. warnings)")
	checkCmd.Flags().Uint("padding", 0, "specify amount of (front) padding to apply")
	checkCmd.Flags().Int("spillage", -1,
		"specify amount of splillage to account for (where -1 indicates this should be inferred)")
}
