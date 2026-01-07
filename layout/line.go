package layout

import (
	"github.com/owen/typesetting/breaking"
	"github.com/owen/typesetting/pdf"
	"github.com/owen/typesetting/shaping"
)

// Line represents a line of text with positioned glyphs.
type Line struct {
	Glyphs    []pdf.PositionedGlyph
	BaselineY float64 // Y coordinate of baseline
	Height    float64 // Line height
	Ascent    float64 // Ascent above baseline
	Descent   float64 // Descent below baseline
	FontSize  float64 // Font size for this line (for headings)
}

// PositionGlyphs converts break items and shaped glyphs into positioned lines.
func PositionGlyphs(items []breaking.Item, breakpoints []breaking.Breakpoint,
	shapedGlyphs []shaping.ShapedGlyph, lineWidth float64) []Line {

	lines := make([]Line, 0, len(breakpoints)-1)

	// Map from item index to shaped glyph index
	itemToGlyph := make(map[int]int)
	glyphIdx := 0
	for i, item := range items {
		if box, ok := item.(breaking.Box); ok {
			itemToGlyph[i] = glyphIdx
			// Find matching shaped glyph by GlyphID and Cluster
			for glyphIdx < len(shapedGlyphs) {
				if shapedGlyphs[glyphIdx].GlyphID == box.GlyphID &&
					shapedGlyphs[glyphIdx].Cluster == box.Cluster {
					glyphIdx++
					break
				}
				glyphIdx++
			}
		}
	}

	// Process each line (between consecutive breakpoints)
	for i := 0; i < len(breakpoints)-1; i++ {
		startPos := breakpoints[i].Position
		endPos := breakpoints[i+1].Position

		// Don't justify the last line of a paragraph
		isLastLine := (i == len(breakpoints)-2)

		// For lines after the first, skip the break item from previous line
		// (glue disappears, penalty was already rendered as hyphen)
		if i > 0 && startPos < len(items) {
			if items[startPos].IsGlue() || items[startPos].IsPenalty() {
				startPos++
			}
		}

		// Check if we're breaking at a flagged penalty (hyphenation)
		breakAtHyphen := false
		if endPos < len(items) {
			if pen, ok := items[endPos].(breaking.Penalty); ok && pen.Flagged {
				breakAtHyphen = true
			}
		}

		// First pass: calculate natural width and total stretch for this line
		// Skip penalties - their width only counts if we break at them (handled separately)
		naturalWidth := 0.0
		totalStretch := 0.0
		totalShrink := 0.0
		for j := startPos; j < endPos && j < len(items); j++ {
			item := items[j]
			if !item.IsPenalty() {
				naturalWidth += item.Width()
			}
			totalStretch += item.Stretchability()
			totalShrink += item.Shrinkability()
		}

		// Build positioned glyphs for this line
		x := 0.0
		lineGlyphs := make([]pdf.PositionedGlyph, 0)

		for j := startPos; j < endPos && j < len(items); j++ {
			item := items[j]

			switch v := item.(type) {
			case breaking.Box:
				// Add glyph at current position
				lineGlyphs = append(lineGlyphs, pdf.PositionedGlyph{
					GlyphID: v.GlyphID,
					X:       x,
					Y:       0, // Relative to baseline
				})
				x += v.Width()

			case breaking.Glue:
				// Adjust glue to justify the line
				glueWidth := v.Width()
				if !isLastLine && totalStretch > 0 && naturalWidth < lineWidth {
					// Calculate this glue's share of the stretch needed
					extraSpace := lineWidth - naturalWidth
					glueWidth += extraSpace * (v.Stretchability() / totalStretch)
				} else if totalShrink > 0 && naturalWidth > lineWidth {
					// Shrink - applies to ALL lines including last line
					// (last line shouldn't stretch, but must shrink if too long)
					shrinkAmount := naturalWidth - lineWidth
					glueWidth -= shrinkAmount * (v.Shrinkability() / totalShrink)
				}
				x += glueWidth

			case breaking.Penalty:
				// Skip penalties within the line (only render at break point)
			}
		}

		// Add hyphen at end of line if breaking at hyphenation point
		if breakAtHyphen {
			if pen, ok := items[endPos].(breaking.Penalty); ok {
				lineGlyphs = append(lineGlyphs, pdf.PositionedGlyph{
					GlyphID: 45, // Hyphen character '-'
					X:       x,
					Y:       0,
				})
				x += pen.Width()
			}
		}

		lines = append(lines, Line{
			Glyphs: lineGlyphs,
		})
	}

	return lines
}
