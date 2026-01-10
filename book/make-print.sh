#!/bin/bash
# Convert a LaTeX file to a printable A4 booklet (folded to A5)
# Usage: ./make-print.sh [-b|-m] input.tex
#   -b: Book mode (separate signatures stacked)
#   -m: Magazine mode (saddle-stitch, all nested)

set -e

MODE=""
INPUT_TEX=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--book)
            MODE="book"
            shift
            ;;
        -m|--magazine)
            MODE="magazine"
            shift
            ;;
        *)
            INPUT_TEX="$1"
            shift
            ;;
    esac
done

# Validate inputs
if [ -z "$MODE" ]; then
    echo "Usage: $0 [-b|-m] input.tex"
    echo "  -b, --book      Book mode (separate signatures)"
    echo "  -m, --magazine  Magazine mode (saddle-stitch)"
    exit 1
fi

if [ -z "$INPUT_TEX" ]; then
    echo "Error: No input file specified"
    echo "Usage: $0 [-b|-m] input.tex"
    exit 1
fi

BASENAME="${INPUT_TEX%.tex}"
OUTPUT_PDF="${BASENAME}.pdf"
BOOKLET_PDF="${BASENAME}-${MODE}.pdf"

echo "Step 1: Compiling LaTeX to PDF (A5 pages)..."
docker run --rm -v "$(pwd)":/workdir -w /workdir texlive/texlive:latest \
    pdflatex -interaction=nonstopmode "$INPUT_TEX"

# Run twice for TOC/refs if needed
docker run --rm -v "$(pwd)":/workdir -w /workdir texlive/texlive:latest \
    pdflatex -interaction=nonstopmode "$INPUT_TEX"

if [ ! -f "$OUTPUT_PDF" ]; then
    echo "Error: PDF generation failed"
    exit 1
fi

# Clean up intermediate files
mkdir -p .extra
mv *.aux *.log *.out .extra/ 2>/dev/null || true

echo "Step 2: Creating booklet layout (${MODE} mode)..."

if [ "$MODE" = "book" ]; then
    # Book mode: separate signatures using pdfjam
    docker run --rm -v "$(pwd)":/workdir -w /workdir texlive/texlive:latest \
        pdfjam --landscape --signature 4 --suffix "${MODE}" --letterpaper false --a4paper "$OUTPUT_PDF"
else
    # Magazine mode: saddle-stitch using pdfbook2 with short-edge binding
    # --no-crop prevents auto-scaling to keep margins consistent with book mode

    # Save existing -book.pdf if it exists to prevent overwriting
    if [ -f "${BASENAME}-book.pdf" ]; then
        mv "${BASENAME}-book.pdf" "${BASENAME}-book.pdf.tmp"
    fi

    docker run --rm -v "$(pwd)":/workdir -w /workdir texlive/texlive:latest \
        pdfbook2 --no-crop --paper=a4paper "$OUTPUT_PDF"

    # pdfbook2 creates the output with -book suffix, rename to -magazine
    mv "${BASENAME}-book.pdf" "${BOOKLET_PDF}"

    # Restore the original -book.pdf if it existed
    if [ -f "${BASENAME}-book.pdf.tmp" ]; then
        mv "${BASENAME}-book.pdf.tmp" "${BASENAME}-book.pdf"
    fi
fi

mv $OUTPUT_PDF .extra/ 2>/dev/null || true

echo ""
echo "✓ Done!"
echo "  Original PDF: .extra/$OUTPUT_PDF"
echo "  Booklet PDF:  $BOOKLET_PDF"
echo ""
echo "Printing instructions:"
echo "  1. Print the booklet PDF two-sided (flip on SHORT edge)"
echo "  2. All sheets should be printed in order"
echo "  3. Fold each sheet in half down the middle"
if [ "$MODE" = "magazine" ]; then
    echo "  4. Nest all sheets inside each other (like a magazine)"
else
    echo "  4. Stack all folded sheets on top of each other"
fi
echo "  5. The result will be an A5-sized booklet"
