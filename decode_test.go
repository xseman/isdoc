package isdoc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeFixtures(t *testing.T) {
	fixtures := []string{
		"sample.isdoc",
		"sample-no-reference.isdoc",
		"sample-namespaced-references.isdoc",
		"multi-partytax.isdoc",
		"no-vat-applicable.isdoc",
		"test001.isdoc",
		"test002.isdoc",
	}

	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", "fixtures", name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read fixture: %v", err)
			}

			invoice, err := DecodeBytes(data)
			if err != nil {
				if decErrs, ok := err.(DecodeErrors); ok {
					t.Logf("Decode warnings: %v", decErrs)
				} else {
					t.Fatalf("Failed to decode: %v", err)
				}
			}

			if invoice == nil {
				t.Fatal("Invoice is nil")
			}
			if invoice.ID == "" {
				t.Error("Invoice.ID is empty")
			}
			if invoice.UUID == "" {
				t.Error("Invoice.UUID is empty")
			}
			if invoice.DocumentType == 0 {
				t.Error("Invoice.DocumentType is 0")
			}
			if invoice.IssueDate.IsZero() {
				t.Error("Invoice.IssueDate is zero")
			}
			if len(invoice.InvoiceLines.InvoiceLine) == 0 {
				t.Error("Invoice has no lines")
			}

			t.Logf("Decoded invoice %s: ID=%s, UUID=%s, Type=%d, Lines=%d",
				name,
				invoice.ID,
				invoice.UUID,
				invoice.DocumentType,
				len(invoice.InvoiceLines.InvoiceLine),
			)
		})
	}
}

func TestDecodeInvalidXML(t *testing.T) {
	data := []byte(`<Invoice>not valid xml`)
	_, err := DecodeBytes(data)
	if err == nil {
		t.Error("Expected error for invalid XML")
	}
}

func TestDecodeEmptyXML(t *testing.T) {
	data := []byte(``)
	_, err := DecodeBytes(data)
	if err == nil {
		t.Error("Expected error for empty XML")
	}
}
