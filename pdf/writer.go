package pdf

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Writer embeds ISDOC XML into PDF files using incremental updates.
// It creates PDF/A-3 compliant embedded files with proper metadata.
type Writer struct {
	// Filename is the name used for the embedded ISDOC file.
	// Default is "invoice.isdoc".
	Filename string
}

// NewWriter creates a new PDF writer with default settings.
func NewWriter() *Writer {
	return &Writer{
		Filename: "invoice.isdoc",
	}
}

// EmbedFile embeds ISDOC XML into an existing PDF file.
func (w *Writer) EmbedFile(inputPath, outputPath string, xml []byte) error {
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading input PDF: %w", err)
	}

	outputData, err := w.Embed(inputData, xml)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("writing output PDF: %w", err)
	}

	return nil
}

// Embed appends ISDOC XML as an embedded file to a PDF using incremental update.
// Creates PDF/A-3 compliant structure with:
// - EmbeddedFile stream with MIME type and timestamps
// - FileSpec dictionary with AFRelationship and UF (Unicode filename)
// - EmbeddedFiles name tree
// - AF (Associated Files) array in catalog
func (w *Writer) Embed(pdfData []byte, xml []byte) ([]byte, error) {
	startxref, err := w.findStartxref(pdfData)
	if err != nil {
		return nil, err
	}

	size, root, rootObjNum, err := w.parseTrailerWithRoot(pdfData, startxref)
	if err != nil {
		return nil, err
	}

	var out strings.Builder
	out.Write(pdfData)

	// Current timestamp for PDF date format
	now := time.Now()
	pdfDate := w.formatPDFDate(now)

	filename := w.Filename
	if filename == "" {
		filename = "invoice.isdoc"
	}

	xmlLen := len(xml)

	// Object positions for xref table
	objPositions := make(map[int]int)

	// Object 1: EmbeddedFile stream
	efObjNum := size
	objPositions[efObjNum] = out.Len()
	fmt.Fprintf(&out, "%d 0 obj\n", efObjNum)
	fmt.Fprintf(&out, "<< /Type /EmbeddedFile\n")
	fmt.Fprintf(&out, "   /Subtype /text#2Fxml\n")
	fmt.Fprintf(&out, "   /Params << /Size %d /CreationDate %s /ModDate %s >>\n", xmlLen, pdfDate, pdfDate)
	fmt.Fprintf(&out, "   /Length %d >>\n", xmlLen)
	out.WriteString("stream\n")
	out.Write(xml)
	out.WriteString("\nendstream\nendobj\n\n")

	// Object 2: FileSpec dictionary
	fsObjNum := size + 1
	objPositions[fsObjNum] = out.Len()
	fmt.Fprintf(&out, "%d 0 obj\n", fsObjNum)
	fmt.Fprintf(&out, "<< /Type /Filespec\n")
	fmt.Fprintf(&out, "   /F (%s)\n", filename)
	fmt.Fprintf(&out, "   /UF (%s)\n", filename)
	fmt.Fprintf(&out, "   /Desc (ISDOC Electronic Invoice)\n")
	fmt.Fprintf(&out, "   /AFRelationship /Alternative\n")
	fmt.Fprintf(&out, "   /EF << /F %d 0 R /UF %d 0 R >> >>\n", efObjNum, efObjNum)
	out.WriteString("endobj\n\n")

	// Object 3: EmbeddedFiles name tree
	namesObjNum := size + 2
	objPositions[namesObjNum] = out.Len()
	fmt.Fprintf(&out, "%d 0 obj\n", namesObjNum)
	fmt.Fprintf(&out, "<< /Names [(%s) %d 0 R] >>\n", filename, fsObjNum)
	out.WriteString("endobj\n\n")

	// Object 4: Names dictionary
	namesDictObjNum := size + 3
	objPositions[namesDictObjNum] = out.Len()
	fmt.Fprintf(&out, "%d 0 obj\n", namesDictObjNum)
	fmt.Fprintf(&out, "<< /EmbeddedFiles %d 0 R >>\n", namesObjNum)
	out.WriteString("endobj\n\n")

	// Object 5: Updated catalog with Names and AF
	newCatalogObjNum := size + 4
	objPositions[newCatalogObjNum] = out.Len()
	fmt.Fprintf(&out, "%d 0 obj\n", newCatalogObjNum)
	fmt.Fprintf(&out, "<< /Type /Catalog\n")
	fmt.Fprintf(&out, "   /Names %d 0 R\n", namesDictObjNum)
	fmt.Fprintf(&out, "   /AF [%d 0 R]\n", fsObjNum)
	fmt.Fprintf(&out, "   /BaseFrom %s >>\n", root)
	out.WriteString("endobj\n\n")

	// Write xref table
	newXRefPos := out.Len()
	newSize := size + 5
	fmt.Fprintf(&out, "xref\n")
	fmt.Fprintf(&out, "%d %d\n", size, 5)
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", objPositions[size+i])
	}

	// Also update the root catalog reference
	out.WriteString("trailer\n")
	fmt.Fprintf(&out, "<< /Size %d /Root %d 0 R /Prev %d >>\n", newSize, rootObjNum, startxref)
	out.WriteString("startxref\n")
	fmt.Fprintf(&out, "%d\n", newXRefPos)
	out.WriteString("%%EOF\n")

	return []byte(out.String()), nil
}

// formatPDFDate formats a time.Time as a PDF date string.
// Format: (D:YYYYMMDDHHmmSSOHH'mm')
func (w *Writer) formatPDFDate(t time.Time) string {
	_, offset := t.Zone()
	offsetHours := offset / 3600
	offsetMinutes := (offset % 3600) / 60

	sign := "+"
	if offsetHours < 0 {
		sign = "-"
		offsetHours = -offsetHours
		offsetMinutes = -offsetMinutes
	}

	return fmt.Sprintf("(D:%04d%02d%02d%02d%02d%02d%s%02d'%02d')",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
		sign, offsetHours, offsetMinutes)
}

// findStartxref locates the startxref value in the PDF.
func (w *Writer) findStartxref(data []byte) (int, error) {
	searchLen := 256
	if len(data) < searchLen {
		searchLen = len(data)
	}
	tail := string(data[len(data)-searchLen:])

	idx := strings.LastIndex(tail, "startxref\n")
	if idx == -1 {
		idx = strings.LastIndex(tail, "startxref\r\n")
		if idx == -1 {
			return 0, fmt.Errorf("startxref not found")
		}
		idx += 11
	} else {
		idx += 10
	}

	rest := tail[idx:]
	endIdx := strings.IndexAny(rest, "\r\n")
	if endIdx == -1 {
		return 0, fmt.Errorf("startxref value not found")
	}

	startxref, err := strconv.Atoi(strings.TrimSpace(rest[:endIdx]))
	if err != nil {
		return 0, fmt.Errorf("invalid startxref value: %w", err)
	}

	if startxref <= 0 || startxref >= len(data) {
		return 0, fmt.Errorf("startxref out of bounds: %d", startxref)
	}

	return startxref, nil
}

// parseTrailerWithRoot extracts Size, Root reference, and Root object number from the trailer.
func (w *Writer) parseTrailerWithRoot(data []byte, startxref int) (size int, root string, rootObjNum int, err error) {
	if startxref >= len(data) {
		return 0, "", 0, fmt.Errorf("startxref position out of bounds")
	}

	searchData := string(data[startxref:])
	trailerIdx := strings.Index(searchData, "trailer")
	if trailerIdx == -1 {
		return 0, "", 0, fmt.Errorf("trailer not found")
	}

	trailerSection := searchData[trailerIdx:]
	endIdx := strings.Index(trailerSection, "startxref")
	if endIdx == -1 {
		endIdx = len(trailerSection)
	}
	trailerSection = trailerSection[:endIdx]

	sizeRe := regexp.MustCompile(`/Size\s+(\d+)`)
	sizeMatch := sizeRe.FindStringSubmatch(trailerSection)
	if sizeMatch == nil {
		return 0, "", 0, fmt.Errorf("Size not found in trailer")
	}
	size, err = strconv.Atoi(sizeMatch[1])
	if err != nil {
		return 0, "", 0, fmt.Errorf("invalid Size value: %w", err)
	}

	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	rootMatch := rootRe.FindStringSubmatch(trailerSection)
	if rootMatch == nil {
		return 0, "", 0, fmt.Errorf("Root not found in trailer")
	}
	root = rootMatch[1] + " " + rootMatch[2] + " R"
	rootObjNum, err = strconv.Atoi(rootMatch[1])
	if err != nil {
		return 0, "", 0, fmt.Errorf("invalid Root object number: %w", err)
	}

	return size, root, rootObjNum, nil
}

// EmbedXML is a convenience function to embed ISDOC XML into a PDF file.
func EmbedXML(inputPath, outputPath string, xml []byte) error {
	w := NewWriter()
	return w.EmbedFile(inputPath, outputPath, xml)
}

// EmbedXMLToWriter embeds ISDOC XML into a PDF and writes to a writer.
func EmbedXMLToWriter(out io.Writer, pdfData []byte, xml []byte) error {
	writer := NewWriter()
	result, err := writer.Embed(pdfData, xml)
	if err != nil {
		return err
	}
	_, err = out.Write(result)
	return err
}
