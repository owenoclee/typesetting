package breaking

import (
	"github.com/owen/typesetting/hyphenation"
	"github.com/owen/typesetting/shaping"
)

// ShapedTextToItems converts shaped glyphs into a list of breakable items.
// If hyphenator is non-nil and hyphenWidth > 0, hyphenation points will be inserted.
func ShapedTextToItems(glyphs []shaping.ShapedGlyph, spaceWidth float64, text string, hyphenator *hyphenation.Hyphenator, hyphenWidth float64) []Item {
	items := make([]Item, 0, len(glyphs)*2) // Estimate: glyphs + spaces

	// Build word boundaries from cluster indices
	// A word starts after a space (gap in clusters) or at the beginning
	runes := []rune(text)

	for i := 0; i < len(glyphs); i++ {
		g := glyphs[i]

		// Add box for this glyph
		items = append(items, Box{
			WidthVal: g.XAdvance,
			GlyphID:  g.GlyphID,
			Cluster:  g.Cluster,
		})

		// Check if there's a space after this glyph
		// (detected by a gap in cluster indices)
		if i < len(glyphs)-1 && glyphs[i+1].Cluster > g.Cluster+1 {
			// Insert glue for the space
			// TeX uses approximately 1/3 space width for stretch, but we need more
			// flexibility for narrow columns. Use 100% stretch, 50% shrink.
			items = append(items, Glue{
				WidthVal: spaceWidth,
				Stretch:  spaceWidth * 1.0,   // 100% stretch
				Shrink:   spaceWidth * 0.5,   // 50% shrink
			})
		} else if hyphenator != nil && hyphenWidth > 0 && i < len(glyphs)-1 {
			// Check for hyphenation opportunity within a word
			// Find the current word by looking at cluster indices
			wordStart := findWordStart(glyphs, i, runes)
			wordEnd := findWordEnd(glyphs, i, runes)

			if wordStart >= 0 && wordEnd > wordStart {
				word := string(runes[wordStart:wordEnd])
				glyphPosInWord := g.Cluster - wordStart

				// Get hyphenation points for this word
				hyphenPoints := hyphenator.Hyphenate(word)

				// Check if current position is a hyphenation point
				for _, hp := range hyphenPoints {
					if hp == glyphPosInWord+1 { // +1 because hyphen goes after current char
						// Insert discretionary hyphen (penalty with hyphen width)
						items = append(items, Penalty{
							WidthVal:   hyphenWidth,
							PenaltyVal: 30, // Lower penalty to encourage hyphenation
							Flagged:    true,
						})
						break
					}
				}
			}
		}
	}

	// Add parfillskip: infinite stretch glue that fills the last line
	// This ensures the last line can always "reach" the right margin
	// (visually it's ragged-right, but algorithmically it's fillable)
	items = append(items, Glue{
		WidthVal: 0,
		Stretch:  10000, // "Infinite" stretch for last line
		Shrink:   0,
	})

	// Add terminal penalty to force break at paragraph end
	items = append(items, Penalty{
		WidthVal:   0,
		PenaltyVal: NegativeInfinitePenalty,
		Flagged:    false,
	})

	return items
}

// findWordStart finds the start index in runes of the word containing glyph at position i.
func findWordStart(glyphs []shaping.ShapedGlyph, i int, runes []rune) int {
	cluster := glyphs[i].Cluster

	// Scan backwards to find word start
	start := cluster
	for start > 0 && runes[start-1] != ' ' {
		start--
	}
	return start
}

// findWordEnd finds the end index in runes of the word containing glyph at position i.
func findWordEnd(glyphs []shaping.ShapedGlyph, i int, runes []rune) int {
	cluster := glyphs[i].Cluster

	// Scan forwards to find word end
	end := cluster + 1
	for end < len(runes) && runes[end] != ' ' {
		end++
	}
	return end
}

// TextToItems converts plain text into breakable items using a simple character-based approach.
// This is useful for testing without the shaping layer.
func TextToItems(text string, charWidth float64, spaceWidth float64) []Item {
	items := make([]Item, 0, len(text)*2)
	runes := []rune(text)

	for _, r := range runes {
		if r == ' ' {
			// Space becomes glue
			items = append(items, Glue{
				WidthVal: spaceWidth,
				Stretch:  spaceWidth * 0.5,
				Shrink:   spaceWidth * 0.333,
			})
		} else if r == '\n' {
			// Newline becomes forced break
			items = append(items, Penalty{
				WidthVal:   0,
				PenaltyVal: NegativeInfinitePenalty,
				Flagged:    false,
			})
		} else {
			// Regular character becomes box
			items = append(items, Box{
				WidthVal: charWidth,
				GlyphID:  uint32(r), // Use Unicode value as placeholder
				Cluster:  0,
			})
		}
	}

	// Add terminal penalty
	items = append(items, Penalty{
		WidthVal:   0,
		PenaltyVal: NegativeInfinitePenalty,
		Flagged:    false,
	})

	return items
}

// AddHyphenationPoints inserts penalty items at potential hyphenation points.
// For tier 1, we can skip this and add it later.
func AddHyphenationPoints(items []Item, hyphenWidth float64) []Item {
	// TODO: Implement hyphenation using Liang patterns
	// For now, return items unchanged
	return items
}
