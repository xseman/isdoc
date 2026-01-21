package isdoc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xseman/isdoc/schema"
)

func TestValidateFixtures(t *testing.T) {
	fixtures := []struct {
		name        string
		expectError bool
	}{
		{"sample.isdoc", false},
		{"sample-no-reference.isdoc", true}, // Minimal test file missing required fields
		{"multi-partytax.isdoc", false},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join("testdata", "fixtures", tc.name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read fixture: %v", err)
			}

			invoice, err := DecodeBytes(data)
			if err != nil {
				if _, ok := err.(DecodeErrors); !ok {
					t.Fatalf("Failed to decode: %v", err)
				}
			}

			errs := ValidateInvoice(invoice)
			if tc.expectError {
				if !errs.HasErrors() && !errs.HasWarnings() {
					t.Error("Expected validation issues for this fixture")
				}
				t.Logf("Expected issues: %v", errs)
			} else {
				if errs.HasErrors() {
					t.Errorf("Validation errors: %v", errs.Errors())
				}
				if errs.HasWarnings() {
					t.Logf("Validation warnings: %v", errs.Warnings())
				}
			}
		})
	}
}

func TestValidateStrict(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "sample.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		if _, ok := err.(DecodeErrors); !ok {
			t.Fatalf("Failed to decode: %v", err)
		}
	}

	opts := ValidateOptions{
		Strict: true,
	}
	errs := ValidateInvoiceWithOptions(invoice, opts)

	// Log all issues in strict mode
	for _, e := range errs {
		t.Logf("[%s] %s: %s", e.Severity, e.Field, e.Msg)
	}
}

func TestValidateMissingRequiredFields(t *testing.T) {
	// Create an empty invoice
	inv := &schema.Invoice{
		DocumentType: 0, // Invalid
	}

	errs := ValidateInvoice(inv)
	if !errs.HasErrors() {
		t.Error("Expected errors for empty invoice")
	}

	// Check specific required field errors
	foundDocType := false
	foundID := false
	foundUUID := false
	for _, e := range errs {
		switch e.Field {
		case "Invoice.DocumentType":
			foundDocType = true
		case "Invoice.ID":
			foundID = true
		case "Invoice.UUID":
			foundUUID = true
		}
	}

	if !foundDocType {
		t.Error("Expected error for invalid DocumentType")
	}
	if !foundID {
		t.Error("Expected error for missing ID")
	}
	if !foundUUID {
		t.Error("Expected error for missing UUID")
	}
}
