package pyroscope

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"
)

// FormatQueryTable formats a Pyroscope query response as a table showing top functions.
func FormatQueryTable(w io.Writer, resp *QueryResponse) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "FUNCTION\tSELF\tTOTAL\tPERCENTAGE")

	if resp.Flamegraph == nil || len(resp.Flamegraph.Names) == 0 {
		fmt.Fprintln(tw, "(no profile data)")
		return tw.Flush()
	}

	samples := ExtractTopFunctions(resp.Flamegraph, 20)

	for _, s := range samples {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%.2f%%\n",
			truncateString(s.Name, 60),
			s.Self,
			s.Total,
			s.Percentage)
	}

	return tw.Flush()
}

// FormatProfileTypesTable formats profile types as a table.
func FormatProfileTypesTable(w io.Writer, resp *ProfileTypesResponse) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tSAMPLE_TYPE\tSAMPLE_UNIT")

	for _, pt := range resp.ProfileTypes {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			pt.ID,
			pt.Name,
			pt.SampleType,
			pt.SampleUnit)
	}

	return tw.Flush()
}

// FormatLabelsTable formats label names or values as a table.
func FormatLabelsTable(w io.Writer, labels []string) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "LABEL")

	for _, label := range labels {
		fmt.Fprintln(tw, label)
	}

	return tw.Flush()
}

// ExtractTopFunctions extracts the top N functions by self time from a flame graph.
func ExtractTopFunctions(fg *Flamegraph, limit int) []FunctionSample {
	if fg == nil || len(fg.Levels) == 0 {
		return nil
	}

	// Build a map of function name -> aggregated stats
	funcStats := make(map[string]*FunctionSample)

	// Flame graph levels have values in groups of 4: [offset, total, self, nameIndex]
	for _, level := range fg.Levels {
		for i := 0; i+3 < len(level.Values); i += 4 {
			nameIdx, err := parseInt64(level.Values[i+3])
			if err != nil || nameIdx < 0 || int(nameIdx) >= len(fg.Names) {
				continue
			}
			name := fg.Names[nameIdx]

			// Skip "other" entries
			if name == "other" {
				continue
			}

			total, err := parseInt64(level.Values[i+1])
			if err != nil {
				continue
			}
			self, err := parseInt64(level.Values[i+2])
			if err != nil {
				continue
			}

			if existing, ok := funcStats[name]; ok {
				existing.Self += self
				existing.Total += total
			} else {
				funcStats[name] = &FunctionSample{
					Name:  name,
					Self:  self,
					Total: total,
				}
			}
		}
	}

	// Convert to slice and calculate percentages
	samples := make([]FunctionSample, 0, len(funcStats))
	for _, s := range funcStats {
		if fg.Total > 0 {
			s.Percentage = float64(s.Self) / float64(fg.Total) * 100
		}
		samples = append(samples, *s)
	}

	// Sort by self time descending
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Self > samples[j].Self
	})

	// Limit results
	if len(samples) > limit {
		samples = samples[:limit]
	}

	return samples
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
