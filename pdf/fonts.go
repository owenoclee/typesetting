package pdf

import (
	"compress/zlib"
	"fmt"
	"bytes"
	"strings"
)

// FontEmbedder handles embedding TrueType fonts in PDF documents.
type FontEmbedder struct {
	writer *Writer
}

// NewFontEmbedder creates a new font embedder for the given PDF writer.
func NewFontEmbedder(w *Writer) *FontEmbedder {
	return &FontEmbedder{writer: w}
}

// EmbedTrueTypeFont embeds a TrueType font with the given subset information.
// Returns the font resource object number.
func (fe *FontEmbedder) EmbedTrueTypeFont(fontData []byte, widths []int, descriptor FontDescriptor, toUnicodeCMap string) int {
	// Compress font stream
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(fontData)
	w.Close()
	compressedData := compressed.Bytes()

	// Create font file stream object (embedded font program)
	fontStreamObj := fmt.Sprintf("<< /Length %d /Filter /FlateDecode /Length1 %d >>\nstream\n",
		len(compressedData), len(fontData))
	fontStreamObj += string(compressedData)
	fontStreamObj += "\nendstream"
	fontStreamObjNum := fe.writer.addObject(fontStreamObj)

	// Create FontDescriptor object
	descriptorObj := fmt.Sprintf(`<< /Type /FontDescriptor
/FontName /%s
/Flags %d
/FontBBox [%d %d %d %d]
/ItalicAngle %d
/Ascent %d
/Descent %d
/CapHeight %d
/StemV %d
/FontFile2 %d 0 R
>>`, descriptor.FontName, descriptor.Flags,
		descriptor.FontBBox[0], descriptor.FontBBox[1], descriptor.FontBBox[2], descriptor.FontBBox[3],
		descriptor.ItalicAngle, descriptor.Ascent, descriptor.Descent,
		descriptor.CapHeight, descriptor.StemV, fontStreamObjNum)
	descriptorObjNum := fe.writer.addObject(descriptorObj)

	// Create ToUnicode CMap stream
	toUnicodeObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(toUnicodeCMap), toUnicodeCMap)
	toUnicodeObjNum := fe.writer.addObject(toUnicodeObj)

	// Build widths array
	widthsArray := "["
	for i, w := range widths {
		if i > 0 {
			widthsArray += " "
		}
		widthsArray += fmt.Sprintf("%d", w)
	}
	widthsArray += "]"

	// Create font object (TrueType font)
	fontObj := fmt.Sprintf(`<< /Type /Font
/Subtype /TrueType
/BaseFont /%s
/FirstChar 0
/LastChar %d
/Widths %s
/FontDescriptor %d 0 R
/ToUnicode %d 0 R
>>`, descriptor.FontName, len(widths)-1, widthsArray, descriptorObjNum, toUnicodeObjNum)
	fontObjNum := fe.writer.addObject(fontObj)

	return fontObjNum
}

// EmbedCore14Font creates a font resource for a PDF Core14 font (no embedding needed).
func (fe *FontEmbedder) EmbedCore14Font(fontName string) int {
	fontObj := fmt.Sprintf("<< /Type /Font /Subtype /Type1 /BaseFont /%s >>", fontName)
	return fe.writer.addObject(fontObj)
}

// FontDescriptor contains font descriptor information for PDF embedding.
type FontDescriptor struct {
	FontName    string
	Flags       int // Font flags
	FontBBox    [4]int
	ItalicAngle int
	Ascent      int
	Descent     int
	CapHeight   int
	StemV       int
}

// BuildToUnicodeCMap builds a simple ToUnicode CMap for PDF text extraction.
func BuildToUnicodeCMap(mappings map[uint32]rune) string {
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

	if len(mappings) > 0 {
		fmt.Fprintf(&buf, "%d beginbfchar\n", len(mappings))
		for gid, r := range mappings {
			fmt.Fprintf(&buf, "<%04X> <%04X>\n", gid, r)
		}
		buf.WriteString("endbfchar\n")
	}

	buf.WriteString("endcmap\n")
	buf.WriteString("CMapName currentdict /CMap defineresource pop\n")
	buf.WriteString("end\n")
	buf.WriteString("end\n")

	return buf.String()
}
