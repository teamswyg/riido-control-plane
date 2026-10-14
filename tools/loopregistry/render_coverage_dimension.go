package main

import (
	"fmt"
	"strings"
)

func renderCoverageDimensions(b *strings.Builder, dimensions []loopCoverageDimensionSurface) {
	fmt.Fprintln(b, "## Coverage Dimensions")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Dimension | Loop field | Claim field | Loop token | Claim token |")
	fmt.Fprintln(b, "| --- | --- | --- | --- | --- |")
	for _, dim := range dimensions {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			dim.ID, dim.LoopField, dim.ClaimField,
			dim.LoopTokenLabel, dim.ClaimTokenLabel)
	}
	fmt.Fprintln(b)
}
