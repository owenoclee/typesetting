package input

import (
	"strings"
)

// Parse parses markdown-lite text into a Document AST.
func Parse(text string) (*Document, error) {
	doc := &Document{
		Blocks: make([]Block, 0),
	}

	lines := strings.Split(text, "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			i++
			continue
		}

		// Check for heading
		if strings.HasPrefix(line, "#") {
			heading := parseHeading(line)
			if heading != nil {
				doc.Blocks = append(doc.Blocks, heading)
				i++
				continue
			}
		}

		// Otherwise, it's a paragraph
		// Collect consecutive non-empty, non-heading lines
		paraLines := make([]string, 0)
		for i < len(lines) {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				break
			}
			if strings.HasPrefix(line, "#") {
				break
			}
			paraLines = append(paraLines, line)
			i++
		}

		if len(paraLines) > 0 {
			paraText := strings.Join(paraLines, " ")
			para := parseParagraph(paraText)
			doc.Blocks = append(doc.Blocks, para)
		}
	}

	return doc, nil
}

// parseHeading parses a heading line.
func parseHeading(line string) *Heading {
	level := 0
	for _, ch := range line {
		if ch == '#' {
			level++
		} else {
			break
		}
	}

	if level == 0 || level > 6 {
		return nil
	}

	text := strings.TrimSpace(line[level:])
	return &Heading{
		Level: level,
		Text:  text,
	}
}

// parseParagraph parses a paragraph with inline styles.
func parseParagraph(text string) *Paragraph {
	// For tier 1: simple implementation without full inline style support
	// We'll just create a single run with the text
	// TODO: Parse **bold** and *italic* in future enhancement
	return &Paragraph{
		Runs: []Run{
			{
				Text:   text,
				Bold:   false,
				Italic: false,
			},
		},
	}
}

// parseInlineStyles parses inline styles like **bold** and *italic*.
// This is a simplified implementation for tier 1.
func parseInlineStyles(text string) []Run {
	// TODO: Implement proper inline style parsing
	// For now, return single run
	return []Run{
		{
			Text:   text,
			Bold:   false,
			Italic: false,
		},
	}
}
