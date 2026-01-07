package pdf

import (
	"fmt"
	"strings"
)

// Writer represents a PDF document builder.
type Writer struct {
	objects      []string // PDF objects
	xrefOffsets  []int    // Byte offsets for xref table
	nextObjNum   int      // Next object number
	content      strings.Builder
	pageWidth    float64
	pageHeight   float64
	pages        []int    // Object numbers of page objects
	fontObjNum   int      // Font resource object number
}

// New creates a new PDF writer with the given page dimensions.
func New(pageWidth, pageHeight float64) *Writer {
	return &Writer{
		objects:    make([]string, 0),
		nextObjNum: 1,
		pageWidth:  pageWidth,
		pageHeight: pageHeight,
		pages:      make([]int, 0),
	}
}

// NewA4 creates a new PDF writer with A4 page dimensions.
func NewA4() *Writer {
	// A4 = 210mm x 297mm = 595.28 x 841.89 points (72 points per inch)
	return New(595.28, 841.89)
}

// addObject adds a PDF object and returns its object number.
func (w *Writer) addObject(content string) int {
	objNum := w.nextObjNum
	w.nextObjNum++

	obj := fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objNum, content)
	w.objects = append(w.objects, obj)

	return objNum
}

// SetFontResource sets the font resource object number.
func (w *Writer) SetFontResource(fontObjNum int) {
	w.fontObjNum = fontObjNum
}

// AddTextContent adds text content to the current page.
func (w *Writer) AddTextContent(text string, x, y, fontSize float64, fontName string) {
	fmt.Fprintf(&w.content, "BT\n/%s %.2f Tf\n%.2f %.2f Td\n(%s) Tj\nET\n",
		fontName, fontSize, x, y, text)
}

// AddPositionedGlyphs adds positioned glyphs to the current page.
// This is used for advanced text positioning with kerning and justification.
func (w *Writer) AddPositionedGlyphs(glyphs []PositionedGlyph, baselineX, baselineY, fontSize float64, fontName string) {
	if len(glyphs) == 0 {
		return
	}

	fmt.Fprintf(&w.content, "BT\n/%s %.2f Tf\n", fontName, fontSize)

	// Position each glyph individually using Tm (text matrix)
	// This is simpler and more reliable than using TJ with adjustments
	for _, g := range glyphs {
		// Set text matrix for this glyph's position
		x := baselineX + g.X
		y := baselineY + g.Y
		fmt.Fprintf(&w.content, "1 0 0 1 %.2f %.2f Tm\n", x, y)
		// Use 2-digit hex for Type1/Core14 fonts (single-byte encoding)
		fmt.Fprintf(&w.content, "<%02X> Tj\n", g.GlyphID)
	}

	w.content.WriteString("ET\n")
}

// FinishPage completes the current page and adds it to the document.
func (w *Writer) FinishPage() {
	content := w.content.String()
	if content == "" {
		// Empty page - add placeholder content
		content = "% Empty page\n"
	}

	// Create content stream object
	contentObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content)
	contentObjNum := w.addObject(contentObj)

	// Create page object
	fontResources := ""
	if w.fontObjNum > 0 {
		fontResources = fmt.Sprintf("/Font << /F1 %d 0 R >>", w.fontObjNum)
	}

	pageObj := fmt.Sprintf("<< /Type /Page /MediaBox [0 0 %.2f %.2f] /Contents %d 0 R /Resources << %s >> >>",
		w.pageWidth, w.pageHeight, contentObjNum, fontResources)
	pageObjNum := w.addObject(pageObj)

	w.pages = append(w.pages, pageObjNum)

	// Reset content for next page
	w.content.Reset()
}

// Generate generates the complete PDF as a byte slice.
func (w *Writer) Generate() []byte {
	var pdf strings.Builder
	var xrefOffsets []int

	// Track byte offset for each object
	trackOffset := func(s string) {
		xrefOffsets = append(xrefOffsets, pdf.Len())
		pdf.WriteString(s)
	}

	// PDF Header
	pdf.WriteString("%PDF-1.7\n")
	pdf.WriteString("%\xE2\xE3\xCF\xD3\n") // Binary marker

	// Write all objects
	for _, obj := range w.objects {
		trackOffset(obj)
	}

	// Create Pages object (contains all pages)
	pagesKids := ""
	for i, pageNum := range w.pages {
		if i > 0 {
			pagesKids += " "
		}
		pagesKids += fmt.Sprintf("%d 0 R", pageNum)
	}
	pagesObj := fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", pagesKids, len(w.pages))
	pagesObjNum := w.addObject(pagesObj)

	// Update parent references in page objects
	for _, pageObjNum := range w.pages {
		// Find and update the page object
		objIndex := pageObjNum - 1
		if objIndex >= 0 && objIndex < len(w.objects) {
			oldObj := w.objects[objIndex]
			// Insert /Parent reference
			newObj := strings.Replace(oldObj, "/Type /Page ", fmt.Sprintf("/Type /Page /Parent %d 0 R ", pagesObjNum), 1)
			w.objects[objIndex] = newObj
		}
	}

	// Re-write objects with parent references
	pdf.Reset()
	xrefOffsets = nil
	pdf.WriteString("%PDF-1.7\n")
	pdf.WriteString("%\xE2\xE3\xCF\xD3\n")

	for _, obj := range w.objects {
		trackOffset(obj)
	}

	// Create Catalog
	catalogObj := fmt.Sprintf("<< /Type /Catalog /Pages %d 0 R >>", pagesObjNum)
	catalogObjNum := w.addObject(catalogObj)
	trackOffset(w.objects[len(w.objects)-1])

	// Cross-reference table
	xrefOffset := pdf.Len()
	pdf.WriteString("xref\n")
	pdf.WriteString(fmt.Sprintf("0 %d\n", len(xrefOffsets)+1))
	pdf.WriteString("0000000000 65535 f \n")
	for _, offset := range xrefOffsets {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// Trailer
	pdf.WriteString("trailer\n")
	pdf.WriteString(fmt.Sprintf("<< /Size %d /Root %d 0 R >>\n", len(xrefOffsets)+1, catalogObjNum))
	pdf.WriteString("startxref\n")
	pdf.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	pdf.WriteString("%%EOF\n")

	return []byte(pdf.String())
}

// PositionedGlyph represents a glyph with its position.
type PositionedGlyph struct {
	GlyphID uint32
	X       float64
	Y       float64
}
