package archive

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"testing"
)

func TestNewArchive(t *testing.T) {
	isdocData := []byte(`<?xml version="1.0"?><Invoice/>`)
	arc := NewArchive(isdocData, "test.isdoc")

	if arc.MainDocumentPath != "test.isdoc" {
		t.Errorf("Expected path test.isdoc, got %s", arc.MainDocumentPath)
	}
	if !bytes.Equal(arc.MainDocumentData, isdocData) {
		t.Error("Document data mismatch")
	}
}

func TestNewArchiveDefaultFilename(t *testing.T) {
	arc := NewArchive([]byte("<Invoice/>"), "")
	if arc.MainDocumentPath != "invoice.isdoc" {
		t.Errorf("Expected default path invoice.isdoc, got %s", arc.MainDocumentPath)
	}
}

func TestAddAttachment(t *testing.T) {
	arc := NewArchive([]byte("<Invoice/>"), "test.isdoc")
	arc.AddAttachment("attachment.pdf", []byte("PDF content"))

	if len(arc.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(arc.Attachments))
	}
	if _, ok := arc.Attachments["attachment.pdf"]; !ok {
		t.Error("Attachment not found")
	}
}

func TestWriteAndRead(t *testing.T) {
	// Create archive
	isdocData := []byte(`<?xml version="1.0"?><Invoice version="6.0.2"/>`)
	arc := NewArchive(isdocData, "faktura.isdoc")
	arc.AddAttachment("priloha.pdf", []byte("PDF attachment"))

	// Write to bytes
	data, err := arc.WriteBytes()
	if err != nil {
		t.Fatalf("WriteBytes failed: %v", err)
	}

	// Read back
	arc2, err := ReadBytes(data)
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	// Verify manifest
	if arc2.Manifest == nil {
		t.Error("Expected manifest to be present")
	}
	if arc2.Manifest.MainDocument.Filename != "faktura.isdoc" {
		t.Errorf("Expected filename faktura.isdoc, got %s", arc2.Manifest.MainDocument.Filename)
	}

	// Verify main document
	if arc2.MainDocumentPath != "faktura.isdoc" {
		t.Errorf("Expected path faktura.isdoc, got %s", arc2.MainDocumentPath)
	}
	if !bytes.Equal(arc2.MainDocumentData, isdocData) {
		t.Error("Document data mismatch")
	}

	// Verify attachment
	if len(arc2.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(arc2.Attachments))
	}
	if !bytes.Equal(arc2.Attachments["priloha.pdf"], []byte("PDF attachment")) {
		t.Error("Attachment data mismatch")
	}
}

func TestReadLegacyArchive(t *testing.T) {
	// Create a legacy archive without manifest.xml
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Write ISDOC file directly (no manifest)
	w, _ := zw.Create("invoice.isdoc")
	w.Write([]byte(`<?xml version="1.0"?><Invoice/>`))
	zw.Close()

	// Read
	arc, err := ReadBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	// Should work without manifest
	if arc.Manifest != nil {
		t.Error("Expected no manifest for legacy archive")
	}
	if arc.MainDocumentPath != "invoice.isdoc" {
		t.Errorf("Expected path invoice.isdoc, got %s", arc.MainDocumentPath)
	}
}

func TestReadNoISDOCDocument(t *testing.T) {
	// Create archive with no .isdoc file
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("readme.txt")
	w.Write([]byte("No invoice here"))
	zw.Close()

	_, err := ReadBytes(buf.Bytes())
	if err != ErrNoISDOCDocument {
		t.Errorf("Expected ErrNoISDOCDocument, got %v", err)
	}
}

func TestReadInvalidManifest(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Invalid manifest XML
	w, _ := zw.Create("manifest.xml")
	w.Write([]byte("not valid xml"))

	w2, _ := zw.Create("invoice.isdoc")
	w2.Write([]byte("<Invoice/>"))
	zw.Close()

	_, err := ReadBytes(buf.Bytes())
	if err == nil {
		t.Error("Expected error for invalid manifest")
	}
}

func TestManifestXMLStructure(t *testing.T) {
	manifest := &Manifest{
		XMLName: xml.Name{
			Space: ManifestNamespace,
			Local: "manifest",
		},
		MainDocument: MainDocument{
			Filename: "faktura.isdoc",
		},
	}

	data, err := xml.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify structure
	if !bytes.Contains(data, []byte("manifest")) {
		t.Error("Expected manifest element")
	}
	if !bytes.Contains(data, []byte("maindocument")) {
		t.Error("Expected maindocument element")
	}
	if !bytes.Contains(data, []byte(`filename="faktura.isdoc"`)) {
		t.Error("Expected filename attribute")
	}
}

func TestReadArchiveWithSubdirectory(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Manifest pointing to subdirectory
	manifest := `<?xml version="1.0"?>
<manifest xmlns="http://isdoc.cz/namespace/2013/manifest">
  <maindocument filename="documents/invoice.isdoc"/>
</manifest>`
	w, _ := zw.Create("manifest.xml")
	w.Write([]byte(manifest))

	// ISDOC in subdirectory
	w2, _ := zw.Create("documents/invoice.isdoc")
	w2.Write([]byte(`<?xml version="1.0"?><Invoice/>`))

	// Attachment in subdirectory
	w3, _ := zw.Create("attachments/scan.pdf")
	w3.Write([]byte("PDF"))
	zw.Close()

	arc, err := ReadBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	if arc.MainDocumentPath != "documents/invoice.isdoc" {
		t.Errorf("Expected path documents/invoice.isdoc, got %s", arc.MainDocumentPath)
	}
	if _, ok := arc.Attachments["attachments/scan.pdf"]; !ok {
		t.Error("Expected attachment in subdirectory")
	}
}
