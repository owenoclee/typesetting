package fonts

import (
	"fmt"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// Font represents a loaded TrueType or OpenType font with extracted metrics.
type Font struct {
	sfntFont   *sfnt.Font
	FontData   []byte           // Original font file data
	UnitsPerEm int              // Font units per em square
	Ascent     int              // Typographic ascender
	Descent    int              // Typographic descender (usually negative)
	LineGap    int              // Typographic line gap
	CapHeight  int              // Height of capital letters
	XHeight    int              // Height of lowercase 'x'
	CMap       map[rune]uint32  // Unicode codepoint → Glyph ID
	Widths     map[uint32]int   // Glyph ID → advance width (in font units)
	KernPairs  map[[2]uint32]int // (gid1, gid2) → kerning adjustment
}

// LoadFont loads a TrueType or OpenType font from the given data.
func LoadFont(fontData []byte) (*Font, error) {
	// Parse font using sfnt
	sfntFont, err := sfnt.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	font := &Font{
		sfntFont: sfntFont,
		FontData: fontData,
		CMap:     make(map[rune]uint32),
		Widths:   make(map[uint32]int),
		KernPairs: make(map[[2]uint32]int),
	}

	// Extract metrics
	if err := font.extractMetrics(); err != nil {
		return nil, fmt.Errorf("failed to extract metrics: %w", err)
	}

	// Build cmap (Unicode → Glyph ID)
	if err := font.buildCMap(); err != nil {
		return nil, fmt.Errorf("failed to build cmap: %w", err)
	}

	// Extract glyph widths
	if err := font.extractWidths(); err != nil {
		return nil, fmt.Errorf("failed to extract widths: %w", err)
	}

	// Extract kerning pairs (best effort - some fonts may not have kern table)
	_ = font.extractKerning() // Ignore error if no kern table

	return font, nil
}

// extractMetrics extracts font metrics from head, hhea, and OS/2 tables.
func (f *Font) extractMetrics() error {
	// Get units per em
	unitsPerEm := f.sfntFont.UnitsPerEm()
	f.UnitsPerEm = int(unitsPerEm)

	// Get metrics buffer
	var b sfnt.Buffer
	ppem := fixed.Int26_6(f.UnitsPerEm)

	// Get metrics
	metrics, err := f.sfntFont.Metrics(&b, ppem, font.HintingNone)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// Extract values (these are in fixed.Int26_6 format)
	f.Ascent = int(metrics.Ascent.Round())
	f.Descent = int(metrics.Descent.Round()) // Usually negative
	f.LineGap = 0 // Default to 0 if not available
	f.CapHeight = int(metrics.CapHeight.Round())
	f.XHeight = int(metrics.XHeight.Round())

	// Fallback estimates if values are zero
	if f.CapHeight == 0 {
		f.CapHeight = f.Ascent * 7 / 10
	}
	if f.XHeight == 0 {
		f.XHeight = f.Ascent * 5 / 10
	}

	return nil
}

// buildCMap builds a Unicode → Glyph ID mapping.
func (f *Font) buildCMap() error {
	var b sfnt.Buffer

	// Try to get glyph index for various Unicode ranges
	// We'll iterate through common Latin characters and punctuation
	for r := rune(32); r <= rune(126); r++ { // ASCII printable
		gid, err := f.sfntFont.GlyphIndex(&b, r)
		if err != nil {
			continue
		}
		f.CMap[r] = uint32(gid)
	}

	// Extended Latin (U+00A0 to U+00FF)
	for r := rune(0x00A0); r <= rune(0x00FF); r++ {
		gid, err := f.sfntFont.GlyphIndex(&b, r)
		if err != nil {
			continue
		}
		f.CMap[r] = uint32(gid)
	}

	// Common punctuation and symbols
	commonRunes := []rune{
		'\n', '\r', '\t', // Whitespace
		'\u201C', '\u201D', '\u2018', '\u2019', // Smart quotes
		'\u2014', '\u2013', // Dashes
		'\u2026', // Ellipsis
	}

	for _, r := range commonRunes {
		gid, err := f.sfntFont.GlyphIndex(&b, r)
		if err != nil {
			continue
		}
		f.CMap[r] = uint32(gid)
	}

	if len(f.CMap) == 0 {
		return fmt.Errorf("no valid cmap entries found")
	}

	return nil
}

// extractWidths extracts advance widths for all glyphs in the cmap.
func (f *Font) extractWidths() error {
	var b sfnt.Buffer
	ppem := fixed.Int26_6(f.UnitsPerEm)

	// Extract widths for all glyphs in cmap
	for _, gid := range f.CMap {
		advance, err := f.sfntFont.GlyphAdvance(&b, sfnt.GlyphIndex(gid), ppem, font.HintingNone)
		if err != nil {
			return fmt.Errorf("failed to get advance for glyph %d: %w", gid, err)
		}
		f.Widths[gid] = int(advance.Round())
	}

	return nil
}

// extractKerning extracts kerning pairs from the kern table.
// Returns nil if successful, error if kern table is not available (which is OK).
func (f *Font) extractKerning() error {
	var b sfnt.Buffer
	ppem := fixed.Int26_6(f.UnitsPerEm)

	// Try to extract kerning for common pairs
	// Note: sfnt doesn't expose raw kern table, so we query pair by pair
	commonPairs := [][2]rune{
		{'A', 'V'}, {'A', 'W'}, {'A', 'Y'}, {'A', 'T'},
		{'T', 'o'}, {'T', 'a'}, {'T', 'e'}, {'T', 'y'},
		{'V', 'a'}, {'V', 'e'}, {'V', 'o'}, {'V', 'A'},
		{'W', 'a'}, {'W', 'e'}, {'W', 'o'}, {'W', 'A'},
		{'Y', 'a'}, {'Y', 'e'}, {'Y', 'o'}, {'Y', 'A'},
		{'L', 'T'}, {'L', 'V'}, {'L', 'W'}, {'L', 'Y'},
		{'P', 'a'}, {'P', 'A'}, {'P', '.'}, {'P', ','},
		{'R', 'V'}, {'R', 'W'}, {'R', 'Y'}, {'R', 'T'},
		{'F', 'a'}, {'F', 'A'}, {'F', '.'}, {'F', ','},
	}

	for _, pair := range commonPairs {
		gid1, ok1 := f.CMap[pair[0]]
		gid2, ok2 := f.CMap[pair[1]]
		if !ok1 || !ok2 {
			continue
		}

		// Get kern value
		kern, err := f.sfntFont.Kern(&b, sfnt.GlyphIndex(gid1), sfnt.GlyphIndex(gid2), ppem, font.HintingNone)
		if err != nil {
			continue
		}

		kernVal := int(kern.Round())
		if kernVal != 0 {
			f.KernPairs[[2]uint32{gid1, gid2}] = kernVal
		}
	}

	return nil
}

// GetGlyphID returns the glyph ID for a given Unicode codepoint.
func (f *Font) GetGlyphID(r rune) (uint32, bool) {
	gid, ok := f.CMap[r]
	return gid, ok
}

// GetWidth returns the advance width for a glyph in font units.
func (f *Font) GetWidth(gid uint32) int {
	return f.Widths[gid]
}

// GetKerning returns the kerning adjustment between two glyphs in font units.
func (f *Font) GetKerning(gid1, gid2 uint32) int {
	return f.KernPairs[[2]uint32{gid1, gid2}]
}

// GetSpaceWidth returns the width of the space character at the given font size.
func (f *Font) GetSpaceWidth(fontSize float64) float64 {
	spaceGID, ok := f.GetGlyphID(' ')
	if !ok {
		// Fallback: estimate as 1/4 of em
		return fontSize / 4.0
	}
	width := f.GetWidth(spaceGID)
	return float64(width) / float64(f.UnitsPerEm) * fontSize
}

// GetHyphenWidth returns the width of the hyphen character at the given font size.
func (f *Font) GetHyphenWidth(fontSize float64) float64 {
	hyphenGID, ok := f.GetGlyphID('-')
	if !ok {
		// Fallback: estimate as 1/3 of em
		return fontSize / 3.0
	}
	width := f.GetWidth(hyphenGID)
	return float64(width) / float64(f.UnitsPerEm) * fontSize
}

// ScaleToSize converts a value in font units to points at the given font size.
func (f *Font) ScaleToSize(value int, fontSize float64) float64 {
	return float64(value) / float64(f.UnitsPerEm) * fontSize
}
