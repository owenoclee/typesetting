package breaking

import (
	"math"
)

// Node represents a possible breakpoint in the dynamic programming algorithm.
type Node struct {
	Position     int      // Item index
	Line         int      // Line number
	Fitness      int      // 0=tight, 1=normal, 2=loose, 3=very loose
	TotalWidth   float64  // Cumulative width from start
	TotalStretch float64  // Cumulative stretch from start
	TotalShrink  float64  // Cumulative shrink from start
	Demerits     float64  // Accumulated badness
	Previous     *Node    // Previous node for backtracking
}

const (
	FitnessTight      = 0
	FitnessNormal     = 1
	FitnessLoose      = 2
	FitnessVeryLoose  = 3
)

// BreakParagraph finds optimal line breaks for the given items.
// lineWidth is the target line width, tolerance is the maximum acceptable ratio (typically 1.0).
func BreakParagraph(items []Item, lineWidth float64, tolerance float64) []Breakpoint {
	if len(items) == 0 {
		return nil
	}

	// Build cumulative sums for fast range queries
	// IMPORTANT: Penalty widths are NOT included here - they only count
	// if we actually break at that penalty (handled in computeAdjustmentRatio)
	n := len(items)
	cumWidth := make([]float64, n+1)
	cumStretch := make([]float64, n+1)
	cumShrink := make([]float64, n+1)

	for i := 0; i < n; i++ {
		// Exclude penalty widths from cumulative sum
		if items[i].IsPenalty() {
			cumWidth[i+1] = cumWidth[i] // Don't add penalty width
		} else {
			cumWidth[i+1] = cumWidth[i] + items[i].Width()
		}
		cumStretch[i+1] = cumStretch[i] + items[i].Stretchability()
		cumShrink[i+1] = cumShrink[i] + items[i].Shrinkability()
	}

	// Active nodes, grouped by fitness class
	active := make([][]*Node, 4)
	for i := range active {
		active[i] = make([]*Node, 0)
	}

	// Initialize with node at position 0
	initialNode := &Node{
		Position:     0,
		Line:         0,
		Fitness:      FitnessNormal,
		TotalWidth:   0,
		TotalStretch: 0,
		TotalShrink:  0,
		Demerits:     0,
		Previous:     nil,
	}
	active[FitnessNormal] = append(active[FitnessNormal], initialNode)

	// Main loop: consider each potential breakpoint
	for b := 0; b < n; b++ {
		// Only penalties and glue are valid breakpoints
		if !isLegalBreakpoint(items, b) {
			continue
		}

		// Check if this is a forced break (terminal/paragraph end)
		isForcedBreak := false
		if pen, ok := items[b].(Penalty); ok && pen.PenaltyVal <= NegativeInfinitePenalty {
			isForcedBreak = true
		}

		// For each fitness class
		for fitness := 0; fitness < 4; fitness++ {
			// For each active node in this fitness class
			for _, activeNode := range active[fitness] {
				ratio := computeAdjustmentRatio(activeNode.Position, b, lineWidth,
					cumWidth, cumStretch, cumShrink, items)

				// Skip if line is too tight or too loose
				// BUT: always allow forced breaks (paragraph end) - last line is ragged-right
				if !isForcedBreak && (ratio < -1 || ratio > tolerance) {
					continue
				}

				// Compute demerits for this break
				demerits := computeDemerits(ratio, getPenalty(items, b), fitness, activeNode.Fitness)
				totalDemerits := activeNode.Demerits + demerits

				// Determine fitness class for the new node
				newFitness := classifyFitness(ratio)

				// Check if we already have a node at this position with this fitness
				existingNode := findNode(active[newFitness], b)
				if existingNode == nil || totalDemerits < existingNode.Demerits {
					// Create or update node
					newNode := &Node{
						Position:     b,
						Line:         activeNode.Line + 1,
						Fitness:      newFitness,
						TotalWidth:   cumWidth[b+1],
						TotalStretch: cumStretch[b+1],
						TotalShrink:  cumShrink[b+1],
						Demerits:     totalDemerits,
						Previous:     activeNode,
					}

					if existingNode == nil {
						active[newFitness] = append(active[newFitness], newNode)
					} else {
						*existingNode = *newNode
					}
				}
			}
		}
	}

	// Find best final node - should be at the terminal penalty position
	bestNode := findBestFinalNode(active)

	// Check if we found a valid path to the end of the paragraph
	// The terminal penalty is at position n-1
	terminalPos := n - 1
	if bestNode == nil || bestNode.Position != terminalPos {
		// Fallback: greedy line breaking
		return greedyBreaking(items, lineWidth)
	}

	// Backtrack to get breakpoints
	return backtrackBreakpoints(bestNode, items, lineWidth, cumWidth, cumStretch, cumShrink)
}

// isLegalBreakpoint checks if an item can be a breakpoint.
func isLegalBreakpoint(items []Item, pos int) bool {
	if pos >= len(items) {
		return false
	}
	item := items[pos]
	// Can break at glue or penalty
	return item.IsGlue() || item.IsPenalty()
}

// computeAdjustmentRatio computes the adjustment ratio for a line from active position to break position.
// When breaking at glue, the glue item disappears (not rendered at line end), so we exclude it.
// For lines after the first, we also skip the trailing glue from the previous break.
// When breaking at a flagged penalty (hyphenation), the penalty width (hyphen) is added.
func computeAdjustmentRatio(activePos, breakPos int, lineWidth float64,
	cumWidth, cumStretch, cumShrink []float64, items []Item) float64 {

	// Start position: for lines after the first, skip the previous break's glue
	startIdx := activePos
	if activePos > 0 {
		startIdx = activePos + 1
	}

	// Width of line content (excluding the break item itself - it disappears at line end)
	naturalWidth := cumWidth[breakPos] - cumWidth[startIdx]

	// If breaking at a flagged penalty (hyphenation), add the hyphen width
	if breakPos < len(items) {
		if pen, ok := items[breakPos].(Penalty); ok && pen.Flagged {
			naturalWidth += pen.Width()
		}
	}

	if naturalWidth < lineWidth {
		// Line needs stretching
		stretch := cumStretch[breakPos] - cumStretch[startIdx]
		if stretch > 0 {
			return (lineWidth - naturalWidth) / stretch
		}
		return math.Inf(1) // Can't stretch enough
	} else if naturalWidth > lineWidth {
		// Line needs shrinking
		shrink := cumShrink[breakPos] - cumShrink[startIdx]
		if shrink > 0 {
			return (lineWidth - naturalWidth) / shrink
		}
		return math.Inf(1) // Can't shrink enough
	}

	return 0 // Perfect fit
}

// computeDemerits computes the demerits (badness) for a line break.
func computeDemerits(ratio float64, penalty float64, newFitness, prevFitness int) float64 {
	// Badness based on adjustment ratio
	badness := 100 * math.Pow(math.Abs(ratio), 3)

	// Line penalty
	linePenalty := 1.0 + badness
	if penalty >= 0 {
		linePenalty = math.Pow(linePenalty+penalty, 2)
	} else if penalty > NegativeInfinitePenalty {
		linePenalty = math.Pow(linePenalty, 2) - math.Pow(penalty, 2)
	} else {
		linePenalty = math.Pow(linePenalty, 2)
	}

	// Fitness penalty (adjacent lines with very different looseness)
	fitnessPenalty := 0.0
	if math.Abs(float64(newFitness-prevFitness)) > 1 {
		fitnessPenalty = 100.0
	}

	return linePenalty + fitnessPenalty
}

// classifyFitness determines the fitness class based on adjustment ratio.
func classifyFitness(ratio float64) int {
	if ratio < -0.5 {
		return FitnessTight
	} else if ratio <= 0.5 {
		return FitnessNormal
	} else if ratio <= 1.0 {
		return FitnessLoose
	} else {
		return FitnessVeryLoose
	}
}

// getPenalty returns the penalty value for an item.
func getPenalty(items []Item, pos int) float64 {
	if pos >= len(items) {
		return 0
	}
	item := items[pos]
	if penalty, ok := item.(Penalty); ok {
		return penalty.PenaltyVal
	}
	return 0
}

// findNode finds a node at the given position in a list.
func findNode(nodes []*Node, pos int) *Node {
	for _, node := range nodes {
		if node.Position == pos {
			return node
		}
	}
	return nil
}

// findBestFinalNode finds the node with lowest demerits at the maximum position.
func findBestFinalNode(active [][]*Node) *Node {
	// First, find the maximum position among all active nodes
	maxPos := -1
	for _, nodes := range active {
		for _, node := range nodes {
			if node.Position > maxPos {
				maxPos = node.Position
			}
		}
	}

	// Then find the best node at that position
	var best *Node
	for _, nodes := range active {
		for _, node := range nodes {
			if node.Position == maxPos {
				if best == nil || node.Demerits < best.Demerits {
					best = node
				}
			}
		}
	}
	return best
}

// backtrackBreakpoints reconstructs the breakpoint list from the final node.
func backtrackBreakpoints(finalNode *Node, items []Item, lineWidth float64,
	cumWidth, cumStretch, cumShrink []float64) []Breakpoint {

	// Build list in reverse order
	var breakpoints []Breakpoint
	node := finalNode
	for node != nil && node.Previous != nil {
		ratio := computeAdjustmentRatio(node.Previous.Position, node.Position, lineWidth,
			cumWidth, cumStretch, cumShrink, items)

		breakpoints = append(breakpoints, Breakpoint{
			Position: node.Position,
			Line:     node.Line,
			Ratio:    ratio,
		})
		node = node.Previous
	}

	// Add initial breakpoint
	breakpoints = append(breakpoints, Breakpoint{
		Position: 0,
		Line:     0,
		Ratio:    0,
	})

	// Reverse to get correct order
	for i := 0; i < len(breakpoints)/2; i++ {
		j := len(breakpoints) - 1 - i
		breakpoints[i], breakpoints[j] = breakpoints[j], breakpoints[i]
	}

	return breakpoints
}

// greedyBreaking provides a fallback greedy line breaking algorithm.
func greedyBreaking(items []Item, lineWidth float64) []Breakpoint {
	breakpoints := []Breakpoint{{Position: 0, Line: 0, Ratio: 0}}
	currentWidth := 0.0
	lineNum := 1

	for i, item := range items {
		currentWidth += item.Width()

		if currentWidth > lineWidth && isLegalBreakpoint(items, i) {
			breakpoints = append(breakpoints, Breakpoint{
				Position: i,
				Line:     lineNum,
				Ratio:    0,
			})
			currentWidth = 0
			lineNum++
		}
	}

	// Add final breakpoint at the end (if not already added)
	if len(breakpoints) == 1 || breakpoints[len(breakpoints)-1].Position != len(items)-1 {
		breakpoints = append(breakpoints, Breakpoint{
			Position: len(items) - 1,
			Line:     lineNum,
			Ratio:    0,
		})
	}

	return breakpoints
}
