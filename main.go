package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/owen/typesetting/fonts"
	"github.com/owen/typesetting/input"
	"github.com/owen/typesetting/layout"
	"github.com/owen/typesetting/pdf"
)

func main() {
	// Parse command-line flags
	inputFile := flag.String("input", "", "Input markdown file (if empty, uses demo text)")
	outputFile := flag.String("output", "output.pdf", "Output PDF file")
	fontPath := flag.String("font", "", "Path to TTF/OTF font (if empty, uses Core14 Helvetica)")
	fontSize := flag.Float64("font-size", 10, "Font size in points (LaTeX default: 10pt)")
	flag.Parse()

	// Read input text
	var text string
	if *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
		text = string(data)
	} else {
		// Use demo text
		text = getDemoText()
	}

	// Parse markdown
	doc, err := input.Parse(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing markdown: %v\n", err)
		os.Exit(1)
	}

	// Get content blocks (paragraphs and headings)
	blocks := doc.ContentBlocks()

	if len(blocks) == 0 {
		fmt.Fprintln(os.Stderr, "No content to typeset")
		os.Exit(1)
	}

	// Setup PDF writer
	writer := pdf.NewA4()
	embedder := pdf.NewFontEmbedder(writer)

	// Load font
	var font *fonts.Font
	var fontObjNum int

	if *fontPath != "" {
		// Load custom font
		fontData, err := os.ReadFile(*fontPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading font file: %v\n", err)
			os.Exit(1)
		}

		font, err = fonts.LoadFont(fontData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading font: %v\n", err)
			os.Exit(1)
		}

		// For now, use Core14 as placeholder until we properly embed custom fonts
		fontObjNum = embedder.EmbedCore14Font("Times-Roman")
	} else {
		// Use Core14 Times-Roman (serif font, closer to LaTeX's Computer Modern)
		fontObjNum = embedder.EmbedCore14Font("Times-Roman")

		// Create a minimal font wrapper for Core14
		font = createCore14Font()
	}

	writer.SetFontResource(fontObjNum)

	// Create layout engine
	pageConfig := layout.NewA4Page()
	engine := layout.NewEngine(font, *fontSize, pageConfig)

	// Layout document with headings
	pages := engine.LayoutContentBlocks(blocks)

	// Render pages to PDF
	for _, page := range pages {
		for _, line := range page.Lines {
			// Calculate absolute X position (left margin)
			baselineX := pageConfig.MarginLeft

			// Use line's font size (varies for headings)
			lineFontSize := line.FontSize
			if lineFontSize == 0 {
				lineFontSize = *fontSize // Fallback to default
			}

			// Add positioned glyphs
			writer.AddPositionedGlyphs(line.Glyphs, baselineX, line.BaselineY, lineFontSize, "F1")
		}
		writer.FinishPage()
	}

	// Generate PDF
	pdfData := writer.Generate()

	// Write to file
	err = os.WriteFile(*outputFile, pdfData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created %s\n", *outputFile)
}

// getDemoText returns sample text for testing.
func getDemoText() string {
	return `# A Treatise on Typography

The art of typography is one of the most fundamental aspects of written communication. Through careful arrangement of type, the typographer seeks to create a harmonious balance between readability and aesthetic beauty.

# The Golden Ratio

Many typographers have long relied on mathematical principles to guide their work. The golden ratio, approximately 1.618, appears throughout nature and has been employed in typography to determine ideal proportions between line length, leading, and margins.

# Line Breaking Algorithms

Modern typesetting systems employ sophisticated algorithms to determine optimal line breaks. The Knuth-Plass algorithm, developed for TeX, uses dynamic programming to minimize the overall "badness" of a paragraph by considering multiple possible break points simultaneously.

This approach produces significantly better results than simple greedy algorithms, which make locally optimal choices without considering the global context of the paragraph.`
}

// createCore14Font creates a minimal font wrapper for Times-Roman.
func createCore14Font() *fonts.Font {
	// Metrics for Times-Roman (serif font, closer to LaTeX's Computer Modern)
	return &fonts.Font{
		UnitsPerEm: 1000,
		Ascent:     683,
		Descent:    -217,
		LineGap:    0,
		CapHeight:  662,
		XHeight:    450,
		CMap:       buildTimesRomanCMap(),
		Widths:     buildTimesRomanWidths(),
		KernPairs:  make(map[[2]uint32]int),
	}
}

// buildTimesRomanCMap creates an approximate cmap for Times-Roman.
func buildTimesRomanCMap() map[rune]uint32 {
	cmap := make(map[rune]uint32)
	// Map common ASCII characters
	for r := rune(32); r <= rune(126); r++ {
		cmap[r] = uint32(r)
	}
	return cmap
}

// buildTimesRomanWidths creates widths for Times-Roman based on the PDF spec.
// These are the standard widths in font units (1000 units per em) from Adobe's AFM files.
func buildTimesRomanWidths() map[uint32]int {
	// Times-Roman character widths from Adobe's AFM (Adobe Font Metrics)
	// Source: PDF Reference, Appendix D - Standard Type 1 Fonts
	widths := map[uint32]int{
		32:  250,  // space
		33:  333,  // exclam
		34:  408,  // quotedbl
		35:  500,  // numbersign
		36:  500,  // dollar
		37:  833,  // percent
		38:  778,  // ampersand
		39:  180,  // quotesingle
		40:  333,  // parenleft
		41:  333,  // parenright
		42:  500,  // asterisk
		43:  564,  // plus
		44:  250,  // comma
		45:  333,  // hyphen
		46:  250,  // period
		47:  278,  // slash
		48:  500,  // zero
		49:  500,  // one
		50:  500,  // two
		51:  500,  // three
		52:  500,  // four
		53:  500,  // five
		54:  500,  // six
		55:  500,  // seven
		56:  500,  // eight
		57:  500,  // nine
		58:  278,  // colon
		59:  278,  // semicolon
		60:  564,  // less
		61:  564,  // equal
		62:  564,  // greater
		63:  444,  // question
		64:  921,  // at
		65:  722,  // A
		66:  667,  // B
		67:  667,  // C
		68:  722,  // D
		69:  611,  // E
		70:  556,  // F
		71:  722,  // G
		72:  722,  // H
		73:  333,  // I
		74:  389,  // J
		75:  722,  // K
		76:  611,  // L
		77:  889,  // M
		78:  722,  // N
		79:  722,  // O
		80:  556,  // P
		81:  722,  // Q
		82:  667,  // R
		83:  556,  // S
		84:  611,  // T
		85:  722,  // U
		86:  722,  // V
		87:  944,  // W
		88:  722,  // X
		89:  722,  // Y
		90:  611,  // Z
		91:  333,  // bracketleft
		92:  278,  // backslash
		93:  333,  // bracketright
		94:  469,  // asciicircum
		95:  500,  // underscore
		96:  333,  // grave
		97:  444,  // a
		98:  500,  // b
		99:  444,  // c
		100: 500,  // d
		101: 444,  // e
		102: 333,  // f
		103: 500,  // g
		104: 500,  // h
		105: 278,  // i
		106: 278,  // j
		107: 500,  // k
		108: 278,  // l
		109: 778,  // m
		110: 500,  // n
		111: 500,  // o
		112: 500,  // p
		113: 500,  // q
		114: 333,  // r
		115: 389,  // s
		116: 278,  // t
		117: 500,  // u
		118: 500,  // v
		119: 722,  // w
		120: 500,  // x
		121: 500,  // y
		122: 444,  // z
		123: 480,  // braceleft
		124: 200,  // bar
		125: 480,  // braceright
		126: 541,  // asciitilde
	}
	return widths
}
