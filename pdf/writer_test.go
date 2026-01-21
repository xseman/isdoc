package pdf

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFormatPDFDate(t *testing.T) {
	w := NewWriter()

	// Test with a fixed time
	loc, _ := time.LoadLocation("Europe/Prague")
	testTime := time.Date(2025, 1, 19, 15, 30, 45, 0, loc)

	result := w.formatPDFDate(testTime)

	// Should start with (D: and contain the date
	if !strings.HasPrefix(result, "(D:2025011915") {
		t.Errorf("unexpected date format: %s", result)
	}

	// Should end with closing paren
	if !strings.HasSuffix(result, "')") {
		t.Errorf("date should end with '): %s", result)
	}
}

func TestWriterEmbed(t *testing.T) {
	w := NewWriter()

	// Minimal valid PDF
	minimalPDF := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [] /Count 0 >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Size 3 /Root 1 0 R >>
startxref
112
%%EOF
`)

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="http://isdoc.cz/namespace/2013" version="6.0.2">
<ID>TEST-001</ID>
</Invoice>`)

	result, err := w.Embed(minimalPDF, xml)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	resultStr := string(result)

	// Check for PDF/A-3 required elements
	tests := []struct {
		name    string
		content string
	}{
		{"EmbeddedFile type", "/Type /EmbeddedFile"},
		{"MIME type", "/Subtype /text#2Fxml"},
		{"FileSpec type", "/Type /Filespec"},
		{"Filename", "/F (invoice.isdoc)"},
		{"Unicode filename", "/UF (invoice.isdoc)"},
		{"AFRelationship", "/AFRelationship /Alternative"},
		{"EmbeddedFiles name tree", "/EmbeddedFiles"},
		{"AF array", "/AF ["},
		{"Creation date", "/CreationDate (D:"},
		{"Mod date", "/ModDate (D:"},
		{"Size param", "/Size"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(resultStr, tc.content) {
				t.Errorf("missing %s in output", tc.content)
			}
		})
	}

	// Verify the XML content is embedded
	if !bytes.Contains(result, xml) {
		t.Error("XML content not found in output")
	}

	// Verify PDF structure ends with EOF marker
	if !strings.HasSuffix(resultStr, "%%EOF\n") {
		t.Fatal("PDF should end with EOF marker")
	}

	// Count xref entries (should have trailer)
	if !strings.Contains(resultStr, "trailer") {
		t.Error("missing trailer")
	}
}

func TestWriterCustomFilename(t *testing.T) {
	w := NewWriter()
	w.Filename = "faktura.isdoc"

	minimalPDF := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [] /Count 0 >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Size 3 /Root 1 0 R >>
startxref
112
%%EOF
`)

	xml := []byte(`<Invoice/>`)

	result, err := w.Embed(minimalPDF, xml)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	resultStr := string(result)

	if !strings.Contains(resultStr, "/F (faktura.isdoc)") {
		t.Error("custom filename not used")
	}

	if !strings.Contains(resultStr, "/UF (faktura.isdoc)") {
		t.Error("custom unicode filename not used")
	}
}

func TestFindStartxref(t *testing.T) {
	w := NewWriter()

	tests := []struct {
		name     string
		pdf      string
		expected int
		wantErr  bool
	}{
		{
			name:     "valid startxref",
			pdf:      "0123456789" + strings.Repeat("x", 90) + "\nstartxref\n10\n%%EOF\n",
			expected: 10,
		},
		{
			name:    "missing startxref",
			pdf:     "content\n%%EOF\n",
			wantErr: true,
		},
		{
			name:    "startxref out of bounds",
			pdf:     "short\nstartxref\n9999\n%%EOF\n",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := w.findStartxref([]byte(tc.pdf))

			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tc.expected {
				t.Errorf("got %d, want %d", result, tc.expected)
			}
		})
	}
}

func TestParseTrailerWithRoot(t *testing.T) {
	w := NewWriter()

	// The startxref value points to the beginning of the xref table
	// We need to search from that position to find the trailer
	pdf := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog >>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<< /Size 2 /Root 1 0 R >>
startxref
49
%%EOF
`)
	// startxref 49 points to the "xref" at position 49
	// Verify: "xref" starts at position 49 in this PDF

	size, root, rootNum, err := w.parseTrailerWithRoot(pdf, 49)
	if err != nil {
		t.Fatalf("parseTrailerWithRoot failed: %v", err)
	}

	if size != 2 {
		t.Errorf("size = %d, want 2", size)
	}

	if root != "1 0 R" {
		t.Errorf("root = %s, want '1 0 R'", root)
	}

	if rootNum != 1 {
		t.Errorf("rootNum = %d, want 1", rootNum)
	}
}
