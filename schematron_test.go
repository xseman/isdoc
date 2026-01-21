package isdoc

import (
	"testing"

	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

// TestSchematronOriginalDocumentLink tests DocumentType 2,3,6 requiring OriginalDocumentReferences.
// Based on Schematron rule: "Vazba na původní doklad"
func TestSchematronOriginalDocumentLink(t *testing.T) {
	tests := []struct {
		name          string
		documentType  int
		hasReferences bool
		expectError   bool
	}{
		{"Invoice (1) without references - OK", 1, false, false},
		{"Invoice (1) with references - OK", 1, true, false},
		{"Credit note (2) without references - Error", 2, false, true},
		{"Credit note (2) with references - OK", 2, true, false},
		{"Debit note (3) without references - Error", 3, false, true},
		{"Debit note (3) with references - OK", 3, true, false},
		{"Simplified tax doc (4) without references - OK", 4, false, false},
		{"Simplified credit (5) without references - OK", 5, false, false},
		{"Credit of advance (6) without references - Error", 6, false, true},
		{"Credit of advance (6) with references - OK", 6, true, false},
		{"Advance invoice (7) without references - OK", 7, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := createValidInvoice()
			inv.DocumentType = tt.documentType

			if tt.hasReferences {
				inv.OriginalDocumentReferences = &schema.OriginalDocumentReferences{
					OriginalDocumentReference: []schema.OriginalDocumentReference{
						{ID: "ORG-001"},
					},
				}
			} else {
				inv.OriginalDocumentReferences = nil
			}

			errs := validateOriginalDocumentLink(inv, DefaultValidateOptions())

			if tt.expectError && len(errs) == 0 {
				t.Errorf("Expected error for DocumentType %d without references", tt.documentType)
			}
			if !tt.expectError && len(errs) > 0 {
				t.Errorf("Unexpected error: %v", errs)
			}
		})
	}
}

// TestSchematronForeignCurrencyConsistency tests that ForeignCurrencyCode requires *Curr fields.
// Based on Schematron rule: "Konzistentní uvádění cizí měny"
func TestSchematronForeignCurrencyConsistency(t *testing.T) {
	t.Run("Foreign currency set - requires Curr fields", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "CZK"
		inv.ForeignCurrencyCode = "EUR"
		inv.CurrRate = types.MustDecimal("25.50")
		inv.RefCurrRate = types.MustDecimal("1")

		// No *Curr fields set
		errs := validateForeignCurrencyFieldsPresent(inv, ValidateOptions{Strict: true})

		if len(errs) == 0 {
			t.Error("Expected errors for missing *Curr fields")
		}

		// Check specific fields are flagged
		hasPayableAmountCurrErr := false
		for _, err := range errs {
			if err.Field == "Invoice.LegalMonetaryTotal.PayableAmountCurr" {
				hasPayableAmountCurrErr = true
			}
		}
		if !hasPayableAmountCurrErr {
			t.Error("Expected PayableAmountCurr error")
		}
	})

	t.Run("Foreign currency with all Curr fields - OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "CZK"
		inv.ForeignCurrencyCode = "EUR"
		inv.CurrRate = types.MustDecimal("25.50")
		inv.RefCurrRate = types.MustDecimal("1")

		// Set all required *Curr fields
		inv.LegalMonetaryTotal.TaxExclusiveAmountCurr = types.MustDecimal("100.00")
		inv.LegalMonetaryTotal.TaxInclusiveAmountCurr = types.MustDecimal("121.00")
		inv.LegalMonetaryTotal.PayableAmountCurr = types.MustDecimal("121.00")
		inv.TaxTotal.TaxAmountCurr = types.MustDecimal("21.00")
		// Set TaxSubTotal currency fields
		for i := range inv.TaxTotal.TaxSubTotal {
			inv.TaxTotal.TaxSubTotal[i].TaxableAmountCurr = types.MustDecimal("100.00")
			inv.TaxTotal.TaxSubTotal[i].TaxAmountCurr = types.MustDecimal("21.00")
			inv.TaxTotal.TaxSubTotal[i].TaxInclusiveAmountCurr = types.MustDecimal("121.00")
		}
		inv.InvoiceLines.InvoiceLine[0].LineExtensionAmountCurr = types.MustDecimal("100.00")
		inv.InvoiceLines.InvoiceLine[0].LineExtensionAmountTaxInclusiveCurr = types.MustDecimal("121.00")

		errs := validateForeignCurrencyFieldsPresent(inv, ValidateOptions{Strict: true})

		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %v", errs)
		}
	})
}

// TestSchematronDomesticCurrencyConsistency tests that no ForeignCurrencyCode means no *Curr fields.
// Based on Schematron rule: "Konzistentní uvádění tuzemské měny"
func TestSchematronDomesticCurrencyConsistency(t *testing.T) {
	t.Run("No foreign currency - no Curr fields allowed", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "CZK"
		inv.ForeignCurrencyCode = "" // Domestic only

		// Incorrectly set *Curr fields
		inv.LegalMonetaryTotal.PayableAmountCurr = types.MustDecimal("100.00")

		errs := validateNoForeignCurrencyFields(inv, ValidateOptions{Strict: true})

		if len(errs) == 0 {
			t.Error("Expected error for *Curr field without ForeignCurrencyCode")
		}
	})

	t.Run("No foreign currency - CurrRate must be 1", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "CZK"
		inv.ForeignCurrencyCode = ""
		inv.CurrRate = types.MustDecimal("25.50") // Should be 1

		errs := validateCurrencyConsistency(inv, ValidateOptions{Strict: true})

		hasCurrRateErr := false
		for _, err := range errs {
			if err.Field == "Invoice.CurrRate" {
				hasCurrRateErr = true
			}
		}
		if !hasCurrRateErr {
			t.Error("Expected CurrRate error when not 1 without foreign currency")
		}
	})

	t.Run("No foreign currency - CurrRate=1 is OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "CZK"
		inv.ForeignCurrencyCode = ""
		inv.CurrRate = types.MustDecimal("1")
		inv.RefCurrRate = types.MustDecimal("1")

		errs := validateCurrencyConsistency(inv, ValidateOptions{Strict: true})

		// Filter to CurrRate errors only
		for _, err := range errs {
			if err.Field == "Invoice.CurrRate" || err.Field == "Invoice.RefCurrRate" {
				t.Errorf("Unexpected currency rate error: %v", err)
			}
		}
	})
}

// TestSchematronCurrencyMismatch tests that foreign and local currencies must differ.
// Based on Schematron rule: "Tuzemská a zahraniční měna musí být rozdílná"
func TestSchematronCurrencyMismatch(t *testing.T) {
	t.Run("Same currency codes - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "EUR"
		inv.ForeignCurrencyCode = "EUR" // Same!

		errs := validateCurrencyConsistency(inv, DefaultValidateOptions())

		hasMismatchErr := false
		for _, err := range errs {
			if err.Field == "Invoice.ForeignCurrencyCode" && err.Msg == "ForeignCurrencyCode must differ from LocalCurrencyCode" {
				hasMismatchErr = true
			}
		}
		if !hasMismatchErr {
			t.Error("Expected error for same currency codes")
		}
	})

	t.Run("Different currency codes - OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.LocalCurrencyCode = "CZK"
		inv.ForeignCurrencyCode = "EUR"

		errs := validateCurrencyConsistency(inv, DefaultValidateOptions())

		for _, err := range errs {
			if err.Msg == "ForeignCurrencyCode must differ from LocalCurrencyCode" {
				t.Errorf("Unexpected currency mismatch error")
			}
		}
	})
}

// TestSchematronVATConsistency tests VAT applicability consistency.
// Based on Schematron rule: "Nedaňový doklad nesmí obsahovat řádkové položky podléhající DPH"
func TestSchematronVATConsistency(t *testing.T) {
	t.Run("Non-VAT invoice with VAT line - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.VATApplicable = types.Bool(false)
		inv.InvoiceLines.InvoiceLine[0].ClassifiedTaxCategory.VATApplicable = types.Bool(true)

		errs := validateVATConsistency(inv, DefaultValidateOptions())

		if len(errs) == 0 {
			t.Error("Expected error for VAT line in non-VAT invoice")
		}
	})

	t.Run("Non-VAT invoice with non-VAT line - OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.VATApplicable = types.Bool(false)
		inv.InvoiceLines.InvoiceLine[0].ClassifiedTaxCategory.VATApplicable = types.Bool(false)

		errs := validateVATConsistency(inv, DefaultValidateOptions())

		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %v", errs)
		}
	})

	t.Run("VAT invoice with VAT line - OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.VATApplicable = types.Bool(true)
		inv.InvoiceLines.InvoiceLine[0].ClassifiedTaxCategory.VATApplicable = types.Bool(true)

		errs := validateVATConsistency(inv, DefaultValidateOptions())

		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %v", errs)
		}
	})
}

// TestSchematronItemIdentificationHierarchy tests item identification requirements.
// Based on Schematron rules for identification hierarchy
func TestSchematronItemIdentificationHierarchy(t *testing.T) {
	t.Run("Tertiary without Secondary - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].Item.SellersItemIdentification = &schema.ItemIdentification{ID: "PRIMARY"}
		inv.InvoiceLines.InvoiceLine[0].Item.TertiarySellersItemIdentification = &schema.ItemIdentification{ID: "TERTIARY"}
		// Missing SecondarySellersItemIdentification

		errs := validateItemIdentificationHierarchy(inv, ValidateOptions{Strict: true})

		if len(errs) == 0 {
			t.Error("Expected error for tertiary without secondary")
		}
	})

	t.Run("Secondary without Primary - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].Item.SecondarySellersItemIdentification = &schema.ItemIdentification{ID: "SECONDARY"}
		// Missing SellersItemIdentification

		errs := validateItemIdentificationHierarchy(inv, ValidateOptions{Strict: true})

		if len(errs) == 0 {
			t.Error("Expected error for secondary without primary")
		}
	})

	t.Run("Full hierarchy - OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].Item.SellersItemIdentification = &schema.ItemIdentification{ID: "PRIMARY"}
		inv.InvoiceLines.InvoiceLine[0].Item.SecondarySellersItemIdentification = &schema.ItemIdentification{ID: "SECONDARY"}
		inv.InvoiceLines.InvoiceLine[0].Item.TertiarySellersItemIdentification = &schema.ItemIdentification{ID: "TERTIARY"}

		errs := validateItemIdentificationHierarchy(inv, ValidateOptions{Strict: true})

		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %v", errs)
		}
	})
}

// TestSchematronStoreBatches tests store batch validation.
// Based on Schematron rules: "Jednotky jednotlivých šarží" and "Součet množství za jednotlivé šarže"
func TestSchematronStoreBatches(t *testing.T) {
	t.Run("Store batch quantities sum matches InvoicedQuantity - OK", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].InvoicedQuantity = schema.Quantity{
			Value:    types.MustDecimal("10"),
			UnitCode: "KGM",
		}
		inv.InvoiceLines.InvoiceLine[0].Item.StoreBatches = &schema.StoreBatches{
			StoreBatch: []schema.StoreBatch{
				{
					Name:                "Batch-A",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("6"), UnitCode: "KGM"},
				},
				{
					Name:                "Batch-B",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("4"), UnitCode: "KGM"},
				},
			},
		}

		errs := validateStoreBatches(inv, ValidateOptions{Strict: true})

		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %v", errs)
		}
	})

	t.Run("Store batch quantities sum mismatch - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].InvoicedQuantity = schema.Quantity{
			Value:    types.MustDecimal("10"),
			UnitCode: "KGM",
		}
		inv.InvoiceLines.InvoiceLine[0].Item.StoreBatches = &schema.StoreBatches{
			StoreBatch: []schema.StoreBatch{
				{
					Name:                "Batch-A",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("6"), UnitCode: "KGM"},
				},
				{
					Name:                "Batch-B",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("3"), UnitCode: "KGM"}, // Sum = 9, should be 10
				},
			},
		}

		errs := validateStoreBatches(inv, ValidateOptions{Strict: true})

		if len(errs) == 0 {
			t.Error("Expected error for store batch quantity sum mismatch")
		}
		hasExpectedErr := false
		for _, err := range errs {
			if err.Code == ErrCodeTotalMismatch {
				hasExpectedErr = true
			}
		}
		if !hasExpectedErr {
			t.Error("Expected TOTAL_MISMATCH error")
		}
	})

	t.Run("Store batch unit code mismatch - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].InvoicedQuantity = schema.Quantity{
			Value:    types.MustDecimal("10"),
			UnitCode: "KGM",
		}
		inv.InvoiceLines.InvoiceLine[0].Item.StoreBatches = &schema.StoreBatches{
			StoreBatch: []schema.StoreBatch{
				{
					Name:                "Batch-A",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("6"), UnitCode: "KGM"},
				},
				{
					Name:                "Batch-B",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("4"), UnitCode: "PCS"}, // Wrong unit
				},
			},
		}

		errs := validateStoreBatches(inv, ValidateOptions{Strict: true})

		if len(errs) == 0 {
			t.Error("Expected error for store batch unit code mismatch")
		}
	})

	t.Run("Store batches with inconsistent unit codes - Error", func(t *testing.T) {
		inv := createValidInvoice()
		inv.InvoiceLines.InvoiceLine[0].InvoicedQuantity = schema.Quantity{
			Value: types.MustDecimal("10"),
			// No unit code on line
		}
		inv.InvoiceLines.InvoiceLine[0].Item.StoreBatches = &schema.StoreBatches{
			StoreBatch: []schema.StoreBatch{
				{
					Name:                "Batch-A",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("6"), UnitCode: "KGM"},
				},
				{
					Name:                "Batch-B",
					BatchOrSerialNumber: "B",
					Quantity:            schema.Quantity{Value: types.MustDecimal("4"), UnitCode: "PCS"}, // Different unit
				},
			},
		}

		errs := validateStoreBatches(inv, ValidateOptions{Strict: true})

		// Should error because batch unit codes are inconsistent with each other
		hasExpectedErr := false
		for _, err := range errs {
			if err.Field == "Invoice.InvoiceLines.InvoiceLine[0].Item.StoreBatches" {
				hasExpectedErr = true
			}
		}
		if !hasExpectedErr {
			t.Error("Expected error for inconsistent batch unit codes")
		}
	})
}

// TestSchematronIntegration tests full Schematron validation via validateSemantic.
func TestSchematronIntegration(t *testing.T) {
	t.Run("Credit note without original reference", func(t *testing.T) {
		inv := createValidInvoice()
		inv.DocumentType = 2 // Credit note
		inv.OriginalDocumentReferences = nil

		errs := validateSemantic(inv, DefaultValidateOptions())

		hasOriginalDocErr := false
		for _, err := range errs {
			if err.Field == "Invoice.OriginalDocumentReferences" {
				hasOriginalDocErr = true
			}
		}
		if !hasOriginalDocErr {
			t.Error("Expected OriginalDocumentReferences error")
		}
	})

	t.Run("Valid credit note", func(t *testing.T) {
		inv := createValidInvoice()
		inv.DocumentType = 2 // Credit note
		inv.OriginalDocumentReferences = &schema.OriginalDocumentReferences{
			OriginalDocumentReference: []schema.OriginalDocumentReference{
				{ID: "ORIG-001"},
			},
		}
		inv.CurrRate = types.MustDecimal("1")
		inv.RefCurrRate = types.MustDecimal("1")

		errs := validateSemantic(inv, DefaultValidateOptions())

		// Should pass all Schematron rules
		for _, err := range errs {
			if err.Severity == SeverityError {
				t.Errorf("Unexpected error: %s - %s", err.Field, err.Msg)
			}
		}
	})
}

// createValidInvoice creates a minimal valid invoice for testing.
func createValidInvoice() *schema.Invoice {
	return &schema.Invoice{
		Version:           "6.0.2",
		DocumentType:      1,
		ID:                "INV-001",
		UUID:              types.MustUUID("123e4567-e89b-12d3-a456-426614174000"),
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
					StreetName:     "Main Street",
					BuildingNumber: "123",
					CityName:       "Prague",
					PostalZone:     "10000",
					Country:        schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		AccountingCustomerParty: &schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "87654321"},
				PartyName:           schema.PartyName{Name: "Customer Ltd."},
				PostalAddress: schema.PostalAddress{
					StreetName:     "Second Street",
					BuildingNumber: "456",
					CityName:       "Brno",
					PostalZone:     "60200",
					Country:        schema.Country{IdentificationCode: "CZ"},
				},
			},
		},
		InvoiceLines: schema.InvoiceLines{
			InvoiceLine: []schema.InvoiceLine{
				{
					ID:                              "1",
					InvoicedQuantity:                schema.Quantity{Value: types.MustDecimal("1")},
					LineExtensionAmount:             types.MustDecimal("1000.00"),
					LineExtensionAmountTaxInclusive: types.MustDecimal("1210.00"),
					UnitPrice:                       types.MustDecimal("1000.00"),
					UnitPriceTaxInclusive:           types.MustDecimal("1210.00"),
					ClassifiedTaxCategory: schema.ClassifiedTaxCategory{
						Percent:       types.MustDecimal("21"),
						VATApplicable: types.Bool(true),
					},
					Item: schema.Item{
						Description: "Test Item",
					},
				},
			},
		},
		TaxTotal: schema.TaxTotal{
			TaxAmount: types.MustDecimal("210.00"),
			TaxSubTotal: []schema.TaxSubTotal{
				{
					TaxableAmount: types.MustDecimal("1000.00"),
					TaxAmount:     types.MustDecimal("210.00"),
					TaxCategory: schema.TaxCategory{
						Percent: types.MustDecimal("21"),
					},
				},
			},
		},
		LegalMonetaryTotal: schema.LegalMonetaryTotal{
			TaxExclusiveAmount: types.MustDecimal("1000.00"),
			TaxInclusiveAmount: types.MustDecimal("1210.00"),
			PayableAmount:      types.MustDecimal("1210.00"),
		},
	}
}
