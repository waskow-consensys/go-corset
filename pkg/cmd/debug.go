package cmd

import (
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/schema"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug [flags] constraint_file",
	Short: "print constraints at various levels of expansion.",
	Long: `Print a given set of constraints at specific levels of
	expansion in order to debug them.  Constraints can be given
	either as lisp or bin files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			os.Exit(1)
		}
		hir := getFlag(cmd, "hir")
		mir := getFlag(cmd, "mir")
		air := getFlag(cmd, "air")
		// Parse constraints
		hirSchema := readSchemaFile(args[0])
		mirSchema := hirSchema.LowerToMir()
		airSchema := mirSchema.LowerToAir()
		// Print constraints
		if hir {
			printSchema(hirSchema)
		}
		if mir {
			printSchema(mirSchema)
		}
		if air {
			printSchema(airSchema)
		}
	},
}

// Print out all declarations included in a given
func printSchema(schema schema.Schema) {
	for i := schema.Declarations(); i.HasNext(); {
		fmt.Println(i.Next())
	}

	for i := schema.Constraints(); i.HasNext(); {
		fmt.Println(i.Next())
	}
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().Bool("hir", false, "Print constraints at HIR level")
	debugCmd.Flags().Bool("mir", false, "Print constraints at MIR level")
	debugCmd.Flags().Bool("air", false, "Print constraints at AIR level")
}
