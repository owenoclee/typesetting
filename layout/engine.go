package layout

import (
	"github.com/owen/typesetting/breaking"
	"github.com/owen/typesetting/fonts"
	"github.com/owen/typesetting/hyphenation"
	"github.com/owen/typesetting/input"
	"github.com/owen/typesetting/shaping"
)

// Engine handles the layout of text into pages.
type Engine struct {
	font       *fonts.Font
	fontSize   float64
	pageConfig *PageConfig
	lineHeight float64
	ascent     float64
	descent    float64
	hyphenator *hyphenation.Hyphenator
}

// NewEngine creates a new layout engine.
func NewEngine(font *fonts.Font, fontSize float64, pageConfig *PageConfig) *Engine {
	// Calculate vertical metrics
	lineHeight := fontSize * 1.2 // 120% leading
	ascent := font.ScaleToSize(font.Ascent, fontSize)
	descent := font.ScaleToSize(font.Descent, fontSize)

	return &Engine{
		font:       font,
		fontSize:   fontSize,
		pageConfig: pageConfig,
		lineHeight: lineHeight,
		ascent:     ascent,
		descent:    descent,
		hyphenator: hyphenation.New(),
	}
}

// HeadingStyle defines the typography for a heading level.
type HeadingStyle struct {
	FontSize     float64
	SpaceBefore  float64 // Space above heading
	SpaceAfter   float64 // Space below heading
}

// GetHeadingStyle returns the style for a heading level (LaTeX-inspired).
func (e *Engine) GetHeadingStyle(level int) HeadingStyle {
	// LaTeX article class inspired sizes (10pt base):
	// \section: ~17pt (1.7x), \subsection: ~14pt (1.4x), \subsubsection: ~12pt (1.2x)
	switch level {
	case 1:
		return HeadingStyle{
			FontSize:    e.fontSize * 1.7,  // ~17pt for 10pt base
			SpaceBefore: e.fontSize * 2.4,  // ~24pt
			SpaceAfter:  e.fontSize * 1.0,  // ~10pt
		}
	case 2:
		return HeadingStyle{
			FontSize:    e.fontSize * 1.4,  // ~14pt for 10pt base
			SpaceBefore: e.fontSize * 1.8,  // ~18pt
			SpaceAfter:  e.fontSize * 0.8,  // ~8pt
		}
	default: // level 3+
		return HeadingStyle{
			FontSize:    e.fontSize * 1.2,  // ~12pt for 10pt base
			SpaceBefore: e.fontSize * 1.2,  // ~12pt
			SpaceAfter:  e.fontSize * 0.6,  // ~6pt
		}
	}
}

// LayoutParagraph shapes and breaks a single paragraph into lines.
func (e *Engine) LayoutParagraph(text string) []Line {
	return e.LayoutTextAtSize(text, e.fontSize, true)
}

// LayoutHeading shapes a heading (no hyphenation, no justification).
func (e *Engine) LayoutHeading(text string, level int) []Line {
	style := e.GetHeadingStyle(level)
	return e.LayoutTextAtSize(text, style.FontSize, false)
}

// LayoutTextAtSize shapes and breaks text at a specific font size.
func (e *Engine) LayoutTextAtSize(text string, fontSize float64, justify bool) []Line {
	// Use simple shaper at the specified size
	shaper := shaping.NewSimpleShaper(e.font, fontSize)
	shapedGlyphs := shaper.Shape(text)

	// Convert to break items
	spaceWidth := e.font.GetSpaceWidth(fontSize)
	var items []breaking.Item
	if justify {
		// Full paragraph with hyphenation
		hyphenWidth := e.font.GetHyphenWidth(fontSize)
		items = breaking.ShapedTextToItems(shapedGlyphs, spaceWidth, text, e.hyphenator, hyphenWidth)
	} else {
		// Heading: no hyphenation
		items = breaking.ShapedTextToItems(shapedGlyphs, spaceWidth, text, nil, 0)
	}

	// Break lines using Knuth-Plass
	lineWidth := e.pageConfig.TextWidth()
	// Tolerance controls how "loose" lines can be before being rejected
	// Lower = tighter lines, higher = allows looser lines
	// TeX default is around 1.0; we use 3.0 to allow flexibility with narrow columns
	tolerance := 3.0
	breakpoints := breaking.BreakParagraph(items, lineWidth, tolerance)

	// Position glyphs
	lines := PositionGlyphs(items, breakpoints, shapedGlyphs, lineWidth)

	// Set vertical metrics for each line
	lineHeight := fontSize * 1.2
	ascent := e.font.ScaleToSize(e.font.Ascent, fontSize)
	descent := e.font.ScaleToSize(e.font.Descent, fontSize)
	for i := range lines {
		lines[i].Height = lineHeight
		lines[i].Ascent = ascent
		lines[i].Descent = -descent
		lines[i].FontSize = fontSize
	}

	return lines
}

// LayoutDocument layouts multiple paragraphs with pagination.
// Deprecated: Use LayoutContentBlocks for proper heading support.
func (e *Engine) LayoutDocument(paragraphs []string) []Page {
	// Convert to content blocks for backwards compatibility
	blocks := make([]input.ContentBlock, len(paragraphs))
	for i, text := range paragraphs {
		blocks[i] = input.ContentBlock{
			Type:         "paragraph",
			Text:         text,
			HeadingLevel: 0,
		}
	}
	return e.LayoutContentBlocks(blocks)
}

// LayoutContentBlocks layouts content blocks (paragraphs and headings) with pagination.
func (e *Engine) LayoutContentBlocks(blocks []input.ContentBlock) []Page {
	pages := make([]Page, 0)
	currentPage := Page{Number: 1}

	// Start at top margin
	y := e.pageConfig.Height - e.pageConfig.MarginTop - e.ascent
	isFirstBlock := true

	for _, block := range blocks {
		var lines []Line
		var spaceAfter float64

		if block.Type == "heading" {
			style := e.GetHeadingStyle(block.HeadingLevel)

			// Add space before heading (unless it's the first block)
			if !isFirstBlock {
				y -= style.SpaceBefore
			}

			// Layout heading
			lines = e.LayoutHeading(block.Text, block.HeadingLevel)
			spaceAfter = style.SpaceAfter
		} else {
			// Regular paragraph
			lines = e.LayoutParagraph(block.Text)
			spaceAfter = e.fontSize * 0.5
		}

		// Add lines to pages
		for i, line := range lines {
			// For headings, use the heading's ascent for the first line
			lineAscent := line.Ascent
			if i == 0 && block.Type == "heading" {
				lineAscent = line.Ascent
			}

			// Check if we need a page break
			if y-lineAscent < e.pageConfig.MarginBottom {
				// Finish current page
				pages = append(pages, currentPage)

				// Start new page
				currentPage = Page{Number: len(pages) + 1}
				y = e.pageConfig.Height - e.pageConfig.MarginTop - lineAscent
			}

			// Set baseline Y coordinate
			line.BaselineY = y
			currentPage.Lines = append(currentPage.Lines, line)

			// Move down by line height
			y -= line.Height
		}

		// Add spacing after this block
		y -= spaceAfter
		isFirstBlock = false
	}

	// Add final page if it has content
	if len(currentPage.Lines) > 0 {
		pages = append(pages, currentPage)
	}

	return pages
}

// RenderToPDF renders the pages to a PDF writer.
func (e *Engine) RenderToPDF(pages []Page, writer interface{}) error {
	// TODO: Implement PDF rendering
	// This will use the pdf.Writer to emit positioned glyphs
	return nil
}
