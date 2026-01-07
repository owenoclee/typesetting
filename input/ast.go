package input

// Document represents a parsed markdown document.
type Document struct {
	Blocks []Block
}

// Block is an interface for block-level elements.
type Block interface {
	blockType()
}

// Paragraph represents a paragraph of text.
type Paragraph struct {
	Runs []Run
}

func (p *Paragraph) blockType() {}

// Heading represents a heading.
type Heading struct {
	Level int    // 1-6
	Text  string
}

func (h *Heading) blockType() {}

// Run represents a span of text with styling.
type Run struct {
	Text   string
	Bold   bool
	Italic bool
}

// PlainText returns the plain text content of the document.
func (d *Document) PlainText() string {
	var text string
	for _, block := range d.Blocks {
		switch b := block.(type) {
		case *Paragraph:
			for _, run := range b.Runs {
				text += run.Text
			}
			text += "\n"
		case *Heading:
			text += b.Text + "\n"
		}
	}
	return text
}

// Paragraphs returns all paragraph blocks as plain text.
// Deprecated: Use ContentBlocks() for proper heading support.
func (d *Document) Paragraphs() []string {
	var paragraphs []string
	for _, block := range d.Blocks {
		switch b := block.(type) {
		case *Paragraph:
			var paraText string
			for _, run := range b.Runs {
				paraText += run.Text
			}
			paragraphs = append(paragraphs, paraText)
		case *Heading:
			paragraphs = append(paragraphs, b.Text)
		}
	}
	return paragraphs
}

// ContentBlock represents a block of content with type information.
type ContentBlock struct {
	Type         string // "paragraph" or "heading"
	Text         string
	HeadingLevel int // 1-6 for headings, 0 for paragraphs
}

// ContentBlocks returns all blocks with type information preserved.
func (d *Document) ContentBlocks() []ContentBlock {
	var blocks []ContentBlock
	for _, block := range d.Blocks {
		switch b := block.(type) {
		case *Paragraph:
			var paraText string
			for _, run := range b.Runs {
				paraText += run.Text
			}
			blocks = append(blocks, ContentBlock{
				Type:         "paragraph",
				Text:         paraText,
				HeadingLevel: 0,
			})
		case *Heading:
			blocks = append(blocks, ContentBlock{
				Type:         "heading",
				Text:         b.Text,
				HeadingLevel: b.Level,
			})
		}
	}
	return blocks
}
