// Package isdoc provides encoding, decoding, and validation for ISDOC v6 documents.
//
// ISDOC is the Czech electronic invoice standard. This package provides:
//   - XML parsing with namespace handling
//   - Three-layer validation (structural, semantic, Schematron)
//   - Sequence-preserving XML encoding
//
// # Document Types
//
// ISDOC v6 supports two document types:
//   - Invoice: Tax documents (invoices, credit notes, debit notes)
//   - CommonDocument: Non-payment documents (contracts, certificates)
//
// # Basic Usage
//
// Parse and validate an invoice:
//
//	data, _ := os.ReadFile("invoice.isdoc")
//	invoice, err := isdoc.DecodeBytes(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Validate with default options
//	errs := isdoc.ValidateInvoice(invoice)
//	if errs.HasErrors() {
//	    log.Fatal(errs)
//	}
//
// Create and encode an invoice:
//
//	invoice := &schema.Invoice{
//	    Version:      "6.0.2",
//	    DocumentType: 1, // Invoice
//	    ID:           "FV-2025-001",
//	    UUID:         types.UUID("12345678-1234-1234-1234-123456789012"),
//	    IssueDate:    types.MustParseDate("2025-01-20"),
//	    // ... other required fields
//	}
//
//	xmlOut, err := isdoc.EncodeBytes(invoice)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Validation
//
// The package provides three-layer validation:
//
//  1. Structural validation: Checks XML structure and required fields
//  2. Semantic validation: Validates business logic and calculations
//  3. Schematron validation: Applies official ISDOC business rules
//
// Use ValidateInvoice for default validation, or ValidateInvoiceWithOptions
// for custom behavior:
//
//	opts := isdoc.ValidateOptions{
//	    Strict: true,
//	    AllowRoundingTolerance: false,
//	}
//	errs := isdoc.ValidateInvoiceWithOptions(invoice, opts)
//
// # CommonDocument
//
// For non-payment documents, use the CommonDocument API:
//
//	doc, err := isdoc.DecodeCommonDocumentBytes(data)
//	errs := isdoc.ValidateCommonDocument(doc)
//	xmlOut, err := isdoc.EncodeCommonDocumentBytes(doc)
//
// # Streaming API
//
// For large files, use the streaming API with io.Reader/io.Writer:
//
//	decoder := isdoc.NewDecoder(reader)
//	invoice, err := decoder.Decode()
//
//	encoder := isdoc.NewEncoder(writer)
//	err = encoder.Encode(invoice)
package isdoc

import (
	"io"

	"github.com/xseman/isdoc/schema"
)

// -----------------------------------------------------------------------------
// CommonDocument API
// -----------------------------------------------------------------------------

// DecodeCommonDocumentReader parses ISDOC CommonDocument XML from a reader.
func DecodeCommonDocumentReader(r io.Reader) (*schema.CommonDocument, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewDecodeError("", err)
	}
	return DecodeCommonDocumentBytes(data)
}

// EncodeCommonDocumentWriter encodes a CommonDocument to an io.Writer.
func EncodeCommonDocumentWriter(w io.Writer, doc *schema.CommonDocument) error {
	return NewEncoder(w).EncodeCommonDocument(doc)
}
