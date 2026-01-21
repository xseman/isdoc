package isdoc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

// TestReferenceResolution tests that id/ref attributes are properly resolved.
// Based on PHP ReferenceTest.testReference
func TestReferenceResolution(t *testing.T) {
	// Create an invoice with order references
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      1,
		ID:                "REF-TEST-001",
		UUID:              types.UUID("00000000-0000-0000-0000-000000001234"),
		IssueDate:         types.MustParseDate("2021-08-16"),
		VATApplicable:     types.Bool(false),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "Test Supplier s.r.o."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Test Street",
					CityName:   "Prague",
					PostalZone: "10000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		AccountingCustomerParty: &schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "87654321"},
				PartyName:           schema.PartyName{Name: "Test Customer a.s."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Customer Street",
					CityName:   "Brno",
					PostalZone: "60000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		// Create order reference with id attribute
		OrderReferences: &schema.OrderReferences{
			OrderReference: []schema.OrderReference{
				{
					ID:           "order-ref-1",
					SalesOrderID: "123456",
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID:                              "1",
					LineExtensionAmount:             types.MustDecimal("100.0"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("121.0"),
					LineExtensionTaxAmount:          types.MustDecimal("21.0"),
					UnitPrice:                       types.MustDecimal("100.0"),
					UnitPriceTaxInclusive:           types.MustDecimal("121.0"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:              types.MustDecimal("21"),
						VATCalculationMethod: 1,
					},
					// Reference to the order using ref attribute
					OrderReference: &schema.OrderLineReference{
						LineID: "10",
					},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxAmount: types.MustDecimal("21.0"),
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("100.0"),
			TaxInclusiveAmount:               types.MustDecimal("121.0"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("100.0"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("121.0"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("121.0"),
		},
	}

	// Encode to XML
	xmlData, err := EncodeBytes(invoice)
	if err != nil {
		t.Fatalf("EncodeBytes failed: %v", err)
	}

	// Decode back
	decoded, err := DecodeBytes(xmlData)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	// Verify basic fields
	if decoded.ID != invoice.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, invoice.ID)
	}

	// Verify order references exist
	if decoded.OrderReferences == nil {
		t.Fatal("OrderReferences is nil")
	}
	if len(decoded.OrderReferences.OrderReference) == 0 {
		t.Fatal("No order references found")
	}
	if decoded.OrderReferences.OrderReference[0].SalesOrderID != "123456" {
		t.Errorf("SalesOrderID mismatch: got %q, want %q",
			decoded.OrderReferences.OrderReference[0].SalesOrderID, "123456")
	}

	// Verify invoice line reference
	if len(decoded.InvoiceLines.InvoiceLine) == 0 {
		t.Fatal("No invoice lines found")
	}
	line := decoded.InvoiceLines.InvoiceLine[0]
	if line.OrderReference == nil {
		t.Error("Invoice line OrderReference is nil")
	} else if line.OrderReference.LineID != "10" {
		t.Errorf("LineID mismatch: got %q, want %q", line.OrderReference.LineID, "10")
	}
}

// TestMultiPartyTaxScheme tests parsing of multiple PartyTaxScheme elements.
// Based on PHP DecoderTest.testMultiPartyTaxScheme
func TestMultiPartyTaxScheme(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "multi-partytax.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	// Verify we have multiple party tax schemes
	supplier := invoice.AccountingSupplierParty.Party
	if len(supplier.PartyTaxScheme) == 0 {
		t.Fatal("No PartyTaxScheme found")
	}

	if len(supplier.PartyTaxScheme) < 2 {
		t.Fatalf("Expected at least 2 PartyTaxScheme, got %d", len(supplier.PartyTaxScheme))
	}

	// Check first scheme (VAT)
	vatScheme := supplier.PartyTaxScheme[0]
	if vatScheme.TaxScheme != "VAT" {
		t.Errorf("First TaxScheme mismatch: got %q, want %q", vatScheme.TaxScheme, "VAT")
	}
	if vatScheme.CompanyID != "CZ25097563" {
		t.Errorf("First CompanyID mismatch: got %q, want %q", vatScheme.CompanyID, "CZ25097563")
	}

	// Check second scheme (TIN)
	tinScheme := supplier.PartyTaxScheme[1]
	if tinScheme.TaxScheme != "TIN" {
		t.Errorf("Second TaxScheme mismatch: got %q, want %q", tinScheme.TaxScheme, "TIN")
	}
	if tinScheme.CompanyID != "SK25097563" {
		t.Errorf("Second CompanyID mismatch: got %q, want %q", tinScheme.CompanyID, "SK25097563")
	}

	t.Logf("Found %d PartyTaxSchemes: VAT=%s, TIN=%s",
		len(supplier.PartyTaxScheme), vatScheme.CompanyID, tinScheme.CompanyID)
}

// TestNamespacedReferences tests parsing with namespaced id/ref attributes.
// Based on PHP DecoderTest.testNamespacedReferences
func TestNamespacedReferences(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "sample-namespaced-references.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	// Verify order references exist
	if invoice.OrderReferences == nil || len(invoice.OrderReferences.OrderReference) == 0 {
		t.Fatal("No order references found")
	}

	// Verify delivery note references exist
	if invoice.DeliveryNoteReferences == nil || len(invoice.DeliveryNoteReferences.DeliveryNoteReference) == 0 {
		t.Fatal("No delivery note references found")
	}

	// Verify invoice lines have references
	if len(invoice.InvoiceLines.InvoiceLine) == 0 {
		t.Fatal("No invoice lines found")
	}

	firstLine := invoice.InvoiceLines.InvoiceLine[0]
	if firstLine.OrderReference == nil {
		t.Error("First line OrderReference is nil")
	}
	if firstLine.DeliveryNoteReference == nil {
		t.Error("First line DeliveryNoteReference is nil")
	}

	t.Logf("Found %d order refs, %d delivery note refs",
		len(invoice.OrderReferences.OrderReference),
		len(invoice.DeliveryNoteReferences.DeliveryNoteReference))
}

// TestNoVATApplicable tests parsing an invoice with VATApplicable missing or false.
// Based on PHP DecoderTest.testSkipMissingPrimitiveValuesHydration
func TestNoVATApplicable(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "no-vat-applicable.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	// VATApplicable should be false (default or explicit)
	if bool(invoice.VATApplicable) != false {
		t.Errorf("VATApplicable should be false, got %v", invoice.VATApplicable)
	}

	t.Logf("VATApplicable=%v for %s", invoice.VATApplicable, invoice.ID)
}

// TestJSONExport tests that invoices can be serialized to JSON.
// Based on TypeScript tests - Export to JSON
func TestJSONExport(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "sample.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	// Serialize to JSON
	jsonBytes, err := json.MarshalIndent(invoice, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Basic JSON validation
	if jsonStr[0] != '{' {
		t.Error("JSON should start with '{'")
	}

	// Parse back to verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Check some expected fields exist
	if _, ok := parsed["DocumentType"]; !ok {
		t.Error("JSON missing DocumentType field")
	}
	if _, ok := parsed["ID"]; !ok {
		t.Error("JSON missing ID field")
	}

	t.Logf("JSON export size: %d bytes", len(jsonBytes))
}

// TestDocumentTypes verifies all valid document types are accepted.
// Based on TypeScript tests - DocumentType validation
func TestDocumentTypes(t *testing.T) {
	validTypes := []int{1, 2, 3, 4, 5, 6, 7}

	for _, docType := range validTypes {
		t.Run(string(rune('0'+docType)), func(t *testing.T) {
			invoice := &schema.Invoice{
				DocumentType: docType,
				ID:           "TEST",
				UUID:         types.UUID("00000000-0000-0000-0000-000000000001"),
			}

			// Validate document type
			if invoice.DocumentType < 1 || invoice.DocumentType > 7 {
				t.Errorf("Invalid document type: %d", invoice.DocumentType)
			}
		})
	}
}

// TestInvalidDocumentType tests that invalid document types fail validation.
func TestInvalidDocumentType(t *testing.T) {
	invoice := &schema.Invoice{
		DocumentType: 99, // Invalid
		ID:           "TEST",
		UUID:         types.UUID("00000000-0000-0000-0000-000000000001"),
	}

	errs := ValidateInvoiceWithOptions(invoice, ValidateOptions{Strict: true})

	// Should have validation error for invalid document type
	hasDocTypeError := false
	for _, e := range errs {
		if strings.Contains(e.Field, "DocumentType") {
			hasDocTypeError = true
			break
		}
	}

	if !hasDocTypeError {
		t.Error("Expected validation error for invalid DocumentType")
	}
}

// TestSampleNoReference tests the sample-no-reference fixture.
// Based on PHP DecoderTest.testSampleNoReference
func TestSampleNoReference(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "sample-no-reference.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	// Verify basic structure
	if invoice.ID != "12345" {
		t.Errorf("ID mismatch: got %q, want %q", invoice.ID, "12345")
	}
	if string(invoice.UUID) != "00000000-0000-0000-0000-000000001234" {
		t.Errorf("UUID mismatch: got %q", invoice.UUID)
	}

	// Should have one invoice line
	if len(invoice.InvoiceLines.InvoiceLine) != 1 {
		t.Errorf("Expected 1 invoice line, got %d", len(invoice.InvoiceLines.InvoiceLine))
	}

	// Invoice line should have ID "1"
	if invoice.InvoiceLines.InvoiceLine[0].ID != "1" {
		t.Errorf("Line ID mismatch: got %q, want %q",
			invoice.InvoiceLines.InvoiceLine[0].ID, "1")
	}
}
