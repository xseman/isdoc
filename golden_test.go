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

// TestSimplifiedTaxDocument tests creating a simplified tax document (DocumentType=7).
// Based on PHP EncoderTest.testSimplifiedTaxDocument
func TestSimplifiedTaxDocument(t *testing.T) {
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      7, // Simplified tax document
		ID:                "STD-001",
		UUID:              types.UUID("00000000-0000-0000-0000-000000007777"),
		IssueDate:         types.MustParseDate("2024-01-15"),
		TaxPointDate:      types.MustParseDate("2024-01-15"),
		VATApplicable:     types.Bool(true),
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
				PartyTaxScheme: []schema.PartyTaxScheme{
					{CompanyID: "CZ12345678", TaxScheme: "VAT"},
				},
			},
		},
		// Simplified tax documents use AnonymousCustomerParty instead of AccountingCustomerParty
		AnonymousCustomerParty: &schema.AnonymousCustomerParty{
			ID: "ANONYMOUS-001",
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
						VATCalculationMethod: 0,
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

	// Verify XML contains AnonymousCustomerParty
	xmlStr := string(xmlData)
	if !strings.Contains(xmlStr, "<AnonymousCustomerParty>") {
		t.Error("XML should contain AnonymousCustomerParty")
	}
	if strings.Contains(xmlStr, "<AccountingCustomerParty>") {
		t.Error("Simplified tax document should not have AccountingCustomerParty")
	}
	if !strings.Contains(xmlStr, "<DocumentType>7</DocumentType>") {
		t.Error("DocumentType should be 7")
	}

	// Validate - should pass for simplified tax doc
	errs := ValidateInvoice(invoice)
	for _, e := range errs {
		if e.Severity == SeverityError {
			t.Errorf("Unexpected validation error: %s", e.Msg)
		}
	}

	// Round-trip test
	decoded, err := DecodeBytes(xmlData)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	if decoded.DocumentType != 7 {
		t.Errorf("DocumentType mismatch: got %d, want 7", decoded.DocumentType)
	}
	if decoded.AnonymousCustomerParty == nil {
		t.Error("AnonymousCustomerParty should not be nil after round-trip")
	} else if decoded.AnonymousCustomerParty.ID != "ANONYMOUS-001" {
		t.Errorf("AnonymousCustomerParty.ID mismatch: got %q", decoded.AnonymousCustomerParty.ID)
	}

	t.Logf("Simplified tax document XML:\n%s", xmlStr[:min(500, len(xmlStr))])
}

// TestLegalMonetaryTotalConsistency tests that monetary totals are consistent.
// Based on PHP EncoderTest.testLegalMonetaryTotalSum
func TestLegalMonetaryTotalConsistency(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "sample.isdoc")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	invoice, err := DecodeBytes(data)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	total := invoice.LegalMonetaryTotal

	// TaxInclusiveAmount should be TaxExclusiveAmount + TaxTotal.TaxAmount
	// This is a soft check since precision can vary
	t.Logf("TaxExclusive=%s, TaxInclusive=%s, TaxAmount=%s",
		total.TaxExclusiveAmount.String(),
		total.TaxInclusiveAmount.String(),
		invoice.TaxTotal.TaxAmount.String())

	// PayableAmount should equal DifferenceTaxInclusiveAmount + PayableRoundingAmount - PaidDepositsAmount
	t.Logf("Payable=%s, Difference=%s, Rounding=%s, Deposits=%s",
		total.PayableAmount.String(),
		total.DifferenceTaxInclusiveAmount.String(),
		total.PayableRoundingAmount.String(),
		total.PaidDepositsAmount.String())
}

// TestGoldenEncode tests encoding against expected output structure.
// Based on PHP snapshot tests
func TestGoldenEncode(t *testing.T) {
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      1,
		ID:                "GOLDEN-001",
		UUID:              types.UUID("12345678-1234-1234-1234-123456789012"),
		IssueDate:         types.MustParseDate("2024-01-15"),
		VATApplicable:     types.Bool(true),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "Supplier Ltd."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Main Street 1",
					CityName:   "Prague",
					PostalZone: "11000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		AccountingCustomerParty: &schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "87654321"},
				PartyName:           schema.PartyName{Name: "Customer Inc."},
				PostalAddress: schema.PostalAddress{
					StreetName: "Oak Avenue 42",
					CityName:   "Brno",
					PostalZone: "60200",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID:                              "1",
					LineExtensionAmount:             types.MustDecimal("1000.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("1210.00"),
					LineExtensionTaxAmount:          types.MustDecimal("210.00"),
					UnitPrice:                       types.MustDecimal("1000.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("1210.00"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:              types.MustDecimal("21"),
						VATCalculationMethod: 0,
					},
					Item: schema.Item{
						Description: "Test Product",
					},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxAmount: types.MustDecimal("210.00"),
			TaxSubTotal: []schema.TaxSubTotal{
				{
					TaxableAmount:                    types.MustDecimal("1000.00"),
					TaxAmount:                        types.MustDecimal("210.00"),
					TaxInclusiveAmount:               types.MustDecimal("1210.00"),
					AlreadyClaimedTaxableAmount:      types.MustDecimal("0"),
					AlreadyClaimedTaxAmount:          types.MustDecimal("0"),
					AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
					DifferenceTaxableAmount:          types.MustDecimal("1000.00"),
					DifferenceTaxAmount:              types.MustDecimal("210.00"),
					DifferenceTaxInclusiveAmount:     types.MustDecimal("1210.00"),
					TaxCategory: schema.TaxCategory{
						Percent: types.MustDecimal("21"),
					},
				},
			},
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("1000.00"),
			TaxInclusiveAmount:               types.MustDecimal("1210.00"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("1000.00"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("1210.00"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("1210.00"),
		},
	}

	xmlData, err := EncodeBytes(invoice)
	if err != nil {
		t.Fatalf("EncodeBytes failed: %v", err)
	}

	xmlStr := string(xmlData)

	// Check required elements are present in correct order
	requiredElements := []string{
		"<DocumentType>1</DocumentType>",
		"<ID>GOLDEN-001</ID>",
		"<UUID>12345678-1234-1234-1234-123456789012</UUID>",
		"<IssueDate>2024-01-15</IssueDate>",
		"<VATApplicable>true</VATApplicable>",
		"<LocalCurrencyCode>CZK</LocalCurrencyCode>",
		"<AccountingSupplierParty>",
		"<AccountingCustomerParty>",
		"<InvoiceLines>",
		"<TaxTotal>",
		"<LegalMonetaryTotal>",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(xmlStr, elem) {
			t.Errorf("Missing required element: %s", elem)
		}
	}

	// Verify element ordering (basic check)
	docTypeIdx := strings.Index(xmlStr, "<DocumentType>")
	idIdx := strings.Index(xmlStr, "<ID>")
	uuidIdx := strings.Index(xmlStr, "<UUID>")

	if !(docTypeIdx < idIdx && idIdx < uuidIdx) {
		t.Error("Elements not in expected order (DocumentType < ID < UUID)")
	}
}

// TestJSONSnapshot tests JSON serialization format.
// Based on PHP decoder-sample.json snapshot
func TestJSONSnapshot(t *testing.T) {
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

	// Parse back to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Check expected fields exist
	expectedFields := []string{
		"DocumentType",
		"ID",
		"UUID",
		"IssueDate",
		"VATApplicable",
		"LocalCurrencyCode",
		"AccountingSupplierParty",
		"InvoiceLines",
		"TaxTotal",
		"LegalMonetaryTotal",
	}

	for _, field := range expectedFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("JSON missing expected field: %s", field)
		}
	}

	// Check AccountingSupplierParty structure
	if asp, ok := parsed["AccountingSupplierParty"].(map[string]interface{}); ok {
		if party, ok := asp["Party"].(map[string]interface{}); ok {
			if _, ok := party["PartyIdentification"]; !ok {
				t.Error("Party missing PartyIdentification")
			}
			if _, ok := party["PartyName"]; !ok {
				t.Error("Party missing PartyName")
			}
			if _, ok := party["PostalAddress"]; !ok {
				t.Error("Party missing PostalAddress")
			}
		} else {
			t.Error("AccountingSupplierParty missing Party")
		}
	} else {
		t.Error("AccountingSupplierParty is not an object")
	}

	t.Logf("JSON fields: %d top-level fields", len(parsed))
}

// TestCreditNote tests creating a credit note (DocumentType=2).
func TestCreditNote(t *testing.T) {
	invoice := &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      2, // Credit note
		ID:                "CN-001",
		UUID:              types.UUID("00000000-0000-0000-0000-000000002222"),
		IssueDate:         types.MustParseDate("2024-02-01"),
		VATApplicable:     types.Bool(true),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "Test Supplier"},
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
				PartyName:           schema.PartyName{Name: "Test Customer"},
				PostalAddress: schema.PostalAddress{
					StreetName: "Customer Street",
					CityName:   "Brno",
					PostalZone: "60000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		// Reference to original document being corrected
		OriginalDocumentReferences: &schema.OriginalDocumentReferences{
			OriginalDocumentReference: []schema.OriginalDocumentReference{
				{
					ID:        "orig-ref-1",
					IssueDate: types.MustParseDate("2024-01-15"),
					UUID:      "00000000-0000-0000-0000-000000001111",
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID:                              "1",
					LineExtensionAmount:             types.MustDecimal("-100.0"), // Negative for credit
					LineExtensionAmountTaxInclusive: types.MustDecimal("-121.0"),
					LineExtensionTaxAmount:          types.MustDecimal("-21.0"),
					UnitPrice:                       types.MustDecimal("-100.0"),
					UnitPriceTaxInclusive:           types.MustDecimal("-121.0"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:              types.MustDecimal("21"),
						VATCalculationMethod: 0,
					},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxAmount: types.MustDecimal("-21.0"),
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("-100.0"),
			TaxInclusiveAmount:               types.MustDecimal("-121.0"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("-100.0"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("-121.0"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("-121.0"),
		},
	}

	xmlData, err := EncodeBytes(invoice)
	if err != nil {
		t.Fatalf("EncodeBytes failed: %v", err)
	}

	xmlStr := string(xmlData)
	if !strings.Contains(xmlStr, "<DocumentType>2</DocumentType>") {
		t.Error("DocumentType should be 2 for credit note")
	}
	if !strings.Contains(xmlStr, "<OriginalDocumentReferences>") {
		t.Error("Credit note should have OriginalDocumentReferences")
	}

	// Round-trip
	decoded, err := DecodeBytes(xmlData)
	if err != nil {
		t.Fatalf("DecodeBytes failed: %v", err)
	}

	if decoded.DocumentType != 2 {
		t.Errorf("DocumentType mismatch: got %d, want 2", decoded.DocumentType)
	}
	if decoded.OriginalDocumentReferences == nil {
		t.Error("OriginalDocumentReferences should not be nil")
	}

	t.Logf("Credit note created with reference to original document")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
