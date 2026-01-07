package fonts

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"sort"
	"strings"
)

// SubsetInfo contains information about a font subset for PDF embedding.
type SubsetInfo struct {
	GlyphIDs      []uint32          // Sorted list of glyph IDs to include
	GlyphMapping  map[uint32]uint32 // Old GID → New GID
	UsedCodepoints map[rune]uint32  // Codepoint → Old GID (for ToUnicode CMap)
}

// CreateSubset creates a subset mapping for the given list of used glyphs.
// For tier 1, we keep the mapping simple and don't actually reconstruct the font tables.
func (f *Font) CreateSubset(usedGlyphs map[uint32]bool) *SubsetInfo {
	// Sort glyph IDs
	glyphIDs := make([]uint32, 0, len(usedGlyphs))
	for gid := range usedGlyphs {
		glyphIDs = append(glyphIDs, gid)
	}
	sort.Slice(glyphIDs, func(i, j int) bool {
		return glyphIDs[i] < glyphIDs[j]
	})

	// Create mapping: old GID → new GID
	glyphMapping := make(map[uint32]uint32)
	for newGID, oldGID := range glyphIDs {
		glyphMapping[oldGID] = uint32(newGID)
	}

	// Build reverse map: codepoint → old GID
	usedCodepoints := make(map[rune]uint32)
	for r, gid := range f.CMap {
		if usedGlyphs[gid] {
			usedCodepoints[r] = gid
		}
	}

	return &SubsetInfo{
		GlyphIDs:       glyphIDs,
		GlyphMapping:   glyphMapping,
		UsedCodepoints: usedCodepoints,
	}
}

// GetSubsetFontData returns the font data for embedding.
// For tier 1 implementation, we embed the full font (not truly subsetted).
// True subsetting requires rebuilding TTF tables, which is complex.
func (f *Font) GetSubsetFontData(subset *SubsetInfo) ([]byte, error) {
	// For tier 1: embed full font
	// TODO: Implement true subsetting in phase 2.1
	return f.FontData, nil
}

// GetCompressedFontStream returns a compressed font stream for PDF embedding.
func (f *Font) GetCompressedFontStream(subset *SubsetInfo) ([]byte, error) {
	fontData, err := f.GetSubsetFontData(subset)
	if err != nil {
		return nil, err
	}

	// Compress with zlib
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(fontData); err != nil {
		return nil, fmt.Errorf("failed to compress font: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compressor: %w", err)
	}

	return buf.Bytes(), nil
}

// BuildToUnicodeCMap generates a ToUnicode CMap for PDF text extraction.
func (f *Font) BuildToUnicodeCMap(subset *SubsetInfo) string {
	var buf strings.Builder

	buf.WriteString("/CIDInit /ProcSet findresource begin\n")
	buf.WriteString("12 dict begin\n")
	buf.WriteString("begincmap\n")
	buf.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	buf.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	buf.WriteString("/CMapType 2 def\n")
	buf.WriteString("1 begincodespacerange\n")
	buf.WriteString("<0000> <FFFF>\n")
	buf.WriteString("endcodespacerange\n")

	// Build sorted list of mappings
	type mapping struct {
		newGID uint32
		r      rune
	}
	mappings := make([]mapping, 0, len(subset.UsedCodepoints))
	for r, oldGID := range subset.UsedCodepoints {
		if newGID, ok := subset.GlyphMapping[oldGID]; ok {
			mappings = append(mappings, mapping{newGID: newGID, r: r})
		}
	}
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].newGID < mappings[j].newGID
	})

	// Write bfchar mappings
	if len(mappings) > 0 {
		fmt.Fprintf(&buf, "%d beginbfchar\n", len(mappings))
		for _, m := range mappings {
			fmt.Fprintf(&buf, "<%04X> <%04X>\n", m.newGID, m.r)
		}
		buf.WriteString("endbfchar\n")
	}

	buf.WriteString("endcmap\n")
	buf.WriteString("CMapName currentdict /CMap defineresource pop\n")
	buf.WriteString("end\n")
	buf.WriteString("end\n")

	return buf.String()
}

// BuildWidthsArray builds a PDF widths array for the subset glyphs.
func (f *Font) BuildWidthsArray(subset *SubsetInfo, fontSize float64) []int {
	widths := make([]int, len(subset.GlyphIDs))
	for i, oldGID := range subset.GlyphIDs {
		width := f.GetWidth(oldGID)
		// Convert from font units to PDF units (1/1000 of font size)
		pdfWidth := int(float64(width) / float64(f.UnitsPerEm) * 1000)
		widths[i] = pdfWidth
	}
	return widths
}

// GetFontDescriptor returns font descriptor values for PDF embedding.
type FontDescriptor struct {
	FontName   string
	Flags      int // Font flags (fixed pitch, serif, symbolic, etc.)
	FontBBox   [4]int
	ItalicAngle int
	Ascent     int
	Descent    int
	CapHeight  int
	StemV      int // Stem width (estimated)
}

// GetFontDescriptor returns a font descriptor for PDF embedding.
func (f *Font) GetFontDescriptor() FontDescriptor {
	// Get font name from sfnt (simplified - we'll use a placeholder)
	fontName := "CustomFont"

	// Flags: 32 = symbolic (non-standard character set)
	flags := 32

	// Font BBox (bounding box) - estimate from metrics
	bbox := [4]int{
		0,
		f.Descent,
		int(float64(f.UnitsPerEm) * 0.95),
		f.Ascent,
	}

	// Stem width (estimated as 1/10 of UnitsPerEm)
	stemV := f.UnitsPerEm / 10

	return FontDescriptor{
		FontName:    fontName,
		Flags:       flags,
		FontBBox:    bbox,
		ItalicAngle: 0,
		Ascent:      f.Ascent,
		Descent:     f.Descent,
		CapHeight:   f.CapHeight,
		StemV:       stemV,
	}
}
