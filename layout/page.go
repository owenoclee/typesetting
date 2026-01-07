package layout

// PageConfig defines the page dimensions and margins.
type PageConfig struct {
	Width        float64
	Height       float64
	MarginTop    float64
	MarginBottom float64
	MarginLeft   float64
	MarginRight  float64
}

// NewA4Page creates a PageConfig for A4 paper with LaTeX article class defaults.
// LaTeX 10pt article on A4: textwidth=345pt, textheight=43*12pt=516pt
func NewA4Page() *PageConfig {
	// A4: 595.28pt × 841.89pt
	// LaTeX default textwidth for 10pt article: 345pt
	// Horizontal margins: (595.28 - 345) / 2 = 125.14pt each (symmetric)
	// LaTeX default textheight: 43 * 12pt baselineskip = 516pt
	// Vertical margins: (841.89 - 516) / 2 ≈ 163pt each
	// But LaTeX uses ~72pt top margin with header, so we adjust
	return &PageConfig{
		Width:        595.28, // A4 width in points
		Height:       841.89, // A4 height in points
		MarginTop:    89,     // ~1.24 inch - room for header
		MarginBottom: 89,
		MarginLeft:   125,    // (595.28 - 345) / 2 ≈ 125pt
		MarginRight:  125,    // Symmetric margins
	}
}

// TextWidth returns the available width for text content.
func (p *PageConfig) TextWidth() float64 {
	return p.Width - p.MarginLeft - p.MarginRight
}

// TextHeight returns the available height for text content.
func (p *PageConfig) TextHeight() float64 {
	return p.Height - p.MarginTop - p.MarginBottom
}

// Page represents a single page with positioned lines.
type Page struct {
	Lines  []Line
	Number int
}
