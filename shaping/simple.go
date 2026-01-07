package shaping

import (
	"github.com/owen/typesetting/fonts"
)

// SimpleShaper performs basic "dumb" shaping without ligatures or complex features.
// This is useful for testing and as a fallback when HarfBuzz is not available.
type SimpleShaper struct {
	font     *fonts.Font
	fontSize float64
}

// NewSimpleShaper creates a new simple shaper.
func NewSimpleShaper(font *fonts.Font, fontSize float64) *SimpleShaper {
	return &SimpleShaper{
		font:     font,
		fontSize: fontSize,
	}
}

// Shape performs simple shaping by mapping each rune to a glyph.
func (s *SimpleShaper) Shape(text string) []ShapedGlyph {
	runes := []rune(text)
	result := make([]ShapedGlyph, 0, len(runes))

	for i, r := range runes {
		// Skip spaces - they will be detected as gaps in the input string
		if r == ' ' {
			continue
		}

		// Skip control characters
		if r < 32 && r != '\n' && r != '\t' {
			continue
		}

		// Get glyph ID
		gid, ok := s.font.GetGlyphID(r)
		if !ok {
			// Use .notdef glyph (usually glyph 0)
			gid = 0
		}

		// Get advance width
		width := s.font.GetWidth(gid)
		advance := s.font.ScaleToSize(width, s.fontSize)

		// Apply kerning with previous glyph if available
		if len(result) > 0 {
			prevGID := result[len(result)-1].GlyphID
			kern := s.font.GetKerning(prevGID, gid)
			if kern != 0 {
				kernAdj := s.font.ScaleToSize(kern, s.fontSize)
				// Apply kern by adjusting previous glyph's advance
				result[len(result)-1].XAdvance += kernAdj
			}
		}

		// Use byte position in original text as cluster
		// This allows proper word boundary detection
		result = append(result, ShapedGlyph{
			GlyphID:  gid,
			Cluster:  i,  // Use original position in rune slice
			XAdvance: advance,
			YAdvance: 0,
			XOffset:  0,
			YOffset:  0,
		})
	}

	return result
}
