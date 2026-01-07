package shaping

// ShapedGlyph represents a shaped glyph with positioning information.
type ShapedGlyph struct {
	GlyphID  uint32  // Glyph ID in the font
	Cluster  int     // Byte offset in input text
	XAdvance float64 // Horizontal advance in points
	YAdvance float64 // Vertical advance in points
	XOffset  float64 // Horizontal offset in points
	YOffset  float64 // Vertical offset in points
}

// NOTE: Full HarfBuzz integration is deferred to a future enhancement.
// For tier 1, we use the simple shaper which provides basic functionality
// without ligatures or complex shaping features.
// See simple.go for the implementation.
