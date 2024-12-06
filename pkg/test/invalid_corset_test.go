package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/sexp"
)

// Determines the (relative) location of the test directory.  That is
// where the corset test files (lisp) and the corresponding traces
// (accepts/rejects) are found.
const InvalidTestDir = "../../testdata"

// ===================================================================
// Basic Tests
// ===================================================================

func Test_Invalid_Basic_01(t *testing.T) {
	CheckInvalid(t, "basic_invalid_01")
}

func Test_Invalid_Basic_02(t *testing.T) {
	CheckInvalid(t, "basic_invalid_02")
}

func Test_Invalid_Basic_03(t *testing.T) {
	CheckInvalid(t, "basic_invalid_03")
}

func Test_Invalid_Basic_04(t *testing.T) {
	CheckInvalid(t, "basic_invalid_04")
}

func Test_Invalid_Basic_05(t *testing.T) {
	CheckInvalid(t, "basic_invalid_05")
}

func Test_Invalid_Basic_06(t *testing.T) {
	CheckInvalid(t, "basic_invalid_06")
}

func Test_Invalid_Basic_07(t *testing.T) {
	CheckInvalid(t, "basic_invalid_07")
}

func Test_Invalid_Basic_08(t *testing.T) {
	CheckInvalid(t, "basic_invalid_08")
}

func Test_Invalid_Basic_09(t *testing.T) {
	CheckInvalid(t, "basic_invalid_09")
}

func Test_Invalid_Basic_10(t *testing.T) {
	CheckInvalid(t, "basic_invalid_10")
}

func Test_Invalid_Basic_11(t *testing.T) {
	CheckInvalid(t, "basic_invalid_11")
}

func Test_Invalid_Basic_12(t *testing.T) {
	CheckInvalid(t, "basic_invalid_12")
}

// ===================================================================
// Property Tests
// ===================================================================
func Test_Invalid_Property_01(t *testing.T) {
	CheckInvalid(t, "property_invalid_01")
}

func Test_Invalid_Property_02(t *testing.T) {
	CheckInvalid(t, "property_invalid_02")
}

// ===================================================================
// Shift Tests
// ===================================================================

func Test_Invalid_Shift_01(t *testing.T) {
	CheckInvalid(t, "shift_invalid_01")
}

func Test_Invalid_Shift_02(t *testing.T) {
	CheckInvalid(t, "shift_invalid_02")
}

// ===================================================================
// Normalisation Tests
// ===================================================================

func Test_Invalid_Norm_01(t *testing.T) {
	CheckInvalid(t, "norm_invalid_01")
}

// ===================================================================
// If-Zero
// ===================================================================

func Test_Invalid_If_01(t *testing.T) {
	CheckInvalid(t, "if_invalid_01")
}

func Test_Invalid_If_02(t *testing.T) {
	CheckInvalid(t, "if_invalid_02")
}

// ===================================================================
// Range Constraints
// ===================================================================

func Test_Invalid_Range_01(t *testing.T) {
	CheckInvalid(t, "range_invalid_01")
}

func Test_Invalid_Range_02(t *testing.T) {
	CheckInvalid(t, "range_invalid_02")
}

func Test_Invalid_Range_03(t *testing.T) {
	CheckInvalid(t, "range_invalid_03")
}

func Test_Invalid_Range_04(t *testing.T) {
	CheckInvalid(t, "range_invalid_04")
}

// ===================================================================
// Modules
// ===================================================================

func Test_Invalid_Module_01(t *testing.T) {
	CheckInvalid(t, "module_invalid_01")
}

// ===================================================================
// Permutations
// ===================================================================

func Test_Invalid_Permute_01(t *testing.T) {
	CheckInvalid(t, "permute_invalid_01")
}

func Test_Invalid_Permute_02(t *testing.T) {
	CheckInvalid(t, "permute_invalid_02")
}

func Test_Invalid_Permute_03(t *testing.T) {
	CheckInvalid(t, "permute_invalid_03")
}

func Test_Invalid_Permute_04(t *testing.T) {
	CheckInvalid(t, "permute_invalid_04")
}

func Test_Invalid_Permute_05(t *testing.T) {
	CheckInvalid(t, "permute_invalid_05")
}

func Test_Invalid_Permute_06(t *testing.T) {
	CheckInvalid(t, "permute_invalid_06")
}
func Test_Invalid_Permute_07(t *testing.T) {
	CheckInvalid(t, "permute_invalid_07")
}

// ===================================================================
// Lookups
// ===================================================================

func Test_Invalid_Lookup_01(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_01")
}

func Test_Invalid_Lookup_02(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_02")
}
func Test_Invalid_Lookup_03(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_03")
}

func Test_Invalid_Lookup_04(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_04")
}

func Test_Invalid_Lookup_05(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_05")
}
func Test_Invalid_Lookup_06(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_06")
}
func Test_Invalid_Lookup_07(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_07")
}
func Test_Invalid_Lookup_08(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_08")
}
func Test_Invalid_Lookup_09(t *testing.T) {
	CheckInvalid(t, "lookup_invalid_09")
}

// ===================================================================
// Interleavings
// ===================================================================

func Test_Invalid_Interleave_01(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_01")
}

func Test_Invalid_Interleave_02(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_02")
}

func Test_Invalid_Interleave_03(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_03")
}

func Test_Invalid_Interleave_04(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_04")
}

func Test_Invalid_Interleave_05(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_05")
}

func Test_Invalid_Interleave_06(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_06")
}

func Test_Invalid_Interleave_07(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_07")
}

func Test_Invalid_Interleave_08(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_08")
}

func Test_Invalid_Interleave_09(t *testing.T) {
	CheckInvalid(t, "interleave_invalid_09")
}

// ===================================================================
// Functions
// ===================================================================

func Test_Invalid_PureFun_01(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_01")
}

func Test_Invalid_PureFun_02(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_02")
}

func Test_Invalid_PureFun_03(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_03")
}

/*
	func Test_Invalid_PureFun_04(t *testing.T) {
		CheckInvalid(t, "purefun_invalid_04")
	}
*/
func Test_Invalid_PureFun_05(t *testing.T) {
	CheckInvalid(t, "purefun_invalid_05")
}

/*
	func Test_Invalid_PureFun_06(t *testing.T) {
		CheckInvalid(t, "purefun_invalid_06")
	}
*/

// ===================================================================
// Test Helpers
// ===================================================================

// Check that a given source file fails to compiler.
func CheckInvalid(t *testing.T, test string) {
	filename := fmt.Sprintf("%s.lisp", test)
	// Enable testing each trace in parallel
	t.Parallel()
	// Read constraints file
	bytes, err := os.ReadFile(fmt.Sprintf("%s/%s", InvalidTestDir, filename))
	// Check test file read ok
	if err != nil {
		t.Fatal(err)
	}
	// Package up as source file
	srcfile := sexp.NewSourceFile(filename, bytes)
	// Parse terms into an HIR schema
	_, errs := corset.CompileSourceFile(srcfile)
	// Check program did not compile!
	if len(errs) == 0 {
		t.Fatalf("Error %s should not have compiled\n", filename)
	}
}