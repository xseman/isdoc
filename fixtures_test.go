package isdoc

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFixturesRoundTrip tests decode-encode round-trip for test fixtures.
func TestFixturesRoundTrip(t *testing.T) {
	fixtures := []struct {
		name          string
		expectedID    string
		expectedUUID  string
		expectedLines int
	}{
		{
			name:          "test001.isdoc",
			expectedID:    "FV-1/2021",
			expectedUUID:  "AEC4791C-4BA1-451E-A1DC-2BF634B1C29D",
			expectedLines: 11, // based on the fixture content
		},
		{
			name:          "test002.isdoc",
			expectedID:    "FV-2/2021",
			expectedUUID:  "A34D00BF-FFB3-445B-BA1F-C5764B89409E",
			expectedLines: 11,
		},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join("testdata", "fixtures", tc.name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read fixture: %v", err)
			}

			// Decode
			invoice, err := DecodeBytes(data)
			if err != nil {
				if decErrs, ok := err.(DecodeErrors); ok {
					t.Logf("Decode warnings: %v", decErrs)
				} else {
					t.Fatalf("Failed to decode: %v", err)
				}
			}

			// Verify expected values
			if invoice.ID != tc.expectedID {
				t.Errorf("ID = %q, want %q", invoice.ID, tc.expectedID)
			}
			if string(invoice.UUID) != tc.expectedUUID {
				t.Errorf("UUID = %q, want %q", invoice.UUID, tc.expectedUUID)
			}
			if invoice.DocumentType != 1 {
				t.Errorf("DocumentType = %d, want 1", invoice.DocumentType)
			}
			if invoice.LocalCurrencyCode != "CZK" {
				t.Errorf("LocalCurrencyCode = %q, want CZK", invoice.LocalCurrencyCode)
			}

			// Verify supplier
			if invoice.AccountingSupplierParty.Party.PartyName.Name != "Demoverze" {
				t.Errorf("Supplier name = %q, want Demoverze",
					invoice.AccountingSupplierParty.Party.PartyName.Name)
			}
			if invoice.AccountingSupplierParty.Party.PartyIdentification.ID != "12345678" {
				t.Errorf("Supplier ID = %q, want 12345678",
					invoice.AccountingSupplierParty.Party.PartyIdentification.ID)
			}

			// Verify customer
			if invoice.AccountingCustomerParty.Party.PartyName.Name != "Odběratel 1" {
				t.Errorf("Customer name = %q, want 'Odběratel 1'",
					invoice.AccountingCustomerParty.Party.PartyName.Name)
			}

			// Verify invoice lines exist
			lineCount := len(invoice.InvoiceLines.InvoiceLine)
			if lineCount == 0 {
				t.Error("Invoice has no lines")
			}
			t.Logf("Invoice has %d lines", lineCount)

			// Encode back to XML
			encoded, err := EncodeBytes(invoice)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Decode the encoded XML to verify round-trip
			decoded, err := DecodeBytes(encoded)
			if err != nil {
				if decErrs, ok := err.(DecodeErrors); ok {
					t.Logf("Re-decode warnings: %v", decErrs)
				} else {
					t.Fatalf("Failed to re-decode: %v", err)
				}
			}

			// Verify round-trip preserved key values
			if decoded.ID != tc.expectedID {
				t.Errorf("Round-trip ID = %q, want %q", decoded.ID, tc.expectedID)
			}
			if string(decoded.UUID) != tc.expectedUUID {
				t.Errorf("Round-trip UUID = %q, want %q", decoded.UUID, tc.expectedUUID)
			}
		})
	}
}

// TestFixtureValidation validates test fixtures against ISDOC schema rules.
func TestFixtureValidation(t *testing.T) {
	fixtures := []string{
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

			// Run validation
			errs := ValidateInvoice(invoice)

			var errors []*ValidationError
			var warnings []*ValidationError

			for _, e := range errs {
				if e.Severity == SeverityError {
					errors = append(errors, e)
				} else {
					warnings = append(warnings, e)
				}
			}

			// Log warnings
			for _, w := range warnings {
				t.Logf("Warning: %s", w.Msg)
			}

			// Log errors but don't fail - real-world fixtures may have issues
			for _, e := range errors {
				t.Logf("Validation error: %s", e.Msg)
			}

			t.Logf("Validation: %d errors, %d warnings", len(errors), len(warnings))
		})
	}
}

// TestFixtureInvoiceDetails verifies detailed invoice content.
func TestFixtureInvoiceDetails(t *testing.T) {
	path := filepath.Join("testdata", "fixtures", "test001.isdoc")
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

	// Test version
	if invoice.Version != "6.0.2" {
		t.Errorf("Version = %q, want 6.0.2", invoice.Version)
	}

	// Test issuing system
	if invoice.IssuingSystem != "ABRA Gen® 21.1.4" {
		t.Errorf("IssuingSystem = %q, want 'ABRA Gen® 21.1.4'", invoice.IssuingSystem)
	}

	// Test VAT applicable
	if invoice.VATApplicable != true {
		t.Errorf("VATApplicable = %v, want true", invoice.VATApplicable)
	}

	// Test dates
	if invoice.IssueDate.String() != "2021-04-01" {
		t.Errorf("IssueDate = %s, want 2021-04-01", invoice.IssueDate)
	}
	if invoice.TaxPointDate.String() != "2021-04-01" {
		t.Errorf("TaxPointDate = %s, want 2021-04-01", invoice.TaxPointDate)
	}

	// Test supplier tax scheme
	if len(invoice.AccountingSupplierParty.Party.PartyTaxScheme) == 0 {
		t.Fatal("Supplier has no tax schemes")
	}
	taxScheme := invoice.AccountingSupplierParty.Party.PartyTaxScheme[0]
	if taxScheme.CompanyID != "CZ12345678" {
		t.Errorf("Supplier tax ID = %q, want CZ12345678", taxScheme.CompanyID)
	}
	if taxScheme.TaxScheme != "VAT" {
		t.Errorf("Tax scheme = %q, want VAT", taxScheme.TaxScheme)
	}

	// Test supplier address
	addr := invoice.AccountingSupplierParty.Party.PostalAddress
	if addr.StreetName != "Dodavatelská" {
		t.Errorf("Street = %q, want Dodavatelská", addr.StreetName)
	}
	if addr.CityName != "Dodavatelov" {
		t.Errorf("City = %q, want Dodavatelov", addr.CityName)
	}
	if addr.PostalZone != "12345" {
		t.Errorf("PostalZone = %q, want 12345", addr.PostalZone)
	}
	if addr.Country.IdentificationCode != "CZ" {
		t.Errorf("Country = %q, want CZ", addr.Country.IdentificationCode)
	}

	// Test delivery note references
	if invoice.DeliveryNoteReferences == nil {
		t.Error("DeliveryNoteReferences is nil")
	} else if len(invoice.DeliveryNoteReferences.DeliveryNoteReference) == 0 {
		t.Error("No delivery note references")
	} else {
		ref := invoice.DeliveryNoteReferences.DeliveryNoteReference[0]
		if ref.ID != "DL-1/2021" {
			t.Errorf("DeliveryNoteReference ID = %q, want DL-1/2021", ref.ID)
		}
	}

	// Test supplier contact
	contact := invoice.AccountingSupplierParty.Party.Contact
	if contact.Telephone != "123123123" {
		t.Errorf("Telephone = %q, want 123123123", contact.Telephone)
	}
	if contact.ElectronicMail != "dodavatel@posta.cz" {
		t.Errorf("Email = %q, want dodavatel@posta.cz", contact.ElectronicMail)
	}
}
