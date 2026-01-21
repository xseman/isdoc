package isdoc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

func TestEncodeBasicInvoice(t *testing.T) {
	// Create a minimal invoice
	invoice := &schema.Invoice{
		DocumentType:      1,
		ID:                "TEST-001",
		UUID:              types.UUID("12345678-1234-1234-1234-123456789012"),
		IssueDate:         types.MustParseDate("2024-01-15"),
		VATApplicable:     types.Bool(true),
		LocalCurrencyCode: "CZK",
		CurrRate:          types.MustDecimal("1"),
		RefCurrRate:       types.MustDecimal("1"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{
					ID: "12345678",
				},
				PartyName: schema.PartyName{
					Name: "Test Supplier",
				},
				PostalAddress: schema.PostalAddress{
					StreetName: "Test Street",
					CityName:   "Prague",
					PostalZone: "11000",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		AccountingCustomerParty: &schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{
					ID: "87654321",
				},
				PartyName: schema.PartyName{
					Name: "Test Customer",
				},
				PostalAddress: schema.PostalAddress{
					StreetName: "Customer Street",
					CityName:   "Brno",
					PostalZone: "60200",
					Country:    schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID: "1",
					InvoicedQuantity: schema.Quantity{
						Value:    types.MustDecimal("1"),
						UnitCode: "C62",
					},
					LineExtensionAmount:             types.MustDecimal("100.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("121.00"),
					LineExtensionTaxAmount:          types.MustDecimal("21.00"),
					UnitPrice:                       types.MustDecimal("100.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("121.00"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:              types.MustDecimal("21"),
						VATCalculationMethod: 0,
					},
					Item: schema.Item{
						Description: "Test Item",
					},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxSubTotal: []schema.TaxSubTotal{
				{
					TaxableAmount:                    types.MustDecimal("100.00"),
					TaxAmount:                        types.MustDecimal("21.00"),
					TaxInclusiveAmount:               types.MustDecimal("121.00"),
					AlreadyClaimedTaxableAmount:      types.MustDecimal("0"),
					AlreadyClaimedTaxAmount:          types.MustDecimal("0"),
					AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
					DifferenceTaxableAmount:          types.MustDecimal("100.00"),
					DifferenceTaxAmount:              types.MustDecimal("21.00"),
					DifferenceTaxInclusiveAmount:     types.MustDecimal("121.00"),
					TaxCategory: schema.TaxCategory{
						Percent: types.MustDecimal("21"),
					},
				},
			},
			TaxAmount: types.MustDecimal("21.00"),
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount:               types.MustDecimal("100.00"),
			TaxInclusiveAmount:               types.MustDecimal("121.00"),
			AlreadyClaimedTaxExclusiveAmount: types.MustDecimal("0"),
			AlreadyClaimedTaxInclusiveAmount: types.MustDecimal("0"),
			DifferenceTaxExclusiveAmount:     types.MustDecimal("100.00"),
			DifferenceTaxInclusiveAmount:     types.MustDecimal("121.00"),
			PayableRoundingAmount:            types.MustDecimal("0"),
			PaidDepositsAmount:               types.MustDecimal("0"),
			PayableAmount:                    types.MustDecimal("121.00"),
		},
		PaymentMeans: &schema.PaymentMeans{
			Payment: []schema.Payment{
				{
					PaidAmount:       types.MustDecimal("121.00"),
					PaymentMeansCode: 42,
					Details: &schema.PaymentDetails{
						PaymentDueDate: types.MustParseDate("2024-02-15"),
						VariableSymbol: "123456",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(invoice)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Basic checks
	xmlStr := buf.String()
	if !strings.Contains(xmlStr, "<?xml version=") {
		t.Error("Missing XML declaration")
	}
	if !strings.Contains(xmlStr, "<Invoice") {
		t.Error("Missing Invoice element")
	}
	if !strings.Contains(xmlStr, "<DocumentType>1</DocumentType>") {
		t.Error("Missing DocumentType")
	}
	if !strings.Contains(xmlStr, "<ID>TEST-001</ID>") {
		t.Error("Missing ID")
	}
	if !strings.Contains(xmlStr, "<UUID>12345678-1234-1234-1234-123456789012</UUID>") {
		t.Error("Missing UUID")
	}
	if !strings.Contains(xmlStr, "Test Supplier") {
		t.Error("Missing supplier name")
	}
	if !strings.Contains(xmlStr, "Test Customer") {
		t.Error("Missing customer name")
	}

	t.Logf("Generated XML:\n%s", xmlStr)
}

func TestRoundTrip(t *testing.T) {
	fixtures, err := filepath.Glob("testdata/fixtures/*.isdoc")
	if err != nil {
		t.Fatalf("Failed to list fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		name := filepath.Base(fixture)

		// Skip CommonDocument fixtures in Invoice roundtrip test
		if strings.Contains(name, "commondocument") {
			continue
		}

		t.Run(name, func(t *testing.T) {
			// Read original
			original, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("Failed to read fixture: %v", err)
			}

			// Decode
			invoice, err := DecodeBytes(original)
			if err != nil {
				t.Fatalf("DecodeBytes failed: %v", err)
			}

			// Encode
			encoded, err := EncodeBytes(invoice)
			if err != nil {
				t.Fatalf("EncodeBytes failed: %v", err)
			}

			// Decode again
			invoice2, err := DecodeBytes(encoded)
			if err != nil {
				t.Fatalf("Re-parse failed: %v", err)
			}

			// Compare key fields
			if invoice.ID != invoice2.ID {
				t.Errorf("ID mismatch: %q vs %q", invoice.ID, invoice2.ID)
			}
			if invoice.UUID != invoice2.UUID {
				t.Errorf("UUID mismatch: %q vs %q", invoice.UUID, invoice2.UUID)
			}
			if invoice.DocumentType != invoice2.DocumentType {
				t.Errorf("DocumentType mismatch: %d vs %d", invoice.DocumentType, invoice2.DocumentType)
			}
			if invoice.IssueDate.String() != invoice2.IssueDate.String() {
				t.Errorf("IssueDate mismatch: %q vs %q", invoice.IssueDate, invoice2.IssueDate)
			}

			// Compare line counts (InvoiceLines is a value type, not pointer)
			if len(invoice.InvoiceLines.InvoiceLine) != len(invoice2.InvoiceLines.InvoiceLine) {
				t.Errorf("Line count mismatch: %d vs %d",
					len(invoice.InvoiceLines.InvoiceLine),
					len(invoice2.InvoiceLines.InvoiceLine))
			}

			t.Logf("Round-trip OK for %s", name)
		})
	}
}

func TestEncodeWithWriter(t *testing.T) {
	invoice := &schema.Invoice{
		DocumentType: 1,
		ID:           "WRITER-TEST",
		UUID:         types.UUID("12345678-1234-1234-1234-123456789012"),
	}

	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(invoice)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Output is empty")
	}

	xml := buf.String()
	if !strings.Contains(xml, "WRITER-TEST") {
		t.Error("Missing expected content")
	}
}

func TestEncodeWriter(t *testing.T) {
	invoice := &schema.Invoice{
		DocumentType: 1,
		ID:           "ENCODE-WRITER-TEST",
		UUID:         types.UUID("12345678-1234-1234-1234-123456789012"),
	}

	var buf bytes.Buffer
	encoder := NewEncoder(&buf)
	if err := encoder.Encode(invoice); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(buf.String(), "ENCODE-WRITER-TEST") {
		t.Error("Missing expected content")
	}
}
