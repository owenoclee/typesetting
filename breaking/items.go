package breaking

// Item represents a breakable element in a paragraph.
type Item interface {
	Width() float64
	Stretchability() float64
	Shrinkability() float64
	IsPenalty() bool
	IsGlue() bool
	IsBox() bool
}

// Box represents a fixed-width element (like a glyph or character).
// Boxes cannot be broken and have no stretch or shrink.
type Box struct {
	WidthVal float64
	GlyphID  uint32 // Glyph ID in the font
	Cluster  int    // Maps back to input position
}

func (b Box) Width() float64          { return b.WidthVal }
func (b Box) Stretchability() float64 { return 0 }
func (b Box) Shrinkability() float64  { return 0 }
func (b Box) IsPenalty() bool         { return false }
func (b Box) IsGlue() bool            { return false }
func (b Box) IsBox() bool             { return true }

// Glue represents stretchable/shrinkable whitespace (like spaces between words).
type Glue struct {
	WidthVal float64 // Natural width
	Stretch  float64 // How much it can grow
	Shrink   float64 // How much it can shrink
}

func (g Glue) Width() float64          { return g.WidthVal }
func (g Glue) Stretchability() float64 { return g.Stretch }
func (g Glue) Shrinkability() float64  { return g.Shrink }
func (g Glue) IsPenalty() bool         { return false }
func (g Glue) IsGlue() bool            { return true }
func (g Glue) IsBox() bool             { return false }

// Penalty represents a potential break point with an associated cost.
// Negative penalties encourage breaks, positive discourage them.
type Penalty struct {
	WidthVal    float64 // Width if break is taken (e.g., for hyphens)
	PenaltyVal  float64 // Cost of breaking here
	Flagged     bool    // Whether to show a hyphen
}

const (
	// InfinitePenalty forces a break to be avoided
	InfinitePenalty = 10000.0
	// NegativeInfinitePenalty forces a break
	NegativeInfinitePenalty = -10000.0
)

func (p Penalty) Width() float64          { return p.WidthVal }
func (p Penalty) Stretchability() float64 { return 0 }
func (p Penalty) Shrinkability() float64  { return 0 }
func (p Penalty) IsPenalty() bool         { return true }
func (p Penalty) IsGlue() bool            { return false }
func (p Penalty) IsBox() bool             { return false }

// Breakpoint represents a chosen line break position.
type Breakpoint struct {
	Position int     // Index in the item list
	Line     int     // Line number (0-indexed)
	Ratio    float64 // Adjustment ratio for justification
}
