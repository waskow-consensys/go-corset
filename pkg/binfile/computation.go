package binfile

import (
	"github.com/consensys/go-corset/pkg/hir"
)

type jsonComputationSet struct {
	Computations []jsonComputation `json:"computations"`
}

type jsonComputation struct {
	Sorted *jsonSortedComputation
}

type jsonSortedComputation struct {
	Froms []string `json:"froms"`
	Tos   []string `json:"tos"`
	Signs []bool   `json:"signs"`
}

// =============================================================================
// Translation
// =============================================================================

func (e jsonComputationSet) addToSchema(schema *hir.Schema) {
	for _, c := range e.Computations {
		if c.Sorted != nil {
			targets := asColumnRefs(c.Sorted.Tos)
			sources := asColumnRefs(c.Sorted.Froms)
			schema.AddPermutationColumns(targets, c.Sorted.Signs, sources)
		}
	}
}