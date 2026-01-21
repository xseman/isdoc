package isdoc

import (
	"os"
	"testing"

	"github.com/xseman/isdoc/schema"
	"github.com/xseman/isdoc/types"
)

func TestParseCommonDocument(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/sample-commondocument.isdoc")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := DecodeCommonDocumentBytes(data)
	if err != nil {
		t.Fatalf("DecodeCommonDocumentBytes failed: %v", err)
	}

	// Verify parsed fields
	if doc.Version != "6.0.2" {
		t.Errorf("Version = %q, want %q", doc.Version, "6.0.2")
	}
	if doc.SubDocumentType != "CONTRACT" {
		t.Errorf("SubDocumentType = %q, want %q", doc.SubDocumentType, "CONTRACT")
	}
	if doc.SubDocumentTypeOrigin != "ACME Corp" {
		t.Errorf("SubDocumentTypeOrigin = %q, want %q", doc.SubDocumentTypeOrigin, "ACME Corp")
	}
	if doc.ID != "DOC-2025-001" {
		t.Errorf("ID = %q, want %q", doc.ID, "DOC-2025-001")
	}
	if doc.UUID.String() != "f47ac10b-58cc-4372-a567-0e02b2c3d479" {
		t.Errorf("UUID = %q, want %q", doc.UUID.String(), "f47ac10b-58cc-4372-a567-0e02b2c3d479")
	}
	if doc.Note == nil || doc.Note.Value != "Smlouva o poskytování služeb" {
		t.Errorf("Note.Value = %v, want %q", doc.Note, "Smlouva o poskytování služeb")
	}

	// Verify supplier
	if doc.AccountingSupplierParty.Party.PartyName.Name != "ACME Corporation s.r.o." {
		t.Errorf("Supplier name = %q, want %q",
			doc.AccountingSupplierParty.Party.PartyName.Name, "ACME Corporation s.r.o.")
	}

	// Verify customer
	if doc.AccountingCustomerParty.Party.PartyName.Name != "Widget Industries a.s." {
		t.Errorf("Customer name = %q, want %q",
			doc.AccountingCustomerParty.Party.PartyName.Name, "Widget Industries a.s.")
	}
}

func TestCommonDocumentRoundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/sample-commondocument.isdoc")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Parse
	doc, err := DecodeCommonDocumentBytes(data)
	if err != nil {
		t.Fatalf("DecodeCommonDocumentBytes failed: %v", err)
	}

	// Encode
	encoded, err := EncodeCommonDocumentBytes(doc)
	if err != nil {
		t.Fatalf("EncodeCommonDocumentBytes failed: %v", err)
	}

	// Parse again
	doc2, err := DecodeCommonDocumentBytes(encoded)
	if err != nil {
		t.Fatalf("second DecodeCommonDocumentBytes failed: %v", err)
	}

	// Verify key fields survived round-trip
	if doc2.ID != doc.ID {
		t.Errorf("ID changed: %q → %q", doc.ID, doc2.ID)
	}
	if doc2.SubDocumentType != doc.SubDocumentType {
		t.Errorf("SubDocumentType changed: %q → %q", doc.SubDocumentType, doc2.SubDocumentType)
	}
	if doc2.UUID.String() != doc.UUID.String() {
		t.Errorf("UUID changed: %q → %q", doc.UUID.String(), doc2.UUID.String())
	}
}

func TestValidateCommonDocument(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/sample-commondocument.isdoc")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := DecodeCommonDocumentBytes(data)
	if err != nil {
		t.Fatalf("DecodeCommonDocumentBytes failed: %v", err)
	}

	errs := ValidateCommonDocument(doc)
	if errs.HasErrors() {
		t.Errorf("valid CommonDocument should not have errors: %v", errs.Errors())
	}
}

func TestValidateCommonDocumentMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*schema.CommonDocument)
		wantErr string
	}{
		{
			name: "missing version",
			modify: func(d *schema.CommonDocument) {
				d.Version = ""
			},
			wantErr: "CommonDocument.@version",
		},
		{
			name: "missing SubDocumentType",
			modify: func(d *schema.CommonDocument) {
				d.SubDocumentType = ""
			},
			wantErr: "CommonDocument.SubDocumentType",
		},
		{
			name: "missing SubDocumentTypeOrigin",
			modify: func(d *schema.CommonDocument) {
				d.SubDocumentTypeOrigin = ""
			},
			wantErr: "CommonDocument.SubDocumentTypeOrigin",
		},
		{
			name: "missing ID",
			modify: func(d *schema.CommonDocument) {
				d.ID = ""
			},
			wantErr: "CommonDocument.ID",
		},
		{
			name: "missing UUID",
			modify: func(d *schema.CommonDocument) {
				d.UUID = ""
			},
			wantErr: "CommonDocument.UUID",
		},
		{
			name: "missing IssueDate",
			modify: func(d *schema.CommonDocument) {
				d.IssueDate = types.Date{}
			},
			wantErr: "CommonDocument.IssueDate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := validCommonDocument()
			tt.modify(doc)

			errs := ValidateCommonDocument(doc)
			if !errs.HasErrors() {
				t.Error("expected validation errors")
				return
			}

			found := false
			for _, e := range errs.Errors() {
				if e.Field == tt.wantErr {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error for field %q, got: %v", tt.wantErr, errs.Errors())
			}
		})
	}
}

func TestCommonDocumentValidation(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/sample-commondocument.isdoc")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	doc, err := DecodeCommonDocumentBytes(data)
	if err != nil {
		t.Fatalf("DecodeCommonDocumentBytes failed: %v", err)
	}

	// Validate
	errs := ValidateCommonDocument(doc)
	if errs.HasErrors() {
		t.Errorf("valid CommonDocument should be valid: %v", errs)
	}

	// Test encoding
	xml, err := EncodeCommonDocumentBytes(doc)
	if err != nil {
		t.Fatalf("EncodeCommonDocumentBytes failed: %v", err)
	}
	if len(xml) == 0 {
		t.Error("EncodeCommonDocumentBytes returned empty result")
	}
}

// validCommonDocument returns a minimal valid CommonDocument for testing.
func validCommonDocument() *schema.CommonDocument {
	return &schema.CommonDocument{
		Version:               "6.0.2",
		SubDocumentType:       "CONTRACT",
		SubDocumentTypeOrigin: "Test Corp",
		ID:                    "DOC-001",
		UUID:                  types.MustUUID("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
		IssueDate:             types.MustParseDate("2025-01-15"),
		AccountingSupplierParty: schema.AccountingSupplierParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "12345678"},
				PartyName:           schema.PartyName{Name: "Supplier Corp"},
				PostalAddress: schema.PostalAddress{
					StreetName:     "Main St",
					BuildingNumber: "1",
					CityName:       "Prague",
					PostalZone:     "11000",
					Country:        schema.Country{IdentificationCode: "CZ", Name: "Czech Republic"},
				},
			},
		},
		AccountingCustomerParty: schema.AccountingCustomerParty{
			Party: schema.Party{
				PartyIdentification: schema.PartyIdentification{ID: "87654321"},
				PartyName:           schema.PartyName{Name: "Customer Corp"},
				PostalAddress: schema.PostalAddress{
					StreetName:     "Side St",
					BuildingNumber: "2",
					CityName:       "Brno",
					PostalZone:     "60200",
					Country:        schema.Country{IdentificationCode: "CZ", Name: "Czech Republic"},
				},
			},
		},
	}
}
